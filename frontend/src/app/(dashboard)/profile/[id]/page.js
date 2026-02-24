"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Follow_Bottom from "src/components/ui/Follow_Button";
import { fetchUserData } from "src/lib/services/user";
import { getUserPosts } from "src/lib/services/post";
import { blockUser, getFollowersByUser, getFollowingByUser, unblockUser } from "src/lib/services/follow";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";
import { getApiBaseUrl } from "src/lib/apiClient";

const UserProfilePage = () => {
  const params = useParams();
  const targetUserId = Array.isArray(params?.id) ? params.id[0] : params?.id;

  const [activeTab, setActiveTab] = useState("posts");
  const [profileData, setProfileData] = useState({});
  const [userPosts, setUserPosts] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [followersCount, setFollowersCount] = useState(0);
  const [following, setFollowing] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [isBlockActionLoading, setIsBlockActionLoading] = useState(false);
  const [blockActionError, setBlockActionError] = useState("");

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
    if (!targetUserId) {
      setError("Invalid profile id.");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError("");

    try {
      const [profile, postsData, followersData, followingData] = await Promise.all([
        fetchUserData(targetUserId),
        getUserPosts(targetUserId, 1, 10),
        getFollowersByUser(targetUserId),
        getFollowingByUser(targetUserId),
      ]);

      setProfileData(profile || {});
      setUserPosts(Array.isArray(postsData) ? postsData : []);
      setFollowers(Array.isArray(followersData) ? followersData : []);
      setFollowersCount(
        typeof profile?.followers === "number" ? profile.followers : (Array.isArray(followersData) ? followersData.length : 0)
      );
      setFollowing(Array.isArray(followingData) ? followingData : []);
    } catch (loadError) {
      console.error("Error loading target profile page:", loadError);
      setProfileData({});
      setUserPosts([]);
      setFollowers([]);
      setFollowersCount(0);
      setFollowing([]);
      setError(loadError?.message || "Failed to load profile.");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    loadProfilePageData();
  }, [targetUserId]);

  const fullName = `${profileData.first_name || ""} ${profileData.last_name || ""}`.trim() || "Unknown User";
  const usernameText = profileData.nickname ? `@${profileData.nickname}` : "";
  const relationshipText = profileData.relationship_status || "";
  const locationText = profileData.location || "";
  const employedAtText = profileData.employed_at || "";
  const phoneText = profileData.phone_number || "";
  const emailText = profileData.email || "";
  const aboutText = profileData.about_me || "";
  const birthdayText = profileData.birthday_date || "";
  const privacyText = String(profileData.profile_type || "public").toLowerCase() === "private" ? "Private" : "Public";
  const canShowFollowLists = profileData.own_profile || profileData.follow_vis !== "hidden";
  const currentStatus = String(profileData.current_status || "");
  const isBlockedByMe = currentStatus === "Blocked";
  const isBlockedByTarget = currentStatus === "You_Are_Blocked";

  async function handleToggleBlock() {
    if (!profileData?.id || profileData?.own_profile || isBlockedByTarget) {
      return;
    }

    setBlockActionError("");
    setIsBlockActionLoading(true);
    try {
      if (isBlockedByMe) {
        await unblockUser(profileData.id);
      } else {
        await blockUser(profileData.id);
      }
      await loadProfilePageData();
    } catch (actionError) {
      console.error("Failed to change block status:", actionError);
      setBlockActionError(actionError?.message || "Failed to update block status.");
    } finally {
      setIsBlockActionLoading(false);
    }
  }

  function handleFollowStatusChange(nextStatus) {
    const previousStatus = profileData.current_status;

    setProfileData((prev) => ({
      ...prev,
      current_status: nextStatus,
    }));

    if (nextStatus === "Following" && previousStatus !== "Following") {
      setFollowersCount((prev) => prev + 1);
    } else if (nextStatus === "Follow" && previousStatus === "Following") {
      setFollowersCount((prev) => Math.max(0, prev - 1));
    }
  }

  return (
    <div className="w-full max-w-2xl flex flex-col gap-10">
      <main className="flex flex-col w-full bg-white rounded-lg overflow-hidden gap-2">
        <div
          className="w-full h-36 relative"
          style={{
            backgroundImage: `url('${toCoverUrl(profileData.cover_image)}')`,
            backgroundSize: "cover",
            backgroundPosition: "center",
            backgroundRepeat: "no-repeat",
          }}
        />

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

          <div className="flex justify-between items-center mx-10 gap-6 text-sm text-gray-400">
            <div className="flex flex-wrap items-center gap-6">
              {locationText ? (
                <span className="flex items-center gap-2">
                  <Image src="/location_icon.svg" alt="Location" width={15} height={15} />
                  {locationText}
                </span>
              ) : null}
              {birthdayText ? (
                <span className="flex items-center gap-2 p-1">
                  <Image src="/calendar_icon.svg" alt="Birthday" width={15} height={15} />
                  {birthdayText}
                </span>
              ) : null}
              <span className="flex items-center gap-2 p-1">
                <Image src="/profile_status_icon.svg" alt="Profile visibility" width={15} height={15} />
                {privacyText}
              </span>
            </div>
            {profileData.own_profile ? (
              <Link href="/settings" className="flex items-center gap-2 border rounded-lg px-2 text-sm bg-blue-500 text-white cursor-pointer">
                <Image src="/edit_profile_icon.svg" alt="Edit Profile" width={15} height={15} />
                Edit Profile
              </Link>
            ) : isBlockedByTarget ? (
              <span className="text-sm text-red-500">You are blocked</span>
            ) : (
              <div className="flex items-center gap-2 ml-auto">
                {!isBlockedByMe ? (
                  <Follow_Bottom status={profileData.current_status} targetUserId={profileData.id} onStatusChange={handleFollowStatusChange} />
                ) : null}
                <button
                  type="button"
                  className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
                  onClick={handleToggleBlock}
                  disabled={isBlockActionLoading}
                >
                  {isBlockActionLoading ? "Working..." : isBlockedByMe ? "Unblock" : "Block"}
                </button>
              </div>
            )}
          </div>
          {blockActionError ? <p className="text-red-600 text-sm mx-10 mt-2">{blockActionError}</p> : null}
        </section>

        <section className="flex justify-start gap-8 ml-5">
          <div className="flex flex-col items-center">
            <h1 className="text-4xl text-black">{userPosts.length}</h1>
            <span className="text-gray-400">Posts</span>
          </div>
          {canShowFollowLists ? (
            <>
              <div className="flex flex-col items-center">
                <h1 className="text-4xl text-black">{followersCount}</h1>
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
                Followers({followersCount})
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
            const postDateLabel = formatFriendlyDateTime(post.created_at_time || post.created_at);
            return (
              <article key={post.id} className="border border-gray-200 rounded-lg bg-white text-black w-full p-5">
                <div className="flex items-center gap-2">
                  <Image
                    src={parseProfileImage(post.author_profile_picture || profileData.profile_picture)}
                    alt="Profile Icon"
                    width={30}
                    height={30}
                  />
                  <Link
                    href={post.user_id ? `/profile/${post.user_id}` : (targetUserId ? `/profile/${targetUserId}` : "/profile")}
                    className="font-bold text-lg"
                  >
                    {fullName}
                  </Link>
                </div>
                {postDateLabel ? <span className="text-sm text-gray-500 ml-4 mb-2">{postDateLabel}</span> : null}
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

export default UserProfilePage;
