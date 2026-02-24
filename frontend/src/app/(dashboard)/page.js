'use client';
import Image from "next/image";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { useEffect, useState } from "react";
import { fetchUserData } from "src/lib/services/user";
import { createPost, getFeedPosts, getUserPosts } from "src/lib/services/post";
import { createComment, getPostComments } from "src/lib/services/comment";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { getApiBaseUrl } from "src/lib/apiClient";


export default function App() {
   const [userData, setUserData] = useState({});
   const [posts, setPosts] = useState([]);
   const [isLoading, setIsLoading] = useState(true);
   const [postContent, setPostContent] = useState("");
   const [postImage, setPostImage] = useState(null);
   const [isSubmitting, setIsSubmitting] = useState(false);
   const [submitError, setSubmitError] = useState("");
  const [commentsByPost, setCommentsByPost] = useState({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState({});
  const [commentInputByPost, setCommentInputByPost] = useState({});
  const [commentImageByPost, setCommentImageByPost] = useState({});
  const [commentSubmittingByPost, setCommentSubmittingByPost] = useState({});
  const [commentErrorByPost, setCommentErrorByPost] = useState({});

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
        setCommentsByPost(Object.fromEntries(commentsEntries));
      } catch (error) {
        console.error("Error loading dashboard:", error);
        setUserData({});
        setPosts([]);
        setCommentsByPost({});
      } finally {
        setIsLoading(false);
      }
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

   async function handleSubmit(event) {
      event.preventDefault();

      const trimmedContent = postContent.trim();
      if (!trimmedContent) {
        setSubmitError("Post content is required.");
        return;
      }

      setIsSubmitting(true);
      setSubmitError("");

      const formData = new FormData();
      formData.append("content", trimmedContent);
      formData.append("privacy", "public");

      if (postImage) {
        formData.append("avatar", postImage);
      }

      try {
        await createPost(formData);
        setPostContent("");
        setPostImage(null);
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

          return (
            <article key={post.id} className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
            <div className="flex items-center gap-2">
              <Image
                src={parseProfileImage(post.author_profile_picture)}
                alt="Profile Icon"
                width={30}
                height={30}
              />
              <h1 className="font-bold text-lg">
                {`${post.author_first_name || ""} ${post.author_last_name || ""}`.trim() || "Unknown User"}
              </h1>
            </div>
            <span className="text-sm text-gray-500 ml-4 mb-2">{post.created_at || ""}</span>
            <p>{post.content}</p>
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
                      <div className="flex items-center gap-2 mb-1">
                        <Image
                          src={parseProfileImage(comment.author_profile_picture)}
                          alt="Comment author"
                          width={20}
                          height={20}
                        />
                        <span className="text-sm font-medium">
                          {`${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User"}
                        </span>
                      </div>
                      <p className="text-sm">{comment.content}</p>
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
