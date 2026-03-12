"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { fetchUserData, fetchVisibilitySettings, updateUserCover } from "src/lib/services/user";
import { deletePost, getPostById, getUserPosts, updatePost, restorePost } from "src/lib/services/post";
import {
  acceptFollowRequest,
  getBlockedUsers,
  getFollowers,
  getFollowing,
  getPendingRequests,
  removeFollower,
  rejectFollowRequest,
  unblockUser,
  unfollowUser,
} from "src/lib/services/follow";
import { createComment, deleteComment, getPostComments, updateComment, restoreComment } from "src/lib/services/comment";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import Avatar from "src/components/ui/Avatar";
import { getApiBaseUrl } from "src/lib/apiClient";
import { useToast } from "src/components/ui/Toast";

const Profile = () => {
  const toast = useToast();
  const [activeTab, setActiveTab] = useState("posts");
  const [profileData, setProfileData] = useState({});
  const [userPosts, setUserPosts] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [following, setFollowing] = useState([]);
  const [blockedUsers, setBlockedUsers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [commentsByPost, setCommentsByPost] = useState({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState({});
  const [commentInputByPost, setCommentInputByPost] = useState({});
  const [commentImageByPost, setCommentImageByPost] = useState({});
  const [commentSubmittingByPost, setCommentSubmittingByPost] = useState({});
  const [commentErrorByPost, setCommentErrorByPost] = useState({});
  const [editingCommentIdByPost, setEditingCommentIdByPost] = useState({});
  const [editingCommentContentByPost, setEditingCommentContentByPost] = useState({});
  const [commentActionLoadingById, setCommentActionLoadingById] = useState({});
  const [visibilitySettings, setVisibilitySettings] = useState(null);
  const [isSavingCover, setIsSavingCover] = useState(false);
  const [coverStatus, setCoverStatus] = useState("");
  const [editingPostId, setEditingPostId] = useState(null);
  const [editingPostContent, setEditingPostContent] = useState("");
  const [postActionLoadingById, setPostActionLoadingById] = useState({});
  const [postActionError, setPostActionError] = useState("");
  const [pendingRequests, setPendingRequests] = useState([]);
  const [acceptingByUserId, setAcceptingByUserId] = useState({});
  const [rejectingByUserId, setRejectingByUserId] = useState({});
  const [pendingError, setPendingError] = useState("");
  const [isRemovingByUserId, setIsRemovingByUserId] = useState({});
  const [isUnblockingByUserId, setIsUnblockingByUserId] = useState({});
  const [followListActionError, setFollowListActionError] = useState("");

  const [coverImage, setCoverImage] = useState("/example_cover.png");
  // Ripple state for all posts: { [postId]: { count, rippled } }
  const [rippleStateByPost, setRippleStateByPost] = useState({});

  function toUploadUrl(path) {
    if (!path) return "";
    if (path.startsWith("http://") || path.startsWith("https://") || path.startsWith("data:")) {
      return path;
    }
    if (path.startsWith("/uploads/")) {
      return `${getApiBaseUrl()}${path}`;
    }
    return "";
  }

  function toCoverUrl(path) {
    if (!path) return "/example_cover.png";
    if (path.startsWith("http://") || path.startsWith("https://") || path.startsWith("data:")) {
      return path;
    }
    if (path.startsWith("/uploads/")) {
      return `${getApiBaseUrl()}${path}`;
    }
    return "/example_cover.png";
  }

  async function loadProfilePageData() {
    setIsLoading(true);
    setError("");
    try {
      const [profile, settings] = await Promise.all([
        fetchUserData("me"),
        fetchVisibilitySettings().catch(() => null),
      ]);
      const userId = profile?.id;

      const [postsData, followersData, followingData, blockedData] = await Promise.all([
        userId ? getUserPosts(userId, 1, 10) : Promise.resolve([]),
        getFollowers(),
        getFollowing(),
        getBlockedUsers(),
      ]);

      const pendingData = await getPendingRequests().catch(() => []);

      setProfileData(profile || {});
      setCoverImage(toCoverUrl(profile?.cover_image));
      setVisibilitySettings(settings || null);
      setUserPosts(Array.isArray(postsData) ? postsData : []);
      // Initialize ripple state for posts
      const rippleInit = {};
      (Array.isArray(postsData) ? postsData : []).forEach(post => {
        rippleInit[post.id] = {
          count: post.likes_count || 0,
          rippled: !!post.has_current_user_liked
        };
      });
      setRippleStateByPost(rippleInit);
      setFollowers(Array.isArray(followersData) ? followersData : []);
      setFollowing(Array.isArray(followingData) ? followingData : []);
      setBlockedUsers(Array.isArray(blockedData) ? blockedData : []);
      setPendingRequests(Array.isArray(pendingData) ? pendingData : []);
      setPendingError("");

      const safePosts = Array.isArray(postsData) ? postsData : [];
      const commentsEntries = await Promise.all(
        safePosts.map(async (post) => {
          try {
            const comments = await getPostComments(post.id);
            return [post.id, comments];
          } catch {
            return [post.id, []];
          }
        })
      );
      setCommentsByPost(Object.fromEntries(commentsEntries));
    } catch (loadError) {
      console.error("Error loading profile page:", loadError);
      setProfileData({});
      setUserPosts([]);
      setFollowers([]);
      setFollowing([]);
      setBlockedUsers([]);
      setPendingRequests([]);
      setCommentsByPost({});
      setVisibilitySettings(null);
      setError(loadError?.message || "Failed to load profile.");
    } finally {
      setIsLoading(false);
    }
  }

  const handleChangeCover = async (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    setCoverStatus("");
    const formData = new FormData();
    formData.append("cover", file);

    try {
      setIsSavingCover(true);
      const updated = await updateUserCover(formData);
      const nextCover = updated?.cover_image || "";
      setCoverImage(toCoverUrl(nextCover));
      setProfileData((prev) => ({ ...prev, cover_image: nextCover }));
      setCoverStatus("Cover updated.");
    } catch (saveError) {
      console.error("Failed to update cover:", saveError);
      setCoverStatus(saveError?.message || "Failed to update cover.");
    } finally {
      setIsSavingCover(false);
      event.target.value = "";
    }
  };

  async function loadComments(postId) {
    setCommentsLoadingByPost((prev) => ({ ...prev, [postId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      const comments = await getPostComments(postId);
      setCommentsByPost((prev) => ({ ...prev, [postId]: comments }));
    } catch (loadError) {
      console.error("Error loading profile comments:", loadError);
      setCommentsByPost((prev) => ({ ...prev, [postId]: [] }));
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: loadError?.message || "Failed to load echoes.",
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
      setCommentErrorByPost((prev) => ({ ...prev, [postId]: "Comment content is required." }));
      return;
    }

    setCommentSubmittingByPost((prev) => ({ ...prev, [postId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));

    const formData = new FormData();
    formData.append("content", content);
    formData.append("parent_type", "post");
    formData.append("parent_id", String(postId));
    if (image) {
      formData.append("avatar", image);
    }

    try {
      await createComment(formData);
      setCommentInputByPost((prev) => ({ ...prev, [postId]: "" }));
      setCommentImageByPost((prev) => ({ ...prev, [postId]: null }));
      await loadComments(postId);
    } catch (submitError) {
      console.error("Error creating profile comment:", submitError);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: submitError?.message || "Failed to create echo.",
      }));
    } finally {
      setCommentSubmittingByPost((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleStartEditPost(postId) {
    if (!postId) return;
    setPostActionError("");
    setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
    try {
      const post = await getPostById(postId);
      setEditingPostId(postId);
      setEditingPostContent(post?.content || "");
    } catch (editError) {
      console.error("Failed to load post for editing:", editError);
      setPostActionError(editError?.message || "Failed to load post.");
    } finally {
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleSavePostEdit(postId) {
    const content = editingPostContent.trim();
    if (!content) {
      setPostActionError("Post content is required.");
      return;
    }

    setPostActionError("");
    setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
    try {
      await updatePost(postId, content);
      setUserPosts((prev) => prev.map((post) => (post.id === postId ? { ...post, content } : post)));
      setEditingPostId(null);
      setEditingPostContent("");
    } catch (saveError) {
      console.error("Failed to update post:", saveError);
      setPostActionError(saveError?.message || "Failed to update post.");
    } finally {
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleDeletePost(postId) {
    if (!postId) return;

    setPostActionError("");
    setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
    try {
      const deletedPost = userPosts.find((p) => p.id === postId);
      await deletePost(postId);
      setUserPosts((prev) => prev.filter((post) => post.id !== postId));
      if (editingPostId === postId) {
        setEditingPostId(null);
        setEditingPostContent("");
      }
      toast.success("Post deleted", {
        duration: 5000,
        action: {
          label: "Undo",
          onClick: async () => {
            try {
              await restorePost(postId);
              if (deletedPost) setUserPosts((prev) => [deletedPost, ...prev]);
            } catch (e) {
              toast.error(e?.message || "Failed to restore post");
            }
          },
        },
      });
    } catch (deleteError) {
      console.error("Failed to delete post:", deleteError);
      setPostActionError(deleteError?.message || "Failed to delete post.");
    } finally {
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
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
      console.error("Error deleting profile comment:", error);
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
      setCommentErrorByPost((prev) => ({ ...prev, [postId]: "Comment content is required." }));
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
      console.error("Error updating profile comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to update echo.",
      }));
    } finally {
      setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: false }));
    }
  }

  async function handleAcceptRequest(requestUserId) {
    if (!requestUserId) return;

    setPendingError("");
    setAcceptingByUserId((prev) => ({ ...prev, [requestUserId]: true }));

    try {
      await acceptFollowRequest(requestUserId);
      setPendingRequests((prev) => prev.filter((req) => req.id !== requestUserId));
      const refreshedFollowers = await getFollowers();
      setFollowers(Array.isArray(refreshedFollowers) ? refreshedFollowers : []);
    } catch (acceptError) {
      console.error("Failed to accept follow request:", acceptError);
      setPendingError(acceptError?.message || "Failed to accept request.");
    } finally {
      setAcceptingByUserId((prev) => ({ ...prev, [requestUserId]: false }));
    }
  }

  async function handleUnfollow(followedUserId) {
    if (!followedUserId) return;

    setFollowListActionError("");
    setIsRemovingByUserId((prev) => ({ ...prev, [followedUserId]: true }));

    try {
      await unfollowUser(followedUserId);
      setFollowing((prev) => prev.filter((user) => user.id !== followedUserId));
    } catch (removeError) {
      console.error("Failed to unfollow user:", removeError);
      setFollowListActionError(removeError?.message || "Failed to unfollow user.");
    } finally {
      setIsRemovingByUserId((prev) => ({ ...prev, [followedUserId]: false }));
    }
  }

  async function handleRemoveFollower(followerUserId) {
    if (!followerUserId) return;

    setFollowListActionError("");
    setIsRemovingByUserId((prev) => ({ ...prev, [followerUserId]: true }));

    try {
      await removeFollower(followerUserId);
      setFollowers((prev) => prev.filter((user) => user.id !== followerUserId));
    } catch (removeError) {
      console.error("Failed to remove follower:", removeError);
      setFollowListActionError(removeError?.message || "Failed to remove follower.");
    } finally {
      setIsRemovingByUserId((prev) => ({ ...prev, [followerUserId]: false }));
    }
  }

  async function handleRejectRequest(requestUserId) {
    if (!requestUserId) return;

    setPendingError("");
    setRejectingByUserId((prev) => ({ ...prev, [requestUserId]: true }));

    try {
      await rejectFollowRequest(requestUserId);
      setPendingRequests((prev) => prev.filter((req) => req.id !== requestUserId));
    } catch (rejectError) {
      console.error("Failed to reject follow request:", rejectError);
      setPendingError(rejectError?.message || "Failed to reject request.");
    } finally {
      setRejectingByUserId((prev) => ({ ...prev, [requestUserId]: false }));
    }
  }

  async function handleUnblockUser(targetUserId) {
    if (!targetUserId) return;

    setFollowListActionError("");
    setIsUnblockingByUserId((prev) => ({ ...prev, [targetUserId]: true }));

    try {
      await unblockUser(targetUserId);
      setBlockedUsers((prev) => prev.filter((user) => user.id !== targetUserId));
    } catch (unblockError) {
      console.error("Failed to unblock user:", unblockError);
      setFollowListActionError(unblockError?.message || "Failed to unblock user.");
    } finally {
      setIsUnblockingByUserId((prev) => ({ ...prev, [targetUserId]: false }));
    }
  }

  useEffect(() => {
    loadProfilePageData();
  }, []);

  const fullName = `${profileData.first_name || ""} ${profileData.last_name || ""}`.trim() || "Unknown User";
  const usernameText = profileData.nickname ? `@${profileData.nickname}` : "";
  const relationshipText = profileData.relationship_status || "";
  const locationText = profileData.location || "";
  const employedAtText = profileData.employed_at || "";
  const phoneText = profileData.phone_number || "";
  const emailText = profileData.email || "";
  const aboutText = profileData.about_me || "";
  const birthdayText = profileData.birthday_date || "";
  const privacySource = visibilitySettings?.profile_type || profileData.profile_type || "public";
  const privacyText = String(privacySource).toLowerCase() === "private" ? "Private" : "Public";
  const canShowFollowLists = profileData.own_profile || profileData.follow_vis !== "hidden";

  return (
    <div className="w-full max-w-2xl flex flex-col gap-6 pb-8">
      <main className="flex flex-col w-full bg-[#1a1a2e] rounded-lg border border-purple-500/30 overflow-hidden gap-2">
        <div
          className="w-full h-36 relative"
          style={{
            backgroundImage: `url('${coverImage}')`,
            backgroundSize: "cover",
            backgroundPosition: "center",
            backgroundRepeat: "no-repeat",
          }}
        >
          <label
            htmlFor="cover-upload"
            className="flex items-center gap-1 cursor-pointer absolute bottom-2 right-2 bg-purple-900/70 hover:bg-purple-900/90 p-1 px-2 rounded transition"
          >
            <span className="text-base">📸</span>
            <span className="text-sm text-purple-200">Change Cover</span>
            <input
              id="cover-upload"
              type="file"
              accept="image/*"
              onChange={handleChangeCover}
              disabled={isSavingCover}
              className="font-medium cursor-pointer text-purple-200 hidden"
            />
          </label>
        </div>
        {coverStatus ? <p className="text-xs text-purple-400 px-3">{coverStatus}</p> : null}

        <section className="border-b border-purple-500/20 pb-4 mb-2">
          <div className="flex items-center gap-2 justify-start">
            <div className="flex items-center gap-2 pl-5 pt-5">
              <Avatar
                src={profileData.profile_picture}
                name={fullName}
                size={50}
                className="shadow-[0_0_10px_rgba(168,85,247,0.3)]"
              />

              <div className="mb-4">
                <h1 className="text-3xl font-black text-purple-100">{fullName}</h1>
                <span className="text-purple-400 text-sm">{usernameText}</span>
              </div>
            </div>
          </div>

          <div className="flex justify-between items-center mx-10 gap-6 text-sm text-purple-400/60">
            <div className="flex flex-wrap items-center gap-6">
              {locationText ? (
                <span className="flex items-center gap-2">
                  <span className="text-sm">📍</span>
                  {locationText}
                </span>
              ) : null}
              {birthdayText ? (
                <span className="flex items-center gap-2 p-1">
                  <span className="text-sm">📅</span>
                  {birthdayText}
                </span>
              ) : null}
              <span className="flex items-center gap-2 p-1">
                <span className="text-sm">👁️</span>
                {privacyText}
              </span>
            </div>
            <Link href="/settings" className="flex items-center gap-2 rounded-lg px-3 py-1 text-sm bg-blue-500 hover:bg-blue-600 text-white cursor-pointer transition font-semibold">
              <span className="text-sm">✏️</span>
              Edit Profile
            </Link>
          </div>
        </section>

        <section className="flex justify-start gap-8 ml-5">
          <div className="flex flex-col items-center">
            <h1 className="text-4xl text-purple-100">{userPosts.length}</h1>
            <span className="text-purple-400">Posts</span>
          </div>
          {canShowFollowLists ? (
            <>
              <div className="flex flex-col items-center">
                <h1 className="text-4xl text-purple-100">{followers.length}</h1>
                <span className="text-purple-400">Followers</span>
              </div>
              <div className="flex flex-col items-center">
                <h1 className="text-4xl text-purple-100">{following.length}</h1>
                <span className="text-purple-400">Following</span>
              </div>
            </>
          ) : null}
        </section>

        <section className="text-purple-400 flex justify-around border-t border-purple-500/20 mt-4 pt-2 pb-2">
          <button type="button" onClick={() => setActiveTab("posts")} className={`cursor-pointer transition ${activeTab === 'posts' ? 'text-purple-200' : 'text-purple-400/60 hover:text-purple-300'}`}>
            Posts({userPosts.length})
          </button>
          <button type="button" onClick={() => setActiveTab("about")} className={`cursor-pointer transition ${activeTab === 'about' ? 'text-purple-200' : 'text-purple-400/60 hover:text-purple-300'}`}>
            About
          </button>
          {canShowFollowLists ? (
            <>
              <button type="button" onClick={() => setActiveTab("followers")} className={`cursor-pointer transition ${activeTab === 'followers' ? 'text-purple-200' : 'text-purple-400/60 hover:text-purple-300'}`}>
                Followers({followers.length})
              </button>
              <button type="button" onClick={() => setActiveTab("following")} className={`cursor-pointer transition ${activeTab === 'following' ? 'text-purple-200' : 'text-purple-400/60 hover:text-purple-300'}`}>
                Following({following.length})
              </button>
              <button type="button" onClick={() => setActiveTab("blocked")} className={`cursor-pointer transition ${activeTab === 'blocked' ? 'text-purple-200' : 'text-purple-400/60 hover:text-purple-300'}`}>
                Blocked({blockedUsers.length})
              </button>
            </>
          ) : null}
          <button type="button" onClick={() => setActiveTab("requests")} className={`cursor-pointer transition ${activeTab === 'requests' ? 'text-purple-200' : 'text-purple-400/60 hover:text-purple-300'}`}>
            Follow Requests({pendingRequests.length})
          </button>
        </section>
      </main>

      {isLoading ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-300 w-full p-5">
          Loading profile...
        </article>
      ) : error ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-red-400 w-full p-5">
          {error}
        </article>
      ) : null}

      {!isLoading && !error && activeTab === "posts" ? (
        userPosts.length === 0 ? (
          <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-300 w-full p-5">
            No posts yet.
          </article>
        ) : (
          userPosts.map((post) => {
            const echoSectionId = `profile-echo-section-${post.id}`;
            const echoPhotoUploadId = `profile-echo-photo-upload-${post.id}`;
            const comments = commentsByPost[post.id] || [];
            const isCommentsLoading = commentsLoadingByPost[post.id];
            const commentValue = commentInputByPost[post.id] || "";
            const isCommentSubmitting = commentSubmittingByPost[post.id];
            const commentError = commentErrorByPost[post.id] || "";
            const isPostActionLoading = !!postActionLoadingById[post.id];
            const isEditingPost = editingPostId === post.id;
            const postDateLabel = formatFriendlyDateTime(post.created_at_time || post.created_at);
            // Get ripple state for this post
            const rippleCount = rippleStateByPost[post.id]?.count ?? post.likes_count ?? 0;
            const rippled = rippleStateByPost[post.id]?.rippled ?? !!post.has_current_user_liked;
            const handleRippleChange = (newCount, newRippled) => {
              setRippleStateByPost(prev => ({
                ...prev,
                [post.id]: { count: newCount, rippled: newRippled }
              }));
            };
            return (
              <article key={post.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 w-full p-5 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                <div className="flex items-start justify-between gap-3 mb-2">
                  <div className="flex items-start gap-3">
                    <Avatar
                      src={profileData.profile_picture}
                      name={fullName}
                      size={40}
                      className="shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                    />
                    <div className="flex flex-col">
                      <Link href="/profile" className="font-semibold text-purple-100 hover:underline">{fullName}</Link>
                      {postDateLabel ? <span className="text-sm text-purple-400/60">{postDateLabel}</span> : null}
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition cursor-pointer disabled:opacity-50"
                      onClick={() => handleStartEditPost(post.id)}
                      disabled={isPostActionLoading}
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition cursor-pointer disabled:opacity-50"
                      onClick={() => handleDeletePost(post.id)}
                      disabled={isPostActionLoading}
                    >
                      {isPostActionLoading ? "Working..." : "Delete"}
                    </button>
                  </div>
                </div>
                {isEditingPost ? (
                  <div className="flex items-center gap-2 mb-2">
                    <input
                      type="text"
                      className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-2 py-1 text-sm text-purple-100 flex-1 focus:outline-none focus:border-purple-500/50"
                      value={editingPostContent}
                      onChange={(event) => setEditingPostContent(event.target.value)}
                    />
                    <button
                      type="button"
                      className="text-xs px-2 py-1 rounded-md bg-purple-600 text-white disabled:opacity-50"
                      onClick={() => handleSavePostEdit(post.id)}
                      disabled={isPostActionLoading}
                    >
                      Save
                    </button>
                    <button
                      type="button"
                      className="text-xs px-2 py-1 rounded-md bg-purple-900/30 text-purple-300 border border-purple-500/30"
                      onClick={() => {
                        setEditingPostId(null);
                        setEditingPostContent("");
                      }}
                    >
                      Cancel
                    </button>
                  </div>
                ) : (
                  <p className="text-purple-200">{post.content}</p>
                )}
                {postActionError ? <p className="text-red-400 text-sm mb-1">{postActionError}</p> : null}
                {post.image ? (
                  <div className="mt-3">
                    <Image
                      src={toUploadUrl(post.image)}
                      alt="Post image"
                      width={500}
                      height={300}
                      className="rounded-lg w-full h-auto"
                    />
                  </div>
                ) : null}
                <div className="flex justify-end gap-4 mt-2 border-b border-purple-500/20 pb-1">
                  <span className="text-purple-400/60 text-sm mr-auto">{rippleCount} Ripples</span>
                  <span className="text-purple-400/60 text-sm">{comments.length} Echoes</span>
                </div>
                <div className="flex justify-between gap-8 mt-2 mx-8">
                  <Ripple_Button 
                    postId={post.id}
                    initialRippled={rippled}
                    initialCount={rippleCount}
                    onChange={handleRippleChange}
                  />
                  <Echo_Button
                    targetId={echoSectionId}
                    onToggle={(isOpen) => {
                      if (isOpen && commentsByPost[post.id] === undefined) {
                        loadComments(post.id);
                      }
                    }}
                  />
                </div>
                <div
                  id={echoSectionId}
                  className="border-t border-purple-500/20 rounded mt-2 pt-2 gap-2 hidden flex-col"
                >
                  <form onSubmit={(event) => handleCommentSubmit(event, post.id)} className="flex items-center gap-2 w-full">
                    <Avatar
                      src={profileData.profile_picture}
                      name={fullName}
                      size={25}
                      className="shadow-[0_0_8px_rgba(168,85,247,0.3)]"
                    />
                    <div className="flex justify-between bg-[#0d0d1a] border border-purple-500/30 text-purple-100 w-full rounded-lg resize-none h-10">
                      <input
                        type="text"
                        placeholder="Write a comment..."
                        className="focus:outline-none w-full pl-2 bg-transparent placeholder-purple-400/50"
                        value={commentValue}
                        onChange={(event) =>
                          setCommentInputByPost((prev) => ({ ...prev, [post.id]: event.target.value }))
                        }
                        disabled={isCommentSubmitting}
                      />

                      <label
                        htmlFor={echoPhotoUploadId}
                        className="flex items-center gap-1 cursor-pointer px-1"
                      >
                        <span className="text-lg">📷</span>
                        <input
                          id={echoPhotoUploadId}
                          type="file"
                          className="font-medium cursor-pointer text-black hidden"
                          accept="image/*"
                          onChange={(event) =>
                            setCommentImageByPost((prev) => ({ ...prev, [post.id]: event.target.files?.[0] || null }))
                          }
                          disabled={isCommentSubmitting}
                        />
                      </label>
                    </div>
                    <button
                      type="submit"
                      className="px-3 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] disabled:opacity-50"
                      disabled={isCommentSubmitting}
                    >
                      {isCommentSubmitting ? "Sending..." : "Echo"}
                    </button>
                  </form>

                  {commentError ? <p className="text-red-400 text-sm">{commentError}</p> : null}

                  <div className="flex flex-col gap-2">
                    {isCommentsLoading ? (
                      <p className="text-sm text-purple-400/60">Loading echoes...</p>
                    ) : comments.length === 0 ? (
                      <p className="text-sm text-purple-400/60">No echoes yet.</p>
                    ) : (
                      comments.map((comment) => (
                        <div key={comment.id} className="bg-[#0d0d1a] rounded-md border border-purple-500/20 p-3">
                          <div className="flex items-start justify-between gap-2 mb-1">
                            <div className="flex items-start gap-2">
                              <div className="pt-0.5">
                                <Avatar
                                  src={comment.author_profile_picture}
                                  name={`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim()}
                                  size={20}
                                />
                              </div>
                              <div className="flex flex-col leading-tight">
                                {comment.user_id ? (
                                  <Link
                                    href={comment.user_id === profileData.id ? "/profile" : `/profile/${comment.user_id}`}
                                    className="text-sm font-medium text-purple-200 hover:underline"
                                  >
                                    {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                                  </Link>
                                ) : (
                                  <span className="text-sm font-medium text-purple-200">
                                    {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                                  </span>
                                )}
                                {formatFriendlyDateTime(comment.created_at_time || comment.created_at) ? (
                                  <span className="text-xs text-purple-400/60 mt-0.5">
                                    {formatFriendlyDateTime(comment.created_at_time || comment.created_at)}
                                  </span>
                                ) : null}
                              </div>
                            </div>
                            {comment.user_id === profileData.id ? (
                              <div className="flex gap-2">
                                <button
                                  type="button"
                                  className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition cursor-pointer disabled:opacity-50"
                                  onClick={() => {
                                    setEditingCommentIdByPost((prev) => ({ ...prev, [post.id]: comment.id }));
                                    setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: comment.content || "" }));
                                  }}
                                  disabled={!!commentActionLoadingById[comment.id]}
                                >
                                  Edit
                                </button>
                                <button
                                  type="button"
                                  className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition cursor-pointer disabled:opacity-50"
                                  onClick={() => handleDeleteComment(post.id, comment.id)}
                                  disabled={!!commentActionLoadingById[comment.id]}
                                >
                                  Delete
                                </button>
                              </div>
                            ) : null}
                          </div>
                          {editingCommentIdByPost[post.id] === comment.id ? (
                            <div className="flex items-center gap-2">
                              <input
                                type="text"
                                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-2 py-1 text-sm text-purple-100 flex-1 focus:outline-none focus:border-purple-500/50"
                                value={editingCommentContentByPost[post.id] || ""}
                                onChange={(event) =>
                                  setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: event.target.value }))
                                }
                              />
                              <button
                                type="button"
                                className="text-xs px-2 py-1 rounded-md bg-purple-600 text-white disabled:opacity-50"
                                onClick={() => handleSaveCommentEdit(post.id, comment.id)}
                                disabled={!!commentActionLoadingById[comment.id]}
                              >
                                Save
                              </button>
                              <button
                                type="button"
                                className="text-xs px-2 py-1 rounded-md bg-purple-900/30 text-purple-300 border border-purple-500/30"
                                onClick={() => {
                                  setEditingCommentIdByPost((prev) => ({ ...prev, [post.id]: null }));
                                  setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: "" }));
                                }}
                              >
                                Cancel
                              </button>
                            </div>
                          ) : (
                            <p className="text-sm text-purple-200">{comment.content}</p>
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
                      ))
                    )}
                  </div>
                </div>
              </article>
            );
          })
        )
      ) : null}

      {!isLoading && !error && activeTab === "about" ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl mb-1">User Information</h1>
          <h2 className="font-semibold text-sm text-purple-300 mb-2">Contact Information</h2>
          <ul className="text-sm">
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Email:</span>
              <span>{emailText || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Full Name:</span>
              <span>{fullName || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Nickname:</span>
              <span>{usernameText || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Date of Birth:</span>
              <span>{birthdayText || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Location:</span>
              <span>{locationText || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Relationship:</span>
              <span>{relationshipText || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Employed At:</span>
              <span>{employedAtText || "-"}</span>
            </li>
            <li className="flex justify-between gap-4 py-1 border-b border-purple-500/20">
              <span className="font-semibold">Phone:</span>
              <span>{phoneText || "-"}</span>
            </li>
          </ul>
          <div className="mt-4">
            <h3 className="font-semibold text-sm mb-1">About me</h3>
            <p className="text-sm text-purple-200">{aboutText || "No about info yet."}</p>
          </div>
        </article>
      ) : null}

      {!isLoading && !error && canShowFollowLists && activeTab === "followers" ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl text-purple-200 mb-3">Followers ({followers.length})</h1>
          {followListActionError ? <p className="text-red-400 text-sm mb-3">{followListActionError}</p> : null}
          {followers.length === 0 ? (
            <p className="text-sm text-purple-300">No followers yet.</p>
          ) : (
            <ul className="flex flex-col gap-2.5">
              {followers.map((follower) => (
                <li key={follower.id} className="flex items-center gap-3 rounded-md border border-purple-500/20 bg-[#0d0d1a] px-3 py-2">
                  <Avatar
                    src={follower.profile_picture}
                    name={`${follower.first_name || ""} ${follower.last_name || ""}`.trim()}
                    size={24}
                    className="h-6 w-6"
                  />
                  <span className="flex-1 min-w-0">
                    <span className="block truncate text-sm font-semibold text-purple-100">{`${follower.first_name || ""} ${follower.last_name || ""}`.trim() || "Unknown User"}</span>
                    {follower.username ? (
                      <span className="block truncate text-[11px] text-purple-400">@{follower.username}</span>
                    ) : null}
                  </span>
                  <Link
                    href={`/profile/${follower.id}`}
                    className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition"
                  >
                    View profile
                  </Link>
                  <button
                    type="button"
                    className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition disabled:opacity-50"
                    onClick={() => handleRemoveFollower(follower.id)}
                    disabled={!!isRemovingByUserId[follower.id]}
                  >
                    {!!isRemovingByUserId[follower.id] ? "Removing..." : "Remove"}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </article>
      ) : null}

      {!isLoading && !error && canShowFollowLists && activeTab === "following" ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl text-purple-200 mb-3">Following ({following.length})</h1>
          {followListActionError ? <p className="text-red-400 text-sm mb-3">{followListActionError}</p> : null}
          {following.length === 0 ? (
            <p className="text-sm text-purple-300">Not following anyone yet.</p>
          ) : (
            <ul className="flex flex-col gap-2.5">
              {following.map((followedUser) => (
                <li key={followedUser.id} className="flex items-center gap-3 rounded-md border border-purple-500/20 bg-[#0d0d1a] px-3 py-2">
                  <Avatar
                    src={followedUser.profile_picture}
                    name={`${followedUser.first_name || ""} ${followedUser.last_name || ""}`.trim()}
                    size={24}
                    className="h-6 w-6"
                  />
                  <span className="flex-1 min-w-0">
                    <span className="block truncate text-sm font-semibold text-purple-100">{`${followedUser.first_name || ""} ${followedUser.last_name || ""}`.trim() || "Unknown User"}</span>
                    {followedUser.username ? (
                      <span className="block truncate text-[11px] text-purple-400">@{followedUser.username}</span>
                    ) : null}
                  </span>
                  <Link
                    href={`/profile/${followedUser.id}`}
                    className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition"
                  >
                    View profile
                  </Link>
                  <button
                    type="button"
                    className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition disabled:opacity-50"
                    onClick={() => handleUnfollow(followedUser.id)}
                    disabled={!!isRemovingByUserId[followedUser.id]}
                  >
                    {!!isRemovingByUserId[followedUser.id] ? "Removing..." : "Unfollow"}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </article>
      ) : null}

      {!isLoading && !error && canShowFollowLists && activeTab === "blocked" ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl text-purple-200 mb-3">Blocked ({blockedUsers.length})</h1>
          {followListActionError ? <p className="text-red-400 text-sm mb-3">{followListActionError}</p> : null}
          {blockedUsers.length === 0 ? (
            <p className="text-sm text-purple-300">No blocked users.</p>
          ) : (
            <ul className="flex flex-col gap-2.5">
              {blockedUsers.map((blockedUser) => (
                <li key={blockedUser.id} className="flex items-center gap-3 rounded-md border border-purple-500/20 bg-[#0d0d1a] px-3 py-2">
                  <Avatar
                    src={blockedUser.profile_picture}
                    name={`${blockedUser.first_name || ""} ${blockedUser.last_name || ""}`.trim()}
                    size={24}
                    className="h-6 w-6"
                  />
                  <span className="flex-1 min-w-0">
                    <span className="block truncate text-sm font-semibold text-purple-100">{`${blockedUser.first_name || ""} ${blockedUser.last_name || ""}`.trim() || "Unknown User"}</span>
                    {blockedUser.username ? (
                      <span className="block truncate text-[11px] text-purple-400">@{blockedUser.username}</span>
                    ) : null}
                  </span>
                  <Link
                    href={`/profile/${blockedUser.id}`}
                    className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition"
                  >
                    View profile
                  </Link>
                  <button
                    type="button"
                    className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition disabled:opacity-50"
                    onClick={() => handleUnblockUser(blockedUser.id)}
                    disabled={!!isUnblockingByUserId[blockedUser.id]}
                  >
                    {!!isUnblockingByUserId[blockedUser.id] ? "Unblocking..." : "Unblock"}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </article>
      ) : null}

      {!isLoading && !error && activeTab === "requests" ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl text-purple-200 mb-3">Pending Requests ({pendingRequests.length})</h1>
          {pendingError ? <p className="text-red-400 text-sm mb-3">{pendingError}</p> : null}
          {pendingRequests.length === 0 ? (
            <p className="text-sm text-purple-300">No pending requests.</p>
          ) : (
            <ul className="flex flex-col gap-2.5">
              {pendingRequests.map((requestUser) => {
                const isAccepting = !!acceptingByUserId[requestUser.id];
                const isRejecting = !!rejectingByUserId[requestUser.id];
                return (
                  <li key={requestUser.id} className="flex items-center gap-3 rounded-md border border-purple-500/20 bg-[#0d0d1a] px-3 py-2">
                    <Avatar
                      src={requestUser.profile_picture}
                      name={`${requestUser.first_name || ""} ${requestUser.last_name || ""}`.trim()}
                      size={24}
                      className="h-6 w-6"
                    />
                    <span className="flex-1 min-w-0">
                      <span className="block truncate text-sm font-semibold text-purple-100">{`${requestUser.first_name || ""} ${requestUser.last_name || ""}`.trim() || "Unknown User"}</span>
                      {requestUser.username ? (
                        <span className="block truncate text-[11px] text-purple-400">@{requestUser.username}</span>
                      ) : null}
                    </span>
                    <Link
                      href={`/profile/${requestUser.id}`}
                      className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition"
                    >
                      View profile
                    </Link>
                    <button
                      type="button"
                      className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition disabled:opacity-50"
                      onClick={() => handleAcceptRequest(requestUser.id)}
                      disabled={isAccepting || isRejecting}
                    >
                      {isAccepting ? "Accepting..." : "Accept"}
                    </button>
                    <button
                      type="button"
                      className="text-xs bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1 transition disabled:opacity-50"
                      onClick={() => handleRejectRequest(requestUser.id)}
                      disabled={isAccepting || isRejecting}
                    >
                      {isRejecting ? "Rejecting..." : "Reject"}
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </article>
      ) : null}
    </div>
  );
};

export default Profile;
