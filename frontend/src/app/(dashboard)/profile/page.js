"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
import { fetchUserData, fetchVisibilitySettings, updateUserCover } from "src/lib/services/user";
import { getUserPosts } from "src/lib/services/post";
import { getFollowersByUser, getFollowingByUser } from "src/lib/services/follow";
import { createComment, getPostComments } from "src/lib/services/comment";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { getApiBaseUrl } from "src/lib/apiClient";

const Profile = () => {
  const [activeTab, setActiveTab] = useState("posts");
  const [profileData, setProfileData] = useState({});
  const [userPosts, setUserPosts] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [following, setFollowing] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [commentsByPost, setCommentsByPost] = useState({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState({});
  const [commentInputByPost, setCommentInputByPost] = useState({});
  const [commentImageByPost, setCommentImageByPost] = useState({});
  const [commentSubmittingByPost, setCommentSubmittingByPost] = useState({});
  const [commentErrorByPost, setCommentErrorByPost] = useState({});
  const [visibilitySettings, setVisibilitySettings] = useState(null);
  const [isSavingCover, setIsSavingCover] = useState(false);
  const [coverStatus, setCoverStatus] = useState("");

  const [coverImage, setCoverImage] = useState("/example_cover.png");

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

      const [postsData, followersData, followingData] = await Promise.all([
        userId ? getUserPosts(userId, 1, 10) : Promise.resolve([]),
        userId ? getFollowersByUser(userId) : Promise.resolve([]),
        userId ? getFollowingByUser(userId) : Promise.resolve([]),
      ]);

      setProfileData(profile || {});
      setCoverImage(toCoverUrl(profile?.cover_image));
      setVisibilitySettings(settings || null);
      setUserPosts(Array.isArray(postsData) ? postsData : []);
      setFollowers(Array.isArray(followersData) ? followersData : []);
      setFollowing(Array.isArray(followingData) ? followingData : []);

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

  useEffect(() => {
    loadProfilePageData();
  }, []);

  const fullName = `${profileData.first_name || ""} ${profileData.last_name || ""}`.trim() || "Unknown User";
  const usernameText = profileData.nickname ? `@${profileData.nickname}` : "";
  const relationshipText = profileData.relationship_status || profileData.current_status || "";
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
    <div className="w-full max-w-2xl flex flex-col gap-10">
      <main className="flex flex-col w-full bg-white rounded-lg overflow-hidden gap-2">
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
            className="flex items-center gap-1 cursor-pointer absolute bottom-2 right-2 bg-gray-200 bg-opacity-70 p-1 rounded"
          >
            <Image src="/cover_icon.svg" alt="Cover" width={20} height={20} />
            <span className="text-sm text-black">Change Cover</span>
            <input
              id="cover-upload"
              type="file"
              accept="image/*"
              onChange={handleChangeCover}
              disabled={isSavingCover}
              className="font-medium cursor-pointer text-black hidden"
            />
          </label>
        </div>
        {coverStatus ? <p className="text-xs text-gray-500 px-3">{coverStatus}</p> : null}

        <section className="border-b border-gray-200 pb-4 mb-2">
          <div className="flex items-center gap-2 justify-start">
            <div className="flex items-center gap-2 pl-5 pt-5">
              <Image
                src={parseProfileImage(profileData.profile_picture)}
                alt="Profile Picture"
                width={50}
                height={50}
                className="rounded-full border-white"
              />

              <div className="mb-4">
                <h1 className="text-3xl font-black text-black">{fullName}</h1>
                <span className="text-gray-400 text-sm">{usernameText}</span>
              </div>
            </div>
          </div>

          <div className="flex justify-between mx-10 gap-6 text-sm text-gray-400">
            <span className="flex items-center gap-2">
              <Image src="/location_icon.svg" alt="Location" width={15} height={15} />
              {locationText || "-"}
            </span>
            <span className="flex items-center gap-2 p-1">
              <Image src="/calendar_icon.svg" alt="Birthday" width={15} height={15} />
              {birthdayText || "-"}
            </span>
            <span className="flex items-center gap-2 p-1">
              <Image src="/profile_status_icon.svg" alt="Profile visibility" width={15} height={15} />
              {privacyText}
            </span>
            <Link href="/settings" className="flex items-center gap-2 border rounded-lg px-2 text-sm bg-blue-500 text-white cursor-pointer">
              <Image src="/edit_profile_icon.svg" alt="Edit Profile" width={15} height={15} />
              Edit Profile
            </Link>
          </div>
        </section>

        <section className="flex justify-start gap-8 ml-5">
          <div className="flex flex-col items-center">
            <h1 className="text-4xl text-black">{userPosts.length}</h1>
            <span className="text-gray-400">Posts</span>
          </div>
          {canShowFollowLists ? (
            <>
              <div className="flex flex-col items-center">
                <h1 className="text-4xl text-black">{followers.length}</h1>
                <span className="text-gray-400">Followers</span>
              </div>
              <div className="flex flex-col items-center">
                <h1 className="text-4xl text-black">{following.length}</h1>
                <span className="text-gray-400">Following</span>
              </div>
            </>
          ) : null}
        </section>

        <section className="text-gray-400 flex justify-around border-t border-gray-200 mt-4 pt-2 pb-2">
          <button type="button" onClick={() => setActiveTab("posts")} className="text-gray-400">
            Posts({userPosts.length})
          </button>
          <button type="button" onClick={() => setActiveTab("about")} className="text-gray-400">
            About
          </button>
          {canShowFollowLists ? (
            <>
              <button type="button" onClick={() => setActiveTab("followers")} className="text-gray-400">
                Followers({followers.length})
              </button>
              <button type="button" onClick={() => setActiveTab("following")} className="text-gray-400">
                Following({following.length})
              </button>
            </>
          ) : null}
        </section>
      </main>

      {isLoading ? (
        <article className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
          Loading profile...
        </article>
      ) : error ? (
        <article className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
          {error}
        </article>
      ) : null}

      {!isLoading && !error && activeTab === "posts" ? (
        userPosts.length === 0 ? (
          <article className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
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
            return (
              <article key={post.id} className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
                <div className="flex items-center gap-2">
                  <Image
                    src={parseProfileImage(profileData.profile_picture)}
                    alt="Profile Icon"
                    width={30}
                    height={30}
                  />
                  <h1 className="font-bold text-lg">{fullName}</h1>
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
                <div className="flex justify-end gap-4 mt-2 border-b border-gray-200 pb-1">
                  <span className="text-gray-500 text-sm mr-auto">0 Ripples</span>
                  <span className="text-gray-500 text-sm">{comments.length} Echoes</span>
                </div>
                <div className="flex justify-between gap-8 mt-2 mx-8">
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
                      src={parseProfileImage(profileData.profile_picture)}
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
        )
      ) : null}

      {!isLoading && !error && activeTab === "about" ? (
        <article className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
          <h1 className="font-bold text-lg mb-2">About {profileData.first_name || ""}</h1>
          <p className="mb-3">{aboutText || "No about info yet."}</p>
          <ul className="text-sm text-gray-600 grid grid-cols-2 gap-2">
            <li><strong>Email:</strong> {emailText || "-"}</li>
            <li><strong>Nickname:</strong> {usernameText || "-"}</li>
            <li><strong>Birthday:</strong> {birthdayText || "-"}</li>
            <li><strong>Relationship:</strong> {relationshipText || "-"}</li>
            <li><strong>Employed At:</strong> {employedAtText || "-"}</li>
            <li><strong>Location:</strong> {locationText || "-"}</li>
            <li><strong>Phone:</strong> {phoneText || "-"}</li>
          </ul>
        </article>
      ) : null}

      {!isLoading && !error && canShowFollowLists && activeTab === "followers" ? (
        <article className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
          <h1 className="font-bold text-lg mb-2">Followers</h1>
          {followers.length === 0 ? (
            <p>No followers yet.</p>
          ) : (
            <ul className="flex flex-col gap-2">
              {followers.map((follower) => (
                <li key={follower.id} className="flex items-center gap-2">
                  <Image
                    src={parseProfileImage(follower.profile_picture)}
                    alt="Follower"
                    width={24}
                    height={24}
                  />
                  <span>{`${follower.first_name || ""} ${follower.last_name || ""}`.trim() || "Unknown User"}</span>
                </li>
              ))}
            </ul>
          )}
        </article>
      ) : null}

      {!isLoading && !error && canShowFollowLists && activeTab === "following" ? (
        <article className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
          <h1 className="font-bold text-lg mb-2">Following</h1>
          {following.length === 0 ? (
            <p>Not following anyone yet.</p>
          ) : (
            <ul className="flex flex-col gap-2">
              {following.map((followedUser) => (
                <li key={followedUser.id} className="flex items-center gap-2">
                  <Image
                    src={parseProfileImage(followedUser.profile_picture)}
                    alt="Following"
                    width={24}
                    height={24}
                  />
                  <span>{`${followedUser.first_name || ""} ${followedUser.last_name || ""}`.trim() || "Unknown User"}</span>
                </li>
              ))}
            </ul>
          )}
        </article>
      ) : null}
    </div>
  );
};

export default Profile;
