'use client';
import Image from "next/image";
import Link from "next/link";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { useEffect, useState } from "react";
import { fetchUserData } from "src/lib/services/user";
import { createPost, deletePost, getFeedPosts, getPostById, getUserPosts, updatePost, restorePost } from "src/lib/services/post";
import { createComment, deleteComment, getPostComments, updateComment, restoreComment } from "src/lib/services/comment";
import { getFollowers, getFollowing } from "src/lib/services/follow";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import { getApiBaseUrl } from "src/lib/apiClient";
import EmojiPickerButton from "src/components/ui/EmojiPickerButton";
import { useToast } from "src/components/ui/Toast";


export default function App() {
  const toast = useToast();
  const [userData, setUserData] = useState({});
  const [posts, setPosts] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [postContent, setPostContent] = useState("");
  const [postImage, setPostImage] = useState(null);
  const [postPrivacy, setPostPrivacy] = useState("public");
  const [selectiveUserIds, setSelectiveUserIds] = useState([]);
  const [selectiveUsers, setSelectiveUsers] = useState([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [commentsByPost, setCommentsByPost] = useState({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState({});
  const [commentInputByPost, setCommentInputByPost] = useState({});
  const [commentImageByPost, setCommentImageByPost] = useState({});
  const [commentSubmittingByPost, setCommentSubmittingByPost] = useState({});
  const [commentErrorByPost, setCommentErrorByPost] = useState({});
  const [editingCommentIdByPost, setEditingCommentIdByPost] = useState({});
  const [editingCommentContentByPost, setEditingCommentContentByPost] = useState({});
  const [commentActionLoadingById, setCommentActionLoadingById] = useState({});
  const [editingPostId, setEditingPostId] = useState(null);
  const [editingPostContent, setEditingPostContent] = useState("");
  const [postActionLoadingById, setPostActionLoadingById] = useState({});
  const [postActionErrorById, setPostActionErrorById] = useState({});
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

   async function loadDashboardData() {
      setIsLoading(true);
      try {
        const profile = await fetchUserData("me");
        const userId = profile?.id;
        let loadedPosts = [];

        try {
          loadedPosts = await getFeedPosts(1);
        } catch (feedError) {
          console.warn("Feed load failed, falling back to user posts:", feedError);
          loadedPosts = userId ? await getUserPosts(userId, 1, 10) : [];
        }

        const [followers, following] = await Promise.all([
          getFollowers().catch(() => []),
          getFollowing().catch(() => []),
        ]);

        const usersById = new Map();
        [...followers, ...following].forEach((user) => {
          if (user?.id) {
            usersById.set(user.id, user);
          }
        });

        const safePosts = Array.isArray(loadedPosts) ? loadedPosts : [];

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

        setUserData(profile || {});
        setPosts(safePosts);
        setSelectiveUsers(Array.from(usersById.values()));
        setCommentsByPost(Object.fromEntries(commentsEntries));
        // Initialize ripple state for posts
        const rippleInit = {};
        safePosts.forEach(post => {
          rippleInit[post.id] = {
            count: post.likes_count || 0,
            rippled: !!post.has_current_user_liked
          };
        });
        setRippleStateByPost(rippleInit);
      } catch (error) {
        console.error("Error loading dashboard:", error);
        setUserData({});
        setPosts([]);
        setSelectiveUsers([]);
        setCommentsByPost({});
      } finally {
        setIsLoading(false);
      }
   }

   function handleToggleSelectiveUser(userId) {
      setSelectiveUserIds((prev) => {
        if (prev.includes(userId)) {
          return prev.filter((id) => id !== userId);
        }
        return [...prev, userId];
      });
   }

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
        console.error("Error updating comment:", error);
        setCommentErrorByPost((prev) => ({
          ...prev,
          [postId]: error?.message || "Failed to update echo.",
        }));
      } finally {
        setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: false }));
      }
   }

   async function handleStartEditPost(postId) {
      if (!postId) return;

      setPostActionErrorById((prev) => ({ ...prev, [postId]: "" }));
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
      try {
        const post = await getPostById(postId);
        setEditingPostId(postId);
        setEditingPostContent(post?.content || "");
      } catch (error) {
        console.error("Error loading post for edit:", error);
        setPostActionErrorById((prev) => ({
          ...prev,
          [postId]: error?.message || "Failed to load post.",
        }));
      } finally {
        setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
      }
   }

   async function handleSavePostEdit(postId) {
      const content = editingPostContent.trim();
      if (!content) {
        setPostActionErrorById((prev) => ({ ...prev, [postId]: "Post content is required." }));
        return;
      }

      setPostActionErrorById((prev) => ({ ...prev, [postId]: "" }));
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
      try {
        await updatePost(postId, content);
        setPosts((prev) => prev.map((post) => (post.id === postId ? { ...post, content } : post)));
        setEditingPostId(null);
        setEditingPostContent("");
      } catch (error) {
        console.error("Error updating post:", error);
        setPostActionErrorById((prev) => ({
          ...prev,
          [postId]: error?.message || "Failed to update post.",
        }));
      } finally {
        setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
      }
   }

   async function handleDeletePost(postId) {
      if (!postId) return;

      setPostActionErrorById((prev) => ({ ...prev, [postId]: "" }));
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
      try {
        const deletedPost = posts.find((p) => p.id === postId);
        await deletePost(postId);
        setPosts((prev) => prev.filter((post) => post.id !== postId));
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
                if (deletedPost) setPosts((prev) => [deletedPost, ...prev]);
              } catch (e) {
                toast.error(e?.message || "Failed to restore post");
              }
            },
          },
        });
      } catch (error) {
        console.error("Error deleting post:", error);
        setPostActionErrorById((prev) => ({
          ...prev,
          [postId]: error?.message || "Failed to delete post.",
        }));
      } finally {
        setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
      }
   }

   async function handleSubmit(event) {
      event.preventDefault();

      const trimmedContent = postContent.trim();
      if (!trimmedContent) {
        setSubmitError("Post content is required.");
        return;
      }

      if (postPrivacy === "custom" && selectiveUserIds.length === 0) {
        setSubmitError("Select at least one user for selective post privacy.");
        return;
      }

      setIsSubmitting(true);
      setSubmitError("");

      const formData = new FormData();
      formData.append("content", trimmedContent);
      formData.append("privacy", postPrivacy);

      if (postPrivacy === "custom") {
        selectiveUserIds.forEach((userId) => {
          formData.append("whitelisted_users", String(userId));
        });
      }

      if (postImage) {
        formData.append("avatar", postImage);
      }

      try {
        await createPost(formData);
        setPostContent("");
        setPostImage(null);
        setPostPrivacy("public");
        setSelectiveUserIds([]);
        await loadDashboardData();
      } catch (error) {
        console.error("Error creating post:", error);
        setSubmitError(error?.message || "Failed to create post.");
      } finally {
        setIsSubmitting(false);
      }
   }

   useEffect(() => {
      loadDashboardData();
    }, []);

  return ( 
    <main className="w-full max-w-2xl flex flex-col gap-6">
      <form
        onSubmit={handleSubmit}
        encType="multipart/form-data"
        className="bg-[#1a1a2e] w-full rounded-lg border border-purple-500/30 p-4 sticky top-16 z-10"
      >
        <div className="flex items-start gap-4 mb-4">
          <Image
            src={parseProfileImage(userData.profile_picture)}
            alt="Profile Icon"
            width={25}
            height={25}
            className="rounded-full shadow-[0_0_8px_rgba(168,85,247,0.3)]"
          />
          <div className="relative w-full">
            <textarea
              value={postContent}
              onChange={(event) => setPostContent(event.target.value)}
              className="bg-[#0d0d1a] border border-purple-500/30 rounded-md text-purple-100 placeholder-purple-400/50 w-full h-20 focus:outline-none focus:border-purple-500/50 pl-2 pr-8 resize-none"
              placeholder="What's on your mind?"
              disabled={isSubmitting}
            />
            <div className="absolute bottom-2 right-2">
              <EmojiPickerButton onEmojiSelect={(emoji) => setPostContent((prev) => prev + emoji)} />
            </div>
          </div>
        </div>

        <ul className="flex gap-2 border-t border-purple-500/20 pt-2">
          <li className="flex gap-1 hover:bg-purple-900/20 rounded-lg px-2 py-1">
            <label htmlFor="photo-upload" className="flex items-center gap-1 cursor-pointer">
              <Image
                src="/photo_icon.svg"
                alt="Share Icon"
                width={20}
                height={20}
                className="opacity-60"
              />
              <input
                id="photo-upload"
                type="file"
                onChange={(event) => setPostImage(event.target.files?.[0] || null)}
                className="hidden"
                accept="image/*"
              />
              <span className="font-medium cursor-pointer text-purple-300">
                Photo
              </span>
            </label>
          </li>
          <li className="flex items-center gap-2 hover:bg-purple-900/20 rounded-lg px-2 py-1 text-purple-300 text-sm">
            <label htmlFor="post-privacy" className="font-medium cursor-pointer">Privacy</label>
            <select
              id="post-privacy"
              className="font-medium cursor-pointer text-purple-300 bg-transparent focus:outline-none"
              value={postPrivacy}
              onChange={(event) => setPostPrivacy(event.target.value)}
              disabled={isSubmitting}
            >
              <option value="public" className="bg-[#1a1a2e]">Public</option>
              <option value="followers" className="bg-[#1a1a2e]">Followers</option>
              <option value="custom" className="bg-[#1a1a2e]">Selective</option>
            </select>
          </li>
          <li className="flex ml-auto">
            <button type="submit" className="flex items-center gap-2 text-xs bg-blue-500 hover:bg-blue-600 text-white rounded-md px-4 py-1.5 transition cursor-pointer disabled:opacity-50 font-semibold" disabled={isSubmitting}>
              <Image
                src="/share_icon.svg"
                alt="Share Icon"
                width={16}
                height={16}
                className="opacity-70"
              />
              {isSubmitting ? "Posting..." : "Post"}
            </button>
          </li>
        </ul>
        {postPrivacy === "custom" ? (
          <div className="mt-2 border border-purple-500/20 rounded-md p-2">
            <p className="text-sm text-purple-400 mb-2">Choose users who can see this post:</p>
            {selectiveUsers.length === 0 ? (
              <p className="text-sm text-purple-400/50">No followers/following available to select.</p>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-1 max-h-32 overflow-y-auto pr-1 rounded-md bg-[#0d0d1a] p-2 [scrollbar-width:thin] [scrollbar-color:#7c3aed_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-purple-600 [&::-webkit-scrollbar-track]:bg-transparent">
                {selectiveUsers.map((user) => (
                  <label key={user.id} className="flex items-center gap-2 text-sm text-purple-200">
                    <input
                      type="checkbox"
                      checked={selectiveUserIds.includes(user.id)}
                      onChange={() => handleToggleSelectiveUser(user.id)}
                      disabled={isSubmitting}
                    />
                    <span>{`${user.first_name || ""} ${user.last_name || ""}`.trim() || "Unknown User"}</span>
                  </label>
                ))}
              </div>
            )}
          </div>
        ) : null}
        {submitError ? <p className="text-red-400 text-sm mt-2">{submitError}</p> : null}
      </form>

      {isLoading ? (
        <div className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 w-full p-5 text-center text-purple-300">
          Loading posts...
        </div>
      ) : posts.length === 0 ? (
        <div className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 w-full p-5 text-center text-purple-300">
          No posts yet.
        </div>
      ) : (
        posts.map((post) => {
          const echoSectionId = `echo-section-${post.id}`;
          const echoPhotoUploadId = `echo-photo-upload-${post.id}`;
          const comments = commentsByPost[post.id] || [];
          const isCommentsLoading = commentsLoadingByPost[post.id];
          const commentValue = commentInputByPost[post.id] || "";
          const isCommentSubmitting = commentSubmittingByPost[post.id];
          const commentError = commentErrorByPost[post.id] || "";
          const isOwnPost = post.user_id === userData.id;
          const isEditingPost = editingPostId === post.id;
          const isPostActionLoading = !!postActionLoadingById[post.id];
          const postActionError = postActionErrorById[post.id] || "";
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
                <Image
                  src={parseProfileImage(post.author_profile_picture)}
                  alt="Profile Icon"
                  width={40}
                  height={40}
                  className="rounded-full shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                />
                <div className="flex flex-col">
                  {post.user_id ? (
                    <Link href={`/profile/${post.user_id}`} className="font-semibold text-purple-100 hover:underline">
                      {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() || "Unknown User"}
                    </Link>
                  ) : (
                    <span className="font-semibold text-purple-100">
                      {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() || "Unknown User"}
                    </span>
                  )}
                  {postDateLabel ? <span className="text-sm text-purple-400/60">{postDateLabel}</span> : null}
                </div>
              </div>
              {isOwnPost ? (
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
                    className="text-xs bg-purple-900/30 hover:bg-red-900/30 text-purple-300 hover:text-red-300 border border-purple-500/30 hover:border-red-500/30 rounded-md px-3 py-1 transition cursor-pointer disabled:opacity-50"
                    onClick={() => handleDeletePost(post.id)}
                    disabled={isPostActionLoading}
                  >
                    {isPostActionLoading ? "Working..." : "Delete"}
                  </button>
                </div>
              ) : null}
            </div>
            {isEditingPost ? (
              <div className="flex items-center gap-2 mb-2">
                <input
                  type="text"
                  className="flex-1 px-2 py-1 bg-[#0d0d1a] border border-purple-500/30 rounded-md text-purple-100 text-sm focus:outline-none focus:ring-1 focus:ring-purple-500"
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
              <p className="text-purple-300/80 mt-2">{post.content}</p>
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
            <div className="flex justify-between gap-4 mt-4">
              <span className="text-sm text-purple-400">{rippleCount} Ripples</span>
              <span className="text-sm text-purple-400">{comments.length} Echoes</span>
            </div>
            <div className="flex justify-between gap-8 mt-2 mx-8 border-t border-purple-500/20 pt-2">
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
              className="mt-4 pt-4 border-t border-purple-500/20 hidden flex-col gap-2"
            >
              <form onSubmit={(event) => handleCommentSubmit(event, post.id)} className="flex items-center gap-2 w-full">
                <Image
                  src={parseProfileImage(userData.profile_picture)}
                  alt="Profile Icon"
                  width={32}
                  height={32}
                  className="rounded-full shadow-[0_0_8px_rgba(168,85,247,0.3)]"
                />
                <div className="flex-1 flex items-center bg-[#0d0d1a] border border-purple-500/30 rounded-md">
                  <input
                    type="text"
                    placeholder="Write a comment..."
                    className="flex-1 px-3 py-2 bg-transparent focus:outline-none text-purple-100 placeholder-purple-400/50 text-sm"
                    value={commentValue}
                    onChange={(event) =>
                      setCommentInputByPost((prev) => ({ ...prev, [post.id]: event.target.value }))
                    }
                    disabled={isCommentSubmitting}
                  />

                  <EmojiPickerButton
                    onEmojiSelect={(emoji) =>
                      setCommentInputByPost((prev) => ({ ...prev, [post.id]: (prev[post.id] || "") + emoji }))
                    }
                  />

                  <label
                    htmlFor={echoPhotoUploadId}
                    className="flex items-center gap-1 cursor-pointer px-2"
                  >
                    <Image
                      src="/photo_icon.svg"
                      alt="Share Icon"
                      width={18}
                      height={18}
                      className="opacity-60"
                    />
                    <input
                      id={echoPhotoUploadId}
                      type="file"
                      className="hidden"
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
                  {isCommentSubmitting ? "..." : "Echo"}
                </button>
              </form>

              {commentError ? <p className="text-red-400 text-sm">{commentError}</p> : null}

              <div className="flex flex-col gap-2">
                {isCommentsLoading ? (
                  <p className="text-purple-400/50 text-sm text-center py-2">Loading echoes...</p>
                ) : comments.length === 0 ? (
                  <p className="text-purple-400/50 text-sm text-center py-2">No echoes yet.</p>
                ) : (
                  comments.map((comment) => (
                    <div key={comment.id} className="flex gap-2">
                      <div className="pt-0.5">
                        <Image
                          src={parseProfileImage(comment.author_profile_picture)}
                          alt="Comment author"
                          width={32}
                          height={32}
                          className="rounded-full"
                        />
                      </div>
                      <div className="flex-1 bg-[#0d0d1a] rounded-md p-3 border border-purple-500/20">
                      <div className="flex items-start justify-between gap-2 mb-1">
                        <div className="flex items-center gap-2">
                            {comment.user_id ? (
                              <Link href={`/profile/${comment.user_id}`} className="font-semibold text-purple-100 text-sm hover:underline">
                                {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                              </Link>
                            ) : (
                              <span className="font-semibold text-purple-100 text-sm">
                                {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                              </span>
                            )}
                            {formatFriendlyDateTime(comment.created_at_time || comment.created_at) ? (
                              <span className="text-purple-400/50 text-xs">
                                {formatFriendlyDateTime(comment.created_at_time || comment.created_at)}
                              </span>
                            ) : null}
                        </div>
                        {comment.user_id === userData.id ? (
                          <div className="flex gap-2">
                            <button
                              type="button"
                              className="text-xs text-purple-400 hover:text-purple-200 transition"
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
                              className="text-xs text-purple-400 hover:text-red-300 transition"
                              onClick={() => handleDeleteComment(post.id, comment.id)}
                              disabled={!!commentActionLoadingById[comment.id]}
                            >
                              Delete
                            </button>
                          </div>
                        ) : null}
                      </div>
                      {editingCommentIdByPost[post.id] === comment.id ? (
                        <div className="flex items-center gap-2 mt-1">
                          <input
                            type="text"
                            className="flex-1 px-2 py-1 bg-[#1a1a2e] border border-purple-500/30 rounded text-purple-100 text-sm focus:outline-none focus:ring-1 focus:ring-purple-500"
                            value={editingCommentContentByPost[post.id] || ""}
                            onChange={(event) =>
                              setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: event.target.value }))
                            }
                          />
                          <button
                            type="button"
                            className="text-xs px-2 py-1 rounded bg-purple-600 text-white disabled:opacity-50"
                            onClick={() => handleSaveCommentEdit(post.id, comment.id)}
                            disabled={!!commentActionLoadingById[comment.id]}
                          >
                            Save
                          </button>
                          <button
                            type="button"
                            className="text-xs px-2 py-1 rounded bg-purple-900/30 text-purple-300"
                            onClick={() => {
                              setEditingCommentIdByPost((prev) => ({ ...prev, [post.id]: null }));
                              setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: "" }));
                            }}
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <p className="text-purple-300/80 text-sm mt-1">{comment.content}</p>
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
                  ))
                )}
              </div>
            </div>
            </article>
          );
        })
      )}
    </main>
  );
}
