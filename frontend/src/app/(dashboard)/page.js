'use client';
import Image from "next/image";
import Link from "next/link";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { useEffect, useState } from "react";
import { fetchUserData } from "src/lib/services/user";
import { createPost, getFeedPosts, getUserPosts } from "src/lib/services/post";
import { getFollowers, getFollowing } from "src/lib/services/follow";
import { getPostComments } from "src/lib/services/comment";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import { usePostComments } from "src/lib/hooks/usePostComments";
import { toUploadUrl } from "src/lib/utils/mediaUrl";
import CommentThread from "src/components/posts/CommentThread";


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
  const {
    commentsByPost,
    setCommentsByPost,
    commentsLoadingByPost,
    commentInputByPost,
    setCommentInputByPost,
    commentImageByPost,
    setCommentImageByPost,
    commentSubmittingByPost,
    commentErrorByPost,
    editingCommentIdByPost,
    setEditingCommentIdByPost,
    editingCommentContentByPost,
    setEditingCommentContentByPost,
    commentActionLoadingById,
    editingPostId,
    setEditingPostId,
    editingPostContent,
    setEditingPostContent,
    postActionLoadingById,
    postActionErrorById,
    loadComments,
    handleCommentSubmit,
    handleDeleteComment,
    handleSaveCommentEdit,
    handleStartEditPost,
    handleSavePostEdit,
    handleDeletePost,
  } = usePostComments({ setPosts });

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
          <textarea
            value={postContent}
            onChange={(event) => setPostContent(event.target.value)}
            className="border rounded border-gray-200 text-black w-full h-20 focus:outline-none pl-2 resize-none"
            placeholder="What's on your mind?"
            disabled={isSubmitting}
          />
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
              <span className="text-sm text-gray-500">{post.likes_count || 0} Likes</span>
              <span className="text-sm text-gray-500">{comments.length} Echoes</span>
            </div>
            <div className="flex justify-between gap-8 mt-2 mx-8 border-t border-gray-200 pt-2">
              <Ripple_Button />
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
              <CommentThread
                postId={post.id}
                currentUserId={userData.id}
                currentUserProfilePicture={userData.profile_picture}
                commentValue={commentValue}
                onCommentChange={(value) =>
                  setCommentInputByPost((prev) => ({ ...prev, [post.id]: value }))
                }
                onCommentImageChange={(file) =>
                  setCommentImageByPost((prev) => ({ ...prev, [post.id]: file }))
                }
                onSubmit={(event) => handleCommentSubmit(event, post.id)}
                isCommentSubmitting={isCommentSubmitting}
                commentError={commentError}
                comments={comments}
                isCommentsLoading={isCommentsLoading}
                editingCommentId={editingCommentIdByPost[post.id]}
                editingCommentContent={editingCommentContentByPost[post.id] || ""}
                onStartEdit={(comment) => {
                  setEditingCommentIdByPost((prev) => ({ ...prev, [post.id]: comment.id }));
                  setEditingCommentContentByPost((prev) => ({
                    ...prev,
                    [post.id]: comment.content || "",
                  }));
                }}
                onEditContentChange={(value) =>
                  setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: value }))
                }
                onSaveEdit={(commentId) => handleSaveCommentEdit(post.id, commentId)}
                onCancelEdit={() => {
                  setEditingCommentIdByPost((prev) => ({ ...prev, [post.id]: null }));
                  setEditingCommentContentByPost((prev) => ({ ...prev, [post.id]: "" }));
                }}
                onDelete={(commentId) => handleDeleteComment(post.id, commentId)}
                commentActionLoadingById={commentActionLoadingById}
                toUploadUrl={toUploadUrl}
              />
            </div>
            </article>
          );
        })
      )}
    </main>
  );
}
