"use client";

import Image from "next/image";
import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
  getGroupPage,
  getGroupMembers,
  getGroupPendingRequests,
  getGroupPendingInvites,
  acceptJoinRequest,
  rejectJoinRequest,
  removeInvite,
  kickMember,
  promoteMember,
  demoteMember,
  deleteGroup,
  inviteToGroup,
  requestToJoinGroup,
  getGroupSettings,
  updateGroupSettings,
  createGroupPost,
  getGroupPosts,
  deleteGroupPost,
  createGroupEvent,
  getGroupEventsTimeline,
  respondGroupEvent,
  changeGroupEventResponse,
  deleteGroupEvent,
} from "src/lib/services/group";
import { getFollowers, getFollowing } from "src/lib/services/follow";
import { fetchUserData } from "src/lib/services/user";
import { restorePost } from "src/lib/services/post";
import {
  getPostComments,
  createComment,
  deleteComment,
  updateComment,
  restoreComment,
} from "src/lib/services/comment";
import Avatar from "src/components/ui/Avatar";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import Ripple_Button from "src/components/ui/Ripple_Button";
import Echo_Button from "src/components/ui/Echo_Button";
import EmojiPickerButton from "src/components/ui/EmojiPickerButton";
import { useToast } from "src/components/ui/Toast";

const GroupDetailPage = () => {
  const params = useParams();
  const groupId = params.id;
  const toast = useToast();
  const [currentUser, setCurrentUser] = useState(null);

  const [activeTab, setActiveTab] = useState("posts");
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [showCreateEventModal, setShowCreateEventModal] = useState(false);
  const [showCreatePostModal, setShowCreatePostModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);

  // Loading states
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [userRole, setUserRole] = useState(null); // owner, moderator, member, or null

  // API data
  const [group, setGroup] = useState(null);
  const [members, setMembers] = useState([]);
  const [pendingRequests, setPendingRequests] = useState([]);
  const [pendingInvites, setPendingInvites] = useState([]);

  // More UI state
  const [posts, setPosts] = useState([]);
  const [postsLoading, setPostsLoading] = useState(false);
  const [postsPage, setPostsPage] = useState(1);
  const [hasMorePosts, setHasMorePosts] = useState(true);

  // Comments state (keyed by post id, same pattern as dashboard)
  const [commentsByPost, setCommentsByPost] = useState({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState({});
  const [commentInputByPost, setCommentInputByPost] = useState({});
  const [commentImageByPost, setCommentImageByPost] = useState({});
  const [commentSubmittingByPost, setCommentSubmittingByPost] = useState({});
  const [commentErrorByPost, setCommentErrorByPost] = useState({});
  const [editingCommentIdByPost, setEditingCommentIdByPost] = useState({});
  const [editingCommentContentByPost, setEditingCommentContentByPost] =
    useState({});
  const [commentActionLoadingById, setCommentActionLoadingById] = useState({});

  // Ripple state for group posts: { [postId]: { count, rippled } }
  const [rippleStateByPost, setRippleStateByPost] = useState({});

  const [expandedComments, setExpandedComments] = useState({});

  const [upcomingEvents, setUpcomingEvents] = useState([]);
  const [olderEvents, setOlderEvents] = useState([]);
  const [eventsLoading, setEventsLoading] = useState(false);
  const [eventActionLoading, setEventActionLoading] = useState({});

  // Load current user
  useEffect(() => {
    fetchUserData("me")
      .then((u) => setCurrentUser(u))
      .catch(() => {});
  }, []);

  function toUploadUrl(path) {
    if (!path) return "";
    if (
      path.startsWith("http://") ||
      path.startsWith("https://") ||
      path.startsWith("data:")
    )
      return path;
    if (path.startsWith("/uploads/")) return path;
    return "";
  }

  // Load group posts
  async function loadGroupPosts(page = 1, append = false) {
    setPostsLoading(true);
    try {
      const result = await getGroupPosts(groupId, page);
      const newPosts = result.posts || [];
      if (append) {
        setPosts((prev) => [...prev, ...newPosts]);
      } else {
        setPosts(newPosts);
      }
      setPostsPage(page);
      setHasMorePosts(newPosts.length >= 10);

      // Initialize ripple state
      const rippleInit = {};
      newPosts.forEach((post) => {
        rippleInit[post.id] = {
          count: post.likes_count || post.like_count || 0,
          rippled: !!post.has_current_user_liked,
        };
      });
      setRippleStateByPost((prev) =>
        append ? { ...prev, ...rippleInit } : rippleInit,
      );

      // Load comments for new posts
      const commentsEntries = await Promise.all(
        newPosts.map(async (post) => {
          try {
            const comments = await getPostComments(post.id);
            return [post.id, comments];
          } catch {
            return [post.id, []];
          }
        }),
      );
      const newComments = Object.fromEntries(commentsEntries);
      setCommentsByPost((prev) =>
        append ? { ...prev, ...newComments } : newComments,
      );
    } catch (err) {
      console.error("Failed to load group posts:", err);
    } finally {
      setPostsLoading(false);
    }
  }

  // Load group data on mount
  useEffect(() => {
    const loadGroupData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch group page (includes user's role)
        const groupData = await getGroupPage(groupId);
        if (groupData) {
          setGroup(groupData);
          setUserRole(groupData.role || null); // owner, moderator, member, or null
        }

        // If user is member, fetch members and moderator content
        if (groupData?.role) {
          const membersRaw = await getGroupMembers(groupId);
          const membersData = membersRaw.map((m) => ({
            ...m,
            role: m.role || m.group_status,
          }));
          setMembers(membersData);

          // Load group posts and events
          await loadGroupPosts(1);
          await loadGroupEvents();

          // If moderator/owner, fetch pending requests and invites
          if (groupData.role === "owner" || groupData.role === "moderator") {
            const requestsData = await getGroupPendingRequests(groupId);
            setPendingRequests(requestsData);

            const invitesData = await getGroupPendingInvites(groupId);
            setPendingInvites(invitesData);
          }
        }
      } catch (err) {
        console.error("Failed to load group data:", err);
        setError(err?.message || "Failed to load group data");
      } finally {
        setLoading(false);
      }
    };

    if (groupId) {
      loadGroupData();

      // Set up polling to refresh pending requests and invites every 5 seconds
      const interval = setInterval(async () => {
        try {
          if (userRole === "owner" || userRole === "moderator") {
            const requestsData = await getGroupPendingRequests(groupId);
            setPendingRequests(requestsData);

            const invitesData = await getGroupPendingInvites(groupId);
            setPendingInvites(invitesData);
          }
        } catch (err) {
          console.error("Failed to refresh pending data:", err);
        }
      }, 5000);

      return () => clearInterval(interval);
    }
  }, [groupId, userRole]);

  const formatMembers = (count) => {
    if (count >= 1000) {
      return (count / 1000).toFixed(count % 1000 === 0 ? 0 : 1) + "k";
    }
    return count.toString();
  };

  // Moderator/Owner handlers
  const handleAcceptRequest = async (userId) => {
    try {
      await acceptJoinRequest(groupId, userId);
      setPendingRequests(pendingRequests.filter((r) => r.id !== userId));
    } catch (error) {
      console.error("Failed to accept request:", error);
      toast.error(error?.message || "Failed to accept request");
    }
  };

  const handleDeclineRequest = async (userId) => {
    try {
      await rejectJoinRequest(groupId, userId);
      setPendingRequests(pendingRequests.filter((r) => r.id !== userId));
    } catch (error) {
      console.error("Failed to reject request:", error);
      toast.error(error?.message || "Failed to reject request");
    }
  };

  const handleRemoveInvite = async (userId) => {
    try {
      await removeInvite(groupId, userId);
      setPendingInvites(pendingInvites.filter((i) => i.id !== userId));
    } catch (error) {
      console.error("Failed to remove invite:", error);
      toast.error(error?.message || "Failed to remove invite");
    }
  };

  const handleKickMember = async (userId) => {
    toast.confirm("Are you sure you want to kick this member?", async () => {
      try {
        await kickMember(groupId, userId);
        setMembers(members.filter((m) => m.id !== userId));
      } catch (error) {
        console.error("Failed to kick member:", error);
        toast.error(error?.message || "Failed to kick member");
      }
    });
  };

  const handlePromoteMember = async (userId) => {
    try {
      await promoteMember(groupId, userId);
      setMembers(
        members.map((m) => (m.id === userId ? { ...m, role: "moderator" } : m)),
      );
    } catch (error) {
      console.error("Failed to promote member:", error);
      toast.error(error?.message || "Failed to promote member");
    }
  };

  const handleDemoteMember = async (userId) => {
    try {
      await demoteMember(groupId, userId);
      setMembers(
        members.map((m) => (m.id === userId ? { ...m, role: "member" } : m)),
      );
    } catch (error) {
      console.error("Failed to demote member:", error);
      toast.error(error?.message || "Failed to demote member");
    }
  };

  const handleDeleteGroupPost = async (postId) => {
    try {
      const deletedPost = posts.find((p) => p.id === postId);
      await deleteGroupPost(groupId, postId);
      setPosts((prev) => prev.filter((p) => p.id !== postId));
      toast.success("Post deleted", {
        duration: 5000,
        action: {
          label: "Undo",
          onClick: async () => {
            try {
              await restorePost(postId);
              if (deletedPost) setPosts((prev) => [deletedPost, ...prev]);
            } catch (e) {
              toast.error(e?.message || "Failed to restore post");
            }
          },
        },
      });
    } catch (error) {
      console.error("Failed to delete group post:", error);
      toast.error(error?.message || "Failed to delete post");
    }
  };

  // Comments
  async function loadComments(postId) {
    setCommentsLoadingByPost((prev) => ({ ...prev, [postId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      const comments = await getPostComments(postId);
      setCommentsByPost((prev) => ({ ...prev, [postId]: comments }));
    } catch (error) {
      console.error("Error loading comments:", error);
      setCommentsByPost((prev) => ({ ...prev, [postId]: [] }));
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to load echoes.",
      }));
    } finally {
      setCommentsLoadingByPost((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleCommentSubmit(event, postId) {
    event.preventDefault();
    const content = (commentInputByPost[postId] || "").trim();
    const image = commentImageByPost[postId] || null;
    if (!content) {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: "Comment content is required.",
      }));
      return;
    }
    setCommentSubmittingByPost((prev) => ({ ...prev, [postId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    const formData = new FormData();
    formData.append("content", content);
    formData.append("parent_type", "post");
    formData.append("parent_id", String(postId));
    if (image) formData.append("avatar", image);
    try {
      await createComment(formData);
      setCommentInputByPost((prev) => ({ ...prev, [postId]: "" }));
      setCommentImageByPost((prev) => ({ ...prev, [postId]: null }));
      await loadComments(postId);
    } catch (error) {
      console.error("Error creating comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to create echo.",
      }));
    } finally {
      setCommentSubmittingByPost((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleDeleteComment(postId, commentId) {
    if (!commentId) return;
    setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      await deleteComment(commentId);
      await loadComments(postId);
      toast.success("Comment deleted", {
        duration: 5000,
        action: {
          label: "Undo",
          onClick: async () => {
            try {
              await restoreComment(commentId);
              await loadComments(postId);
            } catch (e) {
              toast.error(e?.message || "Failed to restore comment");
            }
          },
        },
      });
    } catch (error) {
      console.error("Error deleting comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to delete echo.",
      }));
    } finally {
      setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: false }));
    }
  }

  async function handleSaveCommentEdit(postId, commentId) {
    const content = (editingCommentContentByPost[postId] || "").trim();
    if (!content) {
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: "Comment content is required.",
      }));
      return;
    }
    setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      await updateComment(commentId, content);
      setEditingCommentIdByPost((prev) => ({ ...prev, [postId]: null }));
      setEditingCommentContentByPost((prev) => ({ ...prev, [postId]: "" }));
      await loadComments(postId);
    } catch (error) {
      console.error("Error updating comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to update echo.",
      }));
    } finally {
      setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: false }));
    }
  }

  const toggleComments = (postId) => {
    setExpandedComments((prev) => ({
      ...prev,
      [postId]: !prev[postId],
    }));
  };

  async function loadGroupEvents() {
    setEventsLoading(true);
    try {
      const timeline = await getGroupEventsTimeline(groupId);
      setUpcomingEvents(timeline.upcoming_events);
      setOlderEvents(timeline.older_events);
    } catch (err) {
      console.error("Failed to load events:", err);
    } finally {
      setEventsLoading(false);
    }
  }

  const handleEventResponse = async (eventId, currentReaction, newReaction) => {
    setEventActionLoading((prev) => ({ ...prev, [eventId]: true }));
    try {
      if (currentReaction === "pending" || !currentReaction) {
        await respondGroupEvent(groupId, eventId, newReaction);
      } else {
        await changeGroupEventResponse(groupId, eventId, newReaction);
      }
      await loadGroupEvents();
    } catch (err) {
      console.error("Failed to respond to event:", err);
      toast.error(err?.message || "Failed to update response");
    } finally {
      setEventActionLoading((prev) => ({ ...prev, [eventId]: false }));
    }
  };

  const handleDeleteEvent = async (eventId) => {
    toast.confirm("Are you sure you want to cancel this event?", async () => {
      try {
        await deleteGroupEvent(groupId, eventId);
        setUpcomingEvents((prev) => prev.filter((e) => e.id !== eventId));
        setOlderEvents((prev) => prev.filter((e) => e.id !== eventId));
      } catch (err) {
        console.error("Failed to delete event:", err);
        toast.error(err?.message || "Failed to cancel event");
      }
    });
  };

  return (
    <main className="flex flex-col w-full max-w-4xl gap-6 p-4">
      {/* Back Button */}
      <Link
        href="/groups"
        className="flex items-center gap-2 text-purple-400 hover:text-purple-300 transition w-fit"
      >
        <span>←</span>
        <span>Back to Groups</span>
      </Link>

      {loading && (
        <div className="flex justify-center py-20">
          <p className="text-purple-400 text-lg">Loading group...</p>
        </div>
      )}

      {!loading && (error || !group) && (
        <div className="flex justify-center py-20">
          <div className="bg-[#1a1a2e] border border-purple-500/30 rounded-xl p-10 flex flex-col items-center gap-4 shadow-[0_0_15px_rgba(168,85,247,0.15)]">
            <span className="text-5xl">👥</span>
            <h2 className="text-2xl font-bold text-purple-200">
              Group Not Found
            </h2>
            <p className="text-purple-400 text-center max-w-sm">
              {error ||
                "The group you're looking for doesn't exist or may have been removed."}
            </p>
            <Link
              href="/groups"
              className="mt-2 px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white font-semibold rounded-lg transition shadow-custom"
            >
              Back to Groups
            </Link>
          </div>
        </div>
      )}

      {!loading && group && (
        <>
          {/* Group Header */}
          <header className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 overflow-hidden">
            {/* Cover Image */}
            <div className="h-32 bg-linear-to-r from-purple-900 via-purple-600 to-purple-900 relative">
              <div className="absolute inset-0 bg-[url('/grid-pattern.svg')] opacity-20"></div>
            </div>

            {/* Group Info */}
            <div className="p-6 -mt-8 relative">
              <div className="flex items-end gap-4">
                {/* Group Avatar */}
                <div className="w-20 h-20 rounded-lg bg-purple-600 flex items-center justify-center text-white text-3xl font-bold shadow-[0_0_20px_rgba(168,85,247,0.4)] border-4 border-[#1a1a2e]">
                  {group?.name?.[0]}
                </div>

                <div className="flex-1 pb-2">
                  <div className="flex items-center gap-3">
                    <h1 className="text-2xl font-bold text-purple-100">
                      {group.name}
                    </h1>
                    {group.visibility === "private" ? (
                      <span className="text-base">🔒</span>
                    ) : (
                      <span className="text-base">🌐</span>
                    )}
                    {userRole === "owner" && (
                      <span className="bg-green-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(34,197,94,0.4)]">
                        Owner
                      </span>
                    )}
                    {userRole === "moderator" && (
                      <span className="bg-blue-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(59,130,246,0.4)]">
                        Moderator
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 text-sm text-purple-400 mt-1">
                    <span className="flex items-center gap-1">
                      <span className="opacity-60">👥</span>
                      {formatMembers(group.group_members || group.members)}{" "}
                      members
                    </span>
                    <span>Created {group.created_at || group.createdAt}</span>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex gap-2">
                  {(userRole === "owner" || userRole === "moderator") && (
                    <button
                      onClick={() => setShowInviteModal(true)}
                      className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition cursor-pointer text-sm flex items-center gap-2"
                    >
                      <span>+</span> Invite
                    </button>
                  )}
                  {userRole === "owner" && (
                    <button
                      onClick={() => setShowSettingsModal(true)}
                      className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                    >
                      <span className="text-sm">⚙️</span>
                    </button>
                  )}
                  {!userRole && group.can_request && (
                    <button
                      onClick={async () => {
                        try {
                          await requestToJoinGroup(groupId);
                          if (group.join_mode === "auto") {
                            // Auto-join: reload the page data
                            const groupData = await getGroupPage(groupId);
                            if (groupData) {
                              setGroup(groupData);
                              setUserRole(groupData.role || null);
                            }
                            if (groupData?.role) {
                              const membersData =
                                await getGroupMembers(groupId);
                              setMembers(membersData);
                              await loadGroupPosts(1);
                              await loadGroupEvents();
                            }
                          } else {
                            setGroup((prev) => ({
                              ...prev,
                              pending_type: "requested",
                              can_request: false,
                            }));
                          }
                        } catch (err) {
                          console.error("Failed to join group:", err);
                          toast.error(err?.message || "Failed to join group");
                        }
                      }}
                      className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] hover:shadow-[0_0_15px_rgba(168,85,247,0.5)]"
                    >
                      {group.join_mode === "auto"
                        ? "Join Group"
                        : "Request to Join"}
                    </button>
                  )}
                  {!userRole && group.pending_type === "requested" && (
                    <span className="px-4 py-2 bg-purple-900/30 text-purple-400 border border-purple-500/30 rounded-md text-sm cursor-default">
                      Request Pending
                    </span>
                  )}
                  {!userRole && group.pending_type === "invited" && (
                    <span className="px-4 py-2 bg-blue-900/30 text-blue-300 border border-blue-500/30 rounded-md text-sm cursor-default">
                      Invited
                    </span>
                  )}
                </div>
              </div>

              <p className="text-purple-300/80 mt-4">{group.description}</p>
            </div>
          </header>

          {/* Tabs - only show for members */}
          {userRole && (
            <div className="flex flex-row gap-3 w-full flex-wrap">
              <button
                onClick={() => setActiveTab("posts")}
                className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                  activeTab === "posts"
                    ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                    : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                }`}
              >
                Posts
              </button>
              <button
                onClick={() => setActiveTab("members")}
                className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                  activeTab === "members"
                    ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                    : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                }`}
              >
                Members ({members.length})
              </button>
              <button
                onClick={() => setActiveTab("events")}
                className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                  activeTab === "events"
                    ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                    : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                }`}
              >
                Events ({upcomingEvents.length + olderEvents.length})
              </button>
              {(userRole === "owner" || userRole === "moderator") && (
                <>
                  <button
                    onClick={() => setActiveTab("requests")}
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md relative ${
                      activeTab === "requests"
                        ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                        : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                  >
                    Join Requests
                    {pendingRequests.length > 0 && (
                      <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center shadow-[0_0_8px_rgba(239,68,68,0.6)]">
                        {pendingRequests.length}
                      </span>
                    )}
                  </button>
                  <button
                    onClick={() => setActiveTab("invites")}
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md relative ${
                      activeTab === "invites"
                        ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                        : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                  >
                    Pending Invites
                    {pendingInvites.length > 0 && (
                      <span className="absolute -top-1 -right-1 bg-orange-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center shadow-[0_0_8px_rgba(249,115,22,0.6)]">
                        {pendingInvites.length}
                      </span>
                    )}
                  </button>
                </>
              )}
            </div>
          )}

          {/* Non-member info */}
          {!userRole && !loading && group && (
            <div className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-8 text-center">
              <span className="text-5xl mx-auto mb-4 opacity-50 block text-center">
                👥
              </span>
              <h3 className="text-lg font-semibold text-purple-200 mb-2">
                {group.pending_type === "requested"
                  ? "Your request is pending"
                  : "You are not a member"}
              </h3>
              <p className="text-purple-400 text-sm">
                {group.pending_type === "requested"
                  ? "A moderator will review your request soon."
                  : group.can_request
                    ? `${group.join_mode === "auto" ? "Join this group" : "Request to join this group"} to see posts, events, and members.`
                    : "This group is not accepting new members right now."}
              </p>
            </div>
          )}

          {/* Posts Section */}
          {userRole && activeTab === "posts" && (
            <section className="flex flex-col gap-4">
              {/* Create Post */}
              <div className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4">
                <button
                  onClick={() => setShowCreatePostModal(true)}
                  className="w-full text-left px-4 py-3 bg-[#0d0d1a] border border-purple-500/30 rounded-md text-purple-400/50 hover:border-purple-500/50 transition cursor-pointer"
                >
                  Write something to the group...
                </button>
              </div>

              {/* Posts List */}
              {postsLoading && posts.length === 0 ? (
                <p className="text-center text-purple-300">Loading posts...</p>
              ) : posts.length === 0 ? (
                <EmptyState
                  message="No posts yet"
                  subMessage="Be the first to post in this group!"
                />
              ) : (
                posts.map((post) => {
                  const echoSectionId = `echo-section-${post.id}`;
                  const echoPhotoUploadId = `echo-photo-upload-${post.id}`;
                  const comments = commentsByPost[post.id] || [];
                  const isCommentsLoading = commentsLoadingByPost[post.id];
                  const commentValue = commentInputByPost[post.id] || "";
                  const isCommentSubmitting = commentSubmittingByPost[post.id];
                  const commentError = commentErrorByPost[post.id] || "";
                  const isOwnPost = post.user_id === currentUser?.id;
                  const canDelete =
                    isOwnPost ||
                    userRole === "owner" ||
                    userRole === "moderator";
                  const postDateLabel = formatFriendlyDateTime(
                    post.created_at_time || post.created_at,
                  );
                  const rippleCount =
                    rippleStateByPost[post.id]?.count ??
                    post.likes_count ??
                    post.like_count ??
                    0;
                  const rippled =
                    rippleStateByPost[post.id]?.rippled ??
                    !!post.has_current_user_liked;
                  const handleRippleChange = (newCount, newRippled) => {
                    setRippleStateByPost((prev) => ({
                      ...prev,
                      [post.id]: { count: newCount, rippled: newRippled },
                    }));
                  };

                  return (
                    <article
                      key={post.id}
                      className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all"
                    >
                      <div className="flex items-start gap-3">
                        <Avatar
                          src={post.author_profile_picture}
                          name={`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim()}
                          size={40}
                          className="shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                        />
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              {post.user_id ? (
                                <Link
                                  href={`/profile/${post.user_id}`}
                                  className="font-semibold text-purple-100 hover:underline"
                                >
                                  {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() ||
                                    "Unknown User"}
                                </Link>
                              ) : (
                                <span className="font-semibold text-purple-100">
                                  {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() ||
                                    "Unknown User"}
                                </span>
                              )}
                              <span className="text-purple-400/60 text-sm">
                                {postDateLabel}
                              </span>
                            </div>
                            {canDelete && (
                              <button
                                onClick={() => handleDeleteGroupPost(post.id)}
                                className="text-xs bg-purple-900/30 hover:bg-red-900/30 text-purple-300 hover:text-red-300 border border-purple-500/30 hover:border-red-500/30 rounded-md px-3 py-1 transition cursor-pointer"
                              >
                                Delete
                              </button>
                            )}
                          </div>
                          <p className="text-purple-300/80 mt-2">
                            {post.content}
                          </p>
                          {post.extra_content || post.image ? (
                            <div className="mt-3">
                              <Image
                                src={toUploadUrl(
                                  post.extra_content || post.image,
                                )}
                                alt="Post image"
                                width={500}
                                height={300}
                                className="rounded-lg w-full h-auto"
                              />
                            </div>
                          ) : null}
                          <div className="flex items-center gap-6 mt-4 text-sm text-purple-400">
                            <span>{rippleCount} Ripples</span>
                            <span>{comments.length} Echoes</span>
                          </div>
                          <div className="flex items-center gap-6 mt-2 pt-2 border-t border-purple-500/20">
                            <Ripple_Button
                              postId={post.id}
                              initialRippled={rippled}
                              initialCount={rippleCount}
                              onChange={handleRippleChange}
                            />
                            <Echo_Button
                              targetId={echoSectionId}
                              onToggle={(isOpen) => {
                                if (
                                  isOpen &&
                                  commentsByPost[post.id] === undefined
                                ) {
                                  loadComments(post.id);
                                }
                              }}
                            />
                          </div>

                          {/* Comments Section */}
                          <div
                            id={echoSectionId}
                            className="mt-4 pt-4 border-t border-purple-500/20 hidden flex-col gap-2"
                          >
                            {/* Comment Input */}
                            <form
                              onSubmit={(e) => handleCommentSubmit(e, post.id)}
                              className="flex gap-2 mb-2"
                            >
                              <Avatar
                                src={currentUser?.profile_picture}
                                name={`${currentUser?.first_name || ""} ${currentUser?.last_name || ""}`.trim()}
                                size={32}
                                className="shadow-[0_0_8px_rgba(168,85,247,0.3)]"
                              />
                              <div className="flex-1 flex gap-2">
                                <div className="flex-1 flex items-center bg-[#0d0d1a] border border-purple-500/30 rounded-md">
                                  <input
                                    type="text"
                                    value={commentValue}
                                    onChange={(e) =>
                                      setCommentInputByPost((prev) => ({
                                        ...prev,
                                        [post.id]: e.target.value,
                                      }))
                                    }
                                    placeholder="Write a comment..."
                                    className="flex-1 px-3 py-2 bg-transparent focus:outline-none text-purple-100 placeholder-purple-400/50 text-sm"
                                    disabled={isCommentSubmitting}
                                  />
                                  <EmojiPickerButton
                                    onEmojiSelect={(emoji) =>
                                      setCommentInputByPost((prev) => ({
                                        ...prev,
                                        [post.id]:
                                          (prev[post.id] || "") + emoji,
                                      }))
                                    }
                                  />
                                  <label
                                    htmlFor={echoPhotoUploadId}
                                    className="flex items-center px-2 cursor-pointer"
                                  >
                                    <span className="opacity-60 text-base">
                                      📷
                                    </span>
                                    <input
                                      id={echoPhotoUploadId}
                                      type="file"
                                      className="hidden"
                                      accept="image/*"
                                      onChange={(e) =>
                                        setCommentImageByPost((prev) => ({
                                          ...prev,
                                          [post.id]:
                                            e.target.files?.[0] || null,
                                        }))
                                      }
                                      disabled={isCommentSubmitting}
                                    />
                                  </label>
                                </div>
                                <button
                                  type="submit"
                                  disabled={isCommentSubmitting}
                                  className="px-3 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] disabled:opacity-50"
                                >
                                  {isCommentSubmitting ? "..." : "Echo"}
                                </button>
                              </div>
                            </form>

                            {commentError && (
                              <p className="text-red-400 text-sm">
                                {commentError}
                              </p>
                            )}

                            {/* Comments List */}
                            {isCommentsLoading ? (
                              <p className="text-purple-400/50 text-sm text-center py-2">
                                Loading echoes...
                              </p>
                            ) : comments.length > 0 ? (
                              <div className="space-y-3">
                                {comments.map((comment) => (
                                  <div key={comment.id} className="flex gap-2">
                                    <Avatar
                                      src={comment.author_profile_picture}
                                      name={`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim()}
                                      size={32}
                                    />
                                    <div className="flex-1 bg-[#0d0d1a] rounded-md p-3 border border-purple-500/20">
                                      <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-2">
                                          {comment.user_id ? (
                                            <Link
                                              href={`/profile/${comment.user_id}`}
                                              className="font-semibold text-purple-100 text-sm hover:underline"
                                            >
                                              {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() ||
                                                "Unknown User"}
                                            </Link>
                                          ) : (
                                            <span className="font-semibold text-purple-100 text-sm">
                                              {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() ||
                                                "Unknown User"}
                                            </span>
                                          )}
                                          <span className="text-purple-400/50 text-xs">
                                            {formatFriendlyDateTime(
                                              comment.created_at_time ||
                                                comment.created_at,
                                            )}
                                          </span>
                                        </div>
                                        {comment.user_id ===
                                          currentUser?.id && (
                                          <div className="flex gap-2">
                                            <button
                                              type="button"
                                              className="text-xs text-purple-400 hover:text-purple-200 transition"
                                              onClick={() => {
                                                setEditingCommentIdByPost(
                                                  (prev) => ({
                                                    ...prev,
                                                    [post.id]: comment.id,
                                                  }),
                                                );
                                                setEditingCommentContentByPost(
                                                  (prev) => ({
                                                    ...prev,
                                                    [post.id]:
                                                      comment.content || "",
                                                  }),
                                                );
                                              }}
                                              disabled={
                                                !!commentActionLoadingById[
                                                  comment.id
                                                ]
                                              }
                                            >
                                              Edit
                                            </button>
                                            <button
                                              type="button"
                                              className="text-xs text-purple-400 hover:text-red-300 transition"
                                              onClick={() =>
                                                handleDeleteComment(
                                                  post.id,
                                                  comment.id,
                                                )
                                              }
                                              disabled={
                                                !!commentActionLoadingById[
                                                  comment.id
                                                ]
                                              }
                                            >
                                              Delete
                                            </button>
                                          </div>
                                        )}
                                      </div>
                                      {editingCommentIdByPost[post.id] ===
                                      comment.id ? (
                                        <div className="flex items-center gap-2 mt-1">
                                          <input
                                            type="text"
                                            className="flex-1 px-2 py-1 bg-[#1a1a2e] border border-purple-500/30 rounded text-purple-100 text-sm focus:outline-none focus:ring-1 focus:ring-purple-500"
                                            value={
                                              editingCommentContentByPost[
                                                post.id
                                              ] || ""
                                            }
                                            onChange={(e) =>
                                              setEditingCommentContentByPost(
                                                (prev) => ({
                                                  ...prev,
                                                  [post.id]: e.target.value,
                                                }),
                                              )
                                            }
                                          />
                                          <button
                                            type="button"
                                            className="text-xs px-2 py-1 rounded bg-purple-600 text-white disabled:opacity-50"
                                            onClick={() =>
                                              handleSaveCommentEdit(
                                                post.id,
                                                comment.id,
                                              )
                                            }
                                            disabled={
                                              !!commentActionLoadingById[
                                                comment.id
                                              ]
                                            }
                                          >
                                            Save
                                          </button>
                                          <button
                                            type="button"
                                            className="text-xs px-2 py-1 rounded bg-purple-900/30 text-purple-300"
                                            onClick={() => {
                                              setEditingCommentIdByPost(
                                                (prev) => ({
                                                  ...prev,
                                                  [post.id]: null,
                                                }),
                                              );
                                              setEditingCommentContentByPost(
                                                (prev) => ({
                                                  ...prev,
                                                  [post.id]: "",
                                                }),
                                              );
                                            }}
                                          >
                                            Cancel
                                          </button>
                                        </div>
                                      ) : (
                                        <p className="text-purple-300/80 text-sm mt-1">
                                          {comment.content}
                                        </p>
                                      )}
                                      {comment.image ? (
                                        <div className="mt-2">
                                          <Image
                                            src={toUploadUrl(comment.image)}
                                            alt="Comment image"
                                            width={300}
                                            height={180}
                                            className="rounded w-full h-auto"
                                          />
                                        </div>
                                      ) : null}
                                    </div>
                                  </div>
                                ))}
                              </div>
                            ) : (
                              <p className="text-purple-400/50 text-sm text-center py-2">
                                No echoes yet. Be the first to comment!
                              </p>
                            )}
                          </div>
                        </div>
                      </div>
                    </article>
                  );
                })
              )}

              {/* Load More */}
              {hasMorePosts && posts.length > 0 && (
                <button
                  onClick={() => loadGroupPosts(postsPage + 1, true)}
                  disabled={postsLoading}
                  className="w-full py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50"
                >
                  {postsLoading ? "Loading..." : "Load More Posts"}
                </button>
              )}
            </section>
          )}

          {/* Members Section */}
          {userRole && activeTab === "members" && (
            <section className="flex flex-col gap-4">
              {members.map((member) => (
                <article
                  key={member.id}
                  className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                      {(member.first_name || "U")[0]}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-purple-100">
                          {member.first_name} {member.last_name}
                        </span>
                        {member.role === "creator" && (
                          <span className="bg-green-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(34,197,94,0.4)]">
                            Creator
                          </span>
                        )}
                        {member.role === "moderator" && (
                          <span className="bg-purple-600 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(168,85,247,0.4)]">
                            Moderator
                          </span>
                        )}
                      </div>
                    </div>
                    {userRole === "owner" && member.id !== currentUser?.id && (
                      <div className="flex gap-2">
                        {member.role === "member" && (
                          <button
                            onClick={() => handlePromoteMember(member.id)}
                            className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                          >
                            Promote
                          </button>
                        )}
                        {member.role === "moderator" && (
                          <button
                            onClick={() => handleDemoteMember(member.id)}
                            className="px-3 py-1.5 bg-yellow-600 hover:bg-yellow-500 text-white rounded-md transition cursor-pointer text-sm"
                          >
                            Demote
                          </button>
                        )}
                        <button
                          onClick={() => handleKickMember(member.id)}
                          className="px-3 py-1.5 bg-purple-900/30 hover:bg-red-900/30 text-purple-300 hover:text-red-300 border border-purple-500/30 hover:border-red-500/30 rounded-md transition cursor-pointer text-sm"
                        >
                          Remove
                        </button>
                      </div>
                    )}
                  </div>
                </article>
              ))}
            </section>
          )}

          {/* Events Section */}
          {userRole && activeTab === "events" && (
            <section className="flex flex-col gap-4">
              {/* Create Event - any active member can create */}
              <button
                onClick={() => setShowCreateEventModal(true)}
                className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition cursor-pointer text-sm flex items-center gap-2 w-fit"
              >
                <span>+</span> Create Event
              </button>

              {eventsLoading ? (
                <p className="text-center text-purple-300">Loading events...</p>
              ) : upcomingEvents.length === 0 && olderEvents.length === 0 ? (
                <EmptyState
                  message="No events yet"
                  subMessage="Create an event to get started!"
                />
              ) : (
                <>
                  {/* Upcoming Events */}
                  {upcomingEvents.length > 0 && (
                    <>
                      <h3 className="text-purple-200 font-semibold text-sm uppercase tracking-wider">
                        Upcoming
                      </h3>
                      {upcomingEvents.map((event) => (
                        <EventCard
                          key={event.id}
                          event={event}
                          userRole={userRole}
                          currentUserId={currentUser?.id}
                          loading={!!eventActionLoading[event.id]}
                          onRespond={(reaction) =>
                            handleEventResponse(
                              event.id,
                              event.my_reaction,
                              reaction,
                            )
                          }
                          onDelete={() => handleDeleteEvent(event.id)}
                        />
                      ))}
                    </>
                  )}

                  {/* Older Events */}
                  {olderEvents.length > 0 && (
                    <>
                      <h3 className="text-purple-400/60 font-semibold text-sm uppercase tracking-wider mt-4">
                        Past Events
                      </h3>
                      {olderEvents.map((event) => (
                        <EventCard
                          key={event.id}
                          event={event}
                          userRole={userRole}
                          currentUserId={currentUser?.id}
                          loading={!!eventActionLoading[event.id]}
                          onRespond={(reaction) =>
                            handleEventResponse(
                              event.id,
                              event.my_reaction,
                              reaction,
                            )
                          }
                          onDelete={() => handleDeleteEvent(event.id)}
                          isPast
                        />
                      ))}
                    </>
                  )}
                </>
              )}
            </section>
          )}

          {/* Join Requests Section */}
          {activeTab === "requests" &&
            (userRole === "owner" || userRole === "moderator") && (
              <section className="flex flex-col gap-4">
                {pendingRequests.length > 0 ? (
                  pendingRequests.map((request) => (
                    <article
                      key={request.id}
                      className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all"
                    >
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                          {(request.first_name || "U")[0]}
                          {(request.last_name || "")?.[0] || ""}
                        </div>
                        <div className="flex-1">
                          <span className="font-semibold text-purple-100">
                            {request.first_name} {request.last_name}
                          </span>
                          <p className="text-purple-400/60 text-sm">
                            Requested to join
                          </p>
                        </div>
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleAcceptRequest(request.id)}
                            className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                          >
                            Accept
                          </button>
                          <button
                            onClick={() => handleDeclineRequest(request.id)}
                            className="px-3 py-1.5 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                          >
                            Decline
                          </button>
                        </div>
                      </div>
                    </article>
                  ))
                ) : (
                  <EmptyState
                    message="No pending requests"
                    subMessage="All join requests have been handled"
                  />
                )}
              </section>
            )}

          {/* Pending Invites Section */}
          {activeTab === "invites" &&
            (userRole === "owner" || userRole === "moderator") && (
              <section className="flex flex-col gap-4">
                {pendingInvites.length > 0 ? (
                  pendingInvites.map((invite) => (
                    <article
                      key={invite.id}
                      className="bg-[#1a1a2e] rounded-lg border border-orange-500/30 p-4 hover:border-orange-500/50 hover:shadow-[0_0_15px_rgba(249,115,22,0.15)] transition-all"
                    >
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-orange-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(249,115,22,0.3)]">
                          {(invite.first_name || "U")[0]}
                          {(invite.last_name || "")?.[0] || ""}
                        </div>
                        <div className="flex-1">
                          <span className="font-semibold text-purple-100">
                            {invite.first_name} {invite.last_name}
                          </span>
                          <p className="text-purple-400/60 text-sm">
                            Pending invite
                          </p>
                        </div>
                        <div className="flex gap-2">
                          <button
                            onClick={() => handleRemoveInvite(invite.id)}
                            className="px-3 py-1.5 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                          >
                            Cancel
                          </button>
                        </div>
                      </div>
                    </article>
                  ))
                ) : (
                  <EmptyState
                    message="No pending invites"
                    subMessage="All invites have been sent or accepted"
                  />
                )}
              </section>
            )}

          {/* Invite Modal */}
          {showInviteModal && (
            <InviteModal
              onClose={() => setShowInviteModal(false)}
              members={members}
              groupId={groupId}
            />
          )}

          {/* Create Post Modal */}
          {showCreatePostModal && (
            <CreatePostModal
              groupId={groupId}
              onClose={() => setShowCreatePostModal(false)}
              onCreated={(newPost) => {
                setPosts((prev) => [newPost, ...prev]);
                // Initialize ripple state for the new post
                setRippleStateByPost((prev) => ({
                  ...prev,
                  [newPost.id]: { count: 0, rippled: false },
                }));
                setCommentsByPost((prev) => ({ ...prev, [newPost.id]: [] }));
                setShowCreatePostModal(false);
              }}
            />
          )}

          {/* Create Event Modal */}
          {showCreateEventModal && (
            <CreateEventModal
              groupId={groupId}
              onClose={() => setShowCreateEventModal(false)}
              onCreated={() => {
                loadGroupEvents();
                setShowCreateEventModal(false);
              }}
            />
          )}

          {/* Settings Modal */}
          {showSettingsModal && (
            <GroupSettingsModal
              groupId={groupId}
              currentSettings={{
                visibility: group?.visibility,
                join_mode: group?.join_mode,
              }}
              onClose={() => setShowSettingsModal(false)}
              onSaved={(updated) => {
                setGroup((prev) => ({
                  ...prev,
                  visibility: updated.Visibility || updated.visibility,
                  join_mode: updated.JoinMode || updated.join_mode,
                }));
                setShowSettingsModal(false);
              }}
            />
          )}
        </>
      )}
    </main>
  );
};

// Empty State Component
const EmptyState = ({ message, subMessage }) => (
  <div className="text-center py-12 bg-[#1a1a2e] rounded-lg border border-purple-500/30 shadow-[0_0_20px_rgba(168,85,247,0.1)]">
    <span className="text-5xl mx-auto mb-4 opacity-50 block text-center">
      👥
    </span>
    <h3 className="text-lg font-semibold text-purple-200 mb-2">{message}</h3>
    <p className="text-purple-400 text-sm">{subMessage}</p>
  </div>
);

// Invite Modal Component
const InviteModal = ({ onClose, members = [], groupId }) => {
  const toast = useToast();
  const [searchQuery, setSearchQuery] = useState("");
  const [invitedUsers, setInvitedUsers] = useState({});
  const [availableUsers, setAvailableUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [sendingInvites, setSendingInvites] = useState(false);

  // Fetch followers + following on mount
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        setLoading(true);
        setError(null);
        const [followers, following] = await Promise.all([
          getFollowers().catch(() => []),
          getFollowing().catch(() => []),
        ]);

        // Deduplicate by id
        const seen = new Set();
        const combined = [...followers, ...following].filter((u) => {
          if (!u?.id || seen.has(u.id)) return false;
          seen.add(u.id);
          return true;
        });

        // Filter out members already in the group
        const memberIds = new Set(members.map((m) => m.id));
        const filtered = combined.filter((user) => !memberIds.has(user.id));

        setAvailableUsers(filtered);
      } catch (err) {
        console.error("Failed to load users:", err);
        setError(err?.message || "Failed to load users");
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, [members]);

  // Filter users based on search query
  const filteredUsers = availableUsers.filter((user) => {
    const query = searchQuery.toLowerCase();
    return (
      user.first_name.toLowerCase().includes(query) ||
      user.last_name.toLowerCase().includes(query)
    );
  });

  const handleInviteUser = (userId) => {
    setInvitedUsers((prev) => ({
      ...prev,
      [userId]: !prev[userId],
    }));
  };

  const handleSendInvites = async () => {
    const selectedUserIds = Object.keys(invitedUsers).filter(
      (id) => invitedUsers[id],
    );
    if (selectedUserIds.length === 0) {
      toast.warning("Please select at least one user to invite");
      return;
    }

    setSendingInvites(true);
    try {
      // Send invites to each selected user
      await Promise.all(
        selectedUserIds.map((userId) => inviteToGroup(groupId, userId)),
      );
      toast.success(`Invitations sent to ${selectedUserIds.length} user(s)!`);
      onClose();
    } catch (err) {
      console.error("Failed to send invites:", err);
      toast.error(err?.message || "Failed to send invites");
    } finally {
      setSendingInvites(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
        <h2 className="text-xl font-bold mb-4 text-purple-100">
          Invite Members
        </h2>

        <div className="relative mb-4">
          <span className="absolute left-3 top-1/2 transform -translate-y-1/2 opacity-60 text-sm">
            🔍
          </span>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search users to invite..."
            className="w-full pl-10 pr-4 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:shadow-[0_0_10px_rgba(168,85,247,0.3)] text-purple-100 placeholder-purple-400/50 text-sm"
          />
        </div>

        {/* Users List */}
        <div className="max-h-80 overflow-y-auto mb-4 space-y-2">
          {loading ? (
            <p className="text-purple-400/60 text-sm text-center py-8">
              Loading users...
            </p>
          ) : error ? (
            <p className="text-red-400 text-sm text-center py-8">{error}</p>
          ) : filteredUsers.length > 0 ? (
            filteredUsers.map((user) => (
              <div
                key={user.id}
                className="flex items-center gap-3 p-3 bg-[#0d0d1a] border border-purple-500/20 rounded-md hover:border-purple-500/40 transition"
              >
                <div className="w-10 h-10 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shrink-0 shadow-[0_0_8px_rgba(168,85,247,0.3)]">
                  {(user.first_name || "U")[0]}
                  {(user.last_name || "")[0]}
                </div>
                <div className="flex-1">
                  <p className="text-purple-100 font-medium text-sm">
                    {user.first_name} {user.last_name}
                  </p>
                </div>
                <input
                  type="checkbox"
                  checked={invitedUsers[user.id] || false}
                  onChange={() => handleInviteUser(user.id)}
                  disabled={sendingInvites}
                  className="w-5 h-5 accent-purple-500 cursor-pointer disabled:opacity-50"
                />
              </div>
            ))
          ) : (
            <p className="text-purple-400/60 text-sm text-center py-8">
              {searchQuery
                ? "No users found matching your search"
                : "All available users are already in the group"}
            </p>
          )}
        </div>

        <div className="flex gap-3 justify-end pt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={sendingInvites}
            className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSendInvites}
            disabled={
              Object.values(invitedUsers).every((v) => !v) || sendingInvites
            }
            className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {sendingInvites ? "Sending..." : "Send Invites"}
          </button>
        </div>
      </div>
    </div>
  );
};

// Create Post Modal Component
const CreatePostModal = ({ groupId, onClose, onCreated }) => {
  const [content, setContent] = useState("");
  const [image, setImage] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    const trimmed = content.trim();
    if (!trimmed && !image) {
      setError("Post content or image is required.");
      return;
    }

    setSubmitting(true);
    setError("");

    const formData = new FormData();
    if (trimmed) formData.append("content", trimmed);
    if (image) formData.append("avatar", image);

    try {
      const result = await createGroupPost(groupId, formData);
      const newPost = result?.data || result;
      onCreated(newPost);
    } catch (err) {
      console.error("Failed to create group post:", err);
      setError(err?.message || "Failed to create post.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
        <h2 className="text-xl font-bold mb-4 text-purple-100">Create Post</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="relative">
            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={4}
              className="w-full px-3 py-2 pr-10 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:shadow-[0_0_10px_rgba(168,85,247,0.3)] text-purple-100 placeholder-purple-400/50 text-sm resize-none"
              placeholder="What's on your mind?"
              disabled={submitting}
            />
            <div className="absolute bottom-2 right-2">
              <EmojiPickerButton
                onEmojiSelect={(emoji) => setContent((prev) => prev + emoji)}
              />
            </div>
          </div>

          <div className="flex items-center gap-2">
            <label className="flex items-center gap-2 px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md cursor-pointer hover:border-purple-500/50 transition text-sm text-purple-300">
              <span className="opacity-60 text-base">📷</span>
              {image ? image.name : "Add Photo"}
              <input
                type="file"
                className="hidden"
                accept="image/*"
                onChange={(e) => setImage(e.target.files?.[0] || null)}
                disabled={submitting}
              />
            </label>
            {image && (
              <button
                type="button"
                onClick={() => setImage(null)}
                className="text-xs text-purple-400 hover:text-red-300 transition"
              >
                Remove
              </button>
            )}
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <div className="flex gap-3 justify-end">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] disabled:opacity-50"
            >
              {submitting ? "Posting..." : "Post"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// Create Event Modal Component
// Event Card Component
const EventCard = ({
  event,
  userRole,
  currentUserId,
  loading,
  onRespond,
  onDelete,
  isPast,
}) => {
  const myReaction = event.my_reaction;
  const canDelete =
    event.creator_id === currentUserId ||
    userRole === "owner" ||
    userRole === "moderator";

  return (
    <article
      className={`bg-[#1a1a2e] rounded-lg border p-4 transition-all ${
        isPast
          ? "border-purple-500/15 opacity-70"
          : "border-purple-500/30 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)]"
      }`}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <h3 className="font-semibold text-purple-100 text-lg">
            {event.title}
          </h3>
          {event.description && (
            <p className="text-purple-300/70 text-sm mt-1">
              {event.description}
            </p>
          )}
          <div className="flex items-center gap-4 text-sm text-purple-400 mt-2 flex-wrap">
            <span>📅 {event.event_day}</span>
            <span>🕐 {event.event_time}</span>
            <span className="text-green-400">✓ {event.going} going</span>
            <span className="text-red-400">✗ {event.not_going} not going</span>
            <span className="text-purple-400/60">
              📩 {event.pending} pending
            </span>
          </div>
        </div>
        <div className="flex flex-col items-end gap-2 shrink-0">
          {!isPast && (
            <div className="flex gap-2">
              <button
                onClick={() => onRespond("going")}
                disabled={loading || myReaction === "going"}
                className={`px-3 py-1.5 rounded-md transition cursor-pointer text-sm disabled:opacity-50 disabled:cursor-not-allowed ${
                  myReaction === "going"
                    ? "bg-green-600/30 text-green-300 border border-green-500/40"
                    : "bg-purple-600 hover:bg-purple-500 text-white shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                }`}
              >
                {myReaction === "going" ? "Going ✓" : "Going"}
              </button>
              <button
                onClick={() => onRespond("not_going")}
                disabled={loading || myReaction === "not_going"}
                className={`px-3 py-1.5 rounded-md transition cursor-pointer text-sm disabled:opacity-50 disabled:cursor-not-allowed ${
                  myReaction === "not_going"
                    ? "bg-red-600/30 text-red-300 border border-red-500/40"
                    : "bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30"
                }`}
              >
                {myReaction === "not_going" ? "Not Going ✗" : "Not Going"}
              </button>
            </div>
          )}
          {canDelete && (
            <button
              onClick={onDelete}
              className="text-xs text-purple-400 hover:text-red-300 transition"
            >
              Cancel Event
            </button>
          )}
        </div>
      </div>
    </article>
  );
};

// Create Event Modal Component
const CreateEventModal = ({ groupId, onClose, onCreated }) => {
  const [formData, setFormData] = useState({
    title: "",
    event_day: "",
    event_time: "",
    description: "",
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formData.title.trim()) {
      setError("Event title is required.");
      return;
    }
    if (!formData.event_day) {
      setError("Event date is required.");
      return;
    }
    if (!formData.event_time) {
      setError("Event time is required.");
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      await createGroupEvent(groupId, {
        title: formData.title.trim(),
        description: formData.description.trim(),
        event_day: formData.event_day,
        event_time: formData.event_time,
      });
      onCreated();
    } catch (err) {
      console.error("Failed to create event:", err);
      setError(err?.message || "Failed to create event.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
        <h2 className="text-xl font-bold mb-4 text-purple-100">Create Event</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-purple-300 mb-1">
              Event Title
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) =>
                setFormData({ ...formData, title: e.target.value })
              }
              className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 placeholder-purple-400/50 text-sm"
              placeholder="Enter event title..."
              disabled={submitting}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-purple-300 mb-1">
                Date
              </label>
              <input
                type="date"
                value={formData.event_day}
                onChange={(e) =>
                  setFormData({ ...formData, event_day: e.target.value })
                }
                className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 text-sm"
                disabled={submitting}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-purple-300 mb-1">
                Time
              </label>
              <input
                type="time"
                value={formData.event_time}
                onChange={(e) =>
                  setFormData({ ...formData, event_time: e.target.value })
                }
                className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 text-sm"
                disabled={submitting}
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-purple-300 mb-1">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) =>
                setFormData({ ...formData, description: e.target.value })
              }
              rows={3}
              className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 placeholder-purple-400/50 text-sm resize-none"
              placeholder="Describe your event..."
              disabled={submitting}
            />
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <div className="flex gap-3 justify-end">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] disabled:opacity-50"
            >
              {submitting ? "Creating..." : "Create Event"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

// Group Settings Modal Component (Owner only)
const GroupSettingsModal = ({ groupId, currentSettings, onClose, onSaved }) => {
  const toast = useToast();
  const [visibility, setVisibility] = useState(
    currentSettings?.visibility || "public",
  );
  const [joinMode, setJoinMode] = useState(
    currentSettings?.join_mode || "auto",
  );
  const [saving, setSaving] = useState(false);
  const [loadingSettings, setLoadingSettings] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        setLoadingSettings(true);
        const data = await getGroupSettings(groupId);
        if (data) {
          setVisibility(data.Visibility || data.visibility || "public");
          setJoinMode(data.JoinMode || data.join_mode || "auto");
        }
      } catch (err) {
        console.error("Failed to load group settings:", err);
      } finally {
        setLoadingSettings(false);
      }
    };
    load();
  }, [groupId]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const result = await updateGroupSettings(groupId, {
        visibility,
        join_mode: joinMode,
      });
      onSaved(result || { visibility, join_mode: joinMode });
    } catch (err) {
      console.error("Failed to update group settings:", err);
      toast.error(err?.message || "Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-[#1a1a2e] rounded-lg max-w-md w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
        <h2 className="text-xl font-bold mb-6 text-purple-100">
          Group Settings
        </h2>

        {loadingSettings ? (
          <p className="text-purple-400 text-sm text-center py-8">
            Loading settings...
          </p>
        ) : (
          <div className="space-y-5">
            {/* Visibility */}
            <div>
              <label className="block text-sm font-medium text-purple-300 mb-2">
                Visibility
              </label>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setVisibility("public")}
                  className={`flex-1 px-4 py-2 rounded-md text-sm font-medium transition cursor-pointer border ${
                    visibility === "public"
                      ? "bg-purple-600 text-white border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                      : "bg-[#0d0d1a] text-purple-400 border-purple-500/30 hover:border-purple-500/50"
                  }`}
                >
                  🌐 Public
                </button>
                <button
                  type="button"
                  onClick={() => setVisibility("private")}
                  className={`flex-1 px-4 py-2 rounded-md text-sm font-medium transition cursor-pointer border ${
                    visibility === "private"
                      ? "bg-purple-600 text-white border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                      : "bg-[#0d0d1a] text-purple-400 border-purple-500/30 hover:border-purple-500/50"
                  }`}
                >
                  🔒 Private
                </button>
              </div>
            </div>

            {/* Join Mode */}
            <div>
              <label className="block text-sm font-medium text-purple-300 mb-2">
                Join Mode
              </label>
              <div className="grid grid-cols-2 gap-2">
                {[
                  { value: "auto", label: "Auto" },
                  { value: "request", label: "Request" },
                  { value: "invite", label: "Invite Only" },
                  { value: "request_and_invite", label: "Request & Invite" },
                ].map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => setJoinMode(opt.value)}
                    className={`px-3 py-2 rounded-md text-sm font-medium transition cursor-pointer border ${
                      joinMode === opt.value
                        ? "bg-purple-600 text-white border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                        : "bg-[#0d0d1a] text-purple-400 border-purple-500/30 hover:border-purple-500/50"
                    }`}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        <div className="flex gap-3 justify-end pt-6">
          <button
            type="button"
            onClick={onClose}
            disabled={saving}
            className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={saving || loadingSettings}
            className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] disabled:opacity-50"
          >
            {saving ? "Saving..." : "Save Settings"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default GroupDetailPage;
