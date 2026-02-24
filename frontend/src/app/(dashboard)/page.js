'use client';
import Image from "next/image";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { useEffect, useState } from "react";
import { fetchUserData } from "src/lib/services/user";
import { createPost, getUserPosts } from "src/lib/services/post";
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
        const userPosts = userId ? await getUserPosts(userId, 1, 10) : [];
        setUserData(profile || {});
        setPosts(Array.isArray(userPosts) ? userPosts : []);
      } catch (error) {
        console.error("Error loading dashboard:", error);
        setUserData({});
        setPosts([]);
      } finally {
        setIsLoading(false);
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
        formData.append("image", postImage);
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
          <li className="flex gap-1 hover:bg-gray-200  rounded-lg">
            <Image
              src="/feelings_icon.svg"
              alt="Share Icon"
              width={20}
              height={20}
            />
            <button
              type="button"
              className="font-medium cursor-pointer text-black"
            >
              Feeling
            </button>
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
            <div className="flex justify-between gap-8 mt-2 mx-8 border-t border-gray-200 pt-2">
              <Ripple_Button />
              <Echo_Button targetId={echoSectionId} />
              <button className="flex cursor-pointer gap-1">
                <Image
                  src="/spread_icon.svg"
                  alt="Spread Icon"
                  width={20}
                  height={20}
                />
                Spread
              </button>
            </div>
            <div
              id={echoSectionId}
              className="border-t border-gray-200 rounded mt-2 pt-2 gap-1 hidden"
            >
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
                  className="focus:outline-none w-full pl-1"
                />

                <label
                  htmlFor={echoPhotoUploadId}
                  className="flex items-center gap-1 cursor-pointer"
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
                  />
                </label>
              </div>
            </div>
            </article>
          );
        })
      )}
    </main>
  );
}
