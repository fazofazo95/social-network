'use client';
import Image from "next/image";
import Link from "next/link";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { useEffect, useState } from "react";
import { fetchUserData } from "src/lib/services/user";
import { createPost, deletePost, getFeedPosts, getPostById, getUserPosts, updatePost } from "src/lib/services/post";
import { createComment, deleteComment, getPostComments, updateComment } from "src/lib/services/comment";
import { getFollowers, getFollowing } from "src/lib/services/follow";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import { getApiBaseUrl } from "src/lib/apiClient";
import EmojiPickerButton from "src/components/ui/EmojiPickerButton";


export default function App() {
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
        await deletePost(postId);
        setPosts((prev) => prev.filter((post) => post.id !== postId));
        if (editingPostId === postId) {
          setEditingPostId(null);
          setEditingPostContent("");
        }
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
    <main className="w-full max-w-2xl flex flex-col gap-20">
      <form
        onSubmit={handleSubmit}
        encType="multipart/form-data"
        className="bg-white w-full rounded-lg shadow-custom p-4 sticky top-16 z-10"
      >
        <div className="flex items-start gap-4 mb-4">
          <Image
            src={parseProfileImage(userData.profile_picture)}
            alt="Profile Icon"
            width={25}
            height={25}
          />
          <div className="relative w-full">
            <textarea
              value={postContent}
              onChange={(event) => setPostContent(event.target.value)}
              className="border rounded border-gray-200 text-black w-full h-20 focus:outline-none pl-2 pr-8 resize-none"
              placeholder="What's on your mind?"
              disabled={isSubmitting}
            />
            <div className="absolute bottom-2 right-2">
              <EmojiPickerButton onEmojiSelect={(emoji) => setPostContent((prev) => prev + emoji)} />
            </div>
          </div>
        </div>

        <ul className="flex gap-2 border-t border-gray-200 pt-2">
          <li className="flex  gap-1 hover:bg-gray-200   rounded-lg">
            <label htmlFor="photo-upload" className="flex items-center gap-1">
              <Image
                src="/photo_icon.svg"
                alt="Share Icon"
                width={20}
                height={20}
              />
              <input
                id="photo-upload"
                type="file"
                onChange={(event) => setPostImage(event.target.files?.[0] || null)}
                className="font-medium cursor-pointer text-black hidden"
                accept="image/*"
              />
              <span className="font-medium cursor-pointer text-black">
                Photo
              </span>
            </label>
          </li>
          <li className="flex items-center gap-2 hover:bg-gray-200 rounded-lg px-2 text-black text-sm">
            <label htmlFor="post-privacy" className="font-medium cursor-pointer">Privacy</label>
            <select
              id="post-privacy"
              className="font-medium cursor-pointer text-black bg-transparent focus:outline-none"
              value={postPrivacy}
              onChange={(event) => setPostPrivacy(event.target.value)}
              disabled={isSubmitting}
            >
              <option value="public">Public</option>
              <option value="followers">Followers</option>
              <option value="custom">Selective</option>
            </select>
          </li>
          <li className="flex bg-blue-500  hover:bg-blue-700 rounded-lg p-1 ml-auto">
            <Image
              src="/share_icon.svg"
              alt="Share Icon"
              width={20}
              height={20}
            />
            <button type="submit" className="text-white cursor-pointer" disabled={isSubmitting}>
              {isSubmitting ? "Posting..." : "Post"}
            </button>
          </li>
        </ul>
        {postPrivacy === "custom" ? (
          <div className="mt-2 border border-gray-200 rounded p-2">
            <p className="text-sm text-gray-600 mb-2">Choose users who can see this post:</p>
            {selectiveUsers.length === 0 ? (
              <p className="text-sm text-gray-500">No followers/following available to select.</p>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-1 max-h-32 overflow-y-auto pr-1 rounded-md bg-gray-50 p-2 [scrollbar-width:thin] [scrollbar-color:#9CA3AF_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-400 [&::-webkit-scrollbar-track]:bg-transparent">
                {selectiveUsers.map((user) => (
                  <label key={user.id} className="flex items-center gap-2 text-sm text-black">
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
        {submitError ? <p className="text-red-600 text-sm mt-2">{submitError}</p> : null}
      </form>

      {isLoading ? (
        <div className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
          Loading posts...
        </div>
      ) : posts.length === 0 ? (
        <div className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
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
            <article key={post.id} className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
            <div className="flex items-start justify-between gap-3 mb-2">
              <div className="flex items-start gap-2">
                <Image
                  src={parseProfileImage(post.author_profile_picture)}
                  alt="Profile Icon"
                  width={30}
                  height={30}
                />
                <div className="flex flex-col">
                  {post.user_id ? (
                    <Link href={`/profile/${post.user_id}`} className="font-bold text-lg leading-tight">
                      {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() || "Unknown User"}
                    </Link>
                  ) : (
                    <h1 className="font-bold text-lg leading-tight">
                      {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() || "Unknown User"}
                    </h1>
                  )}
                  {postDateLabel ? <span className="text-sm text-gray-500">{postDateLabel}</span> : null}
                </div>
              </div>
              {isOwnPost ? (
                <div className="flex gap-2">
                  <button
                    type="button"
                    className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
                    onClick={() => handleStartEditPost(post.id)}
                    disabled={isPostActionLoading}
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
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
                  className="border rounded px-2 py-1 text-sm flex-1"
                  value={editingPostContent}
                  onChange={(event) => setEditingPostContent(event.target.value)}
                />
                <button
                  type="button"
                  className="text-xs px-2 py-1 rounded bg-blue-500 text-white disabled:opacity-50"
                  onClick={() => handleSavePostEdit(post.id)}
                  disabled={isPostActionLoading}
                >
                  Save
                </button>
                <button
                  type="button"
                  className="text-xs px-2 py-1 rounded bg-gray-300 text-black"
                  onClick={() => {
                    setEditingPostId(null);
                    setEditingPostContent("");
                  }}
                >
                  Cancel
                </button>
              </div>
            ) : (
              <p>{post.content}</p>
            )}
            {postActionError ? <p className="text-red-600 text-sm mb-1">{postActionError}</p> : null}
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
              <span className="text-sm text-gray-500">{rippleCount} Ripples</span>
              <span className="text-sm text-gray-500">{comments.length} Echoes</span>
            </div>
            <div className="flex justify-between gap-8 mt-2 mx-8 border-t border-gray-200 pt-2">
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
              className="border-t border-gray-200 rounded mt-2 pt-2 gap-2 hidden flex-col"
            >
              <form onSubmit={(event) => handleCommentSubmit(event, post.id)} className="flex items-center gap-2 w-full">
                <Image
                  src={parseProfileImage(userData.profile_picture)}
                  alt="Profile Icon"
                  width={25}
                  height={25}
                />
                <div className="flex justify-between bg-gray-100 text-black w-full rounded-lg resize-none h-10">
                  <input
                    type="text"
                    placeholder="Write a comment..."
                    className="focus:outline-none w-full pl-1 bg-transparent"
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
                    className="flex items-center gap-1 cursor-pointer px-1"
                  >
                    <Image
                      src="/photo_icon.svg"
                      alt="Share Icon"
                      width={20}
                      height={20}
                    />
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
                  className="text-sm px-2 py-1 rounded bg-blue-500 text-white disabled:opacity-50"
                  disabled={isCommentSubmitting}
                >
                  {isCommentSubmitting ? "Sending..." : "Send"}
                </button>
              </form>

              {commentError ? <p className="text-red-600 text-sm">{commentError}</p> : null}

              <div className="flex flex-col gap-2">
                {isCommentsLoading ? (
                  <p className="text-sm text-gray-500">Loading echoes...</p>
                ) : comments.length === 0 ? (
                  <p className="text-sm text-gray-500">No echoes yet.</p>
                ) : (
                  comments.map((comment) => (
                    <div key={comment.id} className="bg-gray-50 rounded p-2">
                      <div className="flex items-start justify-between gap-2 mb-1">
                        <div className="flex items-start gap-2">
                          <div className="pt-0.5">
                            <Image
                              src={parseProfileImage(comment.author_profile_picture)}
                              alt="Comment author"
                              width={20}
                              height={20}
                            />
                          </div>
                          <div className="flex flex-col leading-tight">
                            {comment.user_id ? (
                              <Link href={`/profile/${comment.user_id}`} className="text-sm font-medium">
                                {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                              </Link>
                            ) : (
                              <span className="text-sm font-medium">
                                {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                              </span>
                            )}
                            {formatFriendlyDateTime(comment.created_at_time || comment.created_at) ? (
                              <span className="text-xs text-gray-500 mt-0.5">
                                {formatFriendlyDateTime(comment.created_at_time || comment.created_at)}
                              </span>
                            ) : null}
                          </div>
                        </div>
                        {comment.user_id === userData.id ? (
                          <div className="flex gap-2">
                            <button
                              type="button"
                              className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
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
                              className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
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
                            className="border rounded px-2 py-1 text-sm flex-1"
                            value={editingCommentContentByPost[post.id] || ""}
                            onChange={(event) =>
                              setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: event.target.value }))
                            }
                          />
                          <button
                            type="button"
                            className="text-xs px-2 py-1 rounded bg-blue-500 text-white disabled:opacity-50"
                            onClick={() => handleSaveCommentEdit(post.id, comment.id)}
                            disabled={!!commentActionLoadingById[comment.id]}
                          >
                            Save
                          </button>
                          <button
                            type="button"
                            className="text-xs px-2 py-1 rounded bg-gray-300 text-black"
                            onClick={() => {
                              setEditingCommentIdByPost((prev) => ({ ...prev, [post.id]: null }));
                              setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: "" }));
                            }}
                          >
                            Cancel
                          </button>
                        </div>
                      ) : (
                        <p className="text-sm">{comment.content}</p>
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
      )}
    </main>
  );
}
