"use client";

import Image from "next/image";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Follow_Bottom from "src/components/ui/Follow_Button";
import { fetchUserData } from "src/lib/services/user";
import { getUserPosts } from "src/lib/services/post";
import { blockUser, getFollowersByUser, getFollowingByUser, unblockUser } from "src/lib/services/follow";
import Avatar from "src/components/ui/Avatar";
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

  if (isLoading) {
    return (
      <div className="w-full max-w-2xl flex flex-col items-center justify-center py-20">
        <div className="text-purple-400 text-lg">Loading profile...</div>
      </div>
    );
  }

  if (error || (!isLoading && !profileData?.id)) {
    return (
      <div className="w-full max-w-2xl flex flex-col items-center justify-center py-20 gap-4">
        <div className="bg-[#1a1a2e] border border-purple-500/30 rounded-xl p-10 flex flex-col items-center gap-4 shadow-[0_0_15px_rgba(168,85,247,0.15)]">
          <span className="text-5xl">👤</span>
          <h2 className="text-2xl font-bold text-purple-200">User Not Found</h2>
          <p className="text-purple-400 text-center max-w-sm">
            {error || "The profile you're looking for doesn't exist or may have been removed."}
          </p>
          <Link
            href="/"
            className="mt-2 px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white font-semibold rounded-lg transition shadow-custom"
          >
            Back to Home
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full max-w-2xl flex flex-col gap-10 pb-8">
      <main className="flex flex-col w-full bg-[#1a1a2e] rounded-lg overflow-hidden gap-2 border border-purple-500/30 shadow-[0_0_15px_rgba(168,85,247,0.15)]">
        <div
          className="w-full h-36 relative"
          style={{
            backgroundImage: `url('${toCoverUrl(profileData.cover_image)}')`,
            backgroundSize: "cover",
            backgroundPosition: "center",
            backgroundRepeat: "no-repeat",
          }}
        />

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

          <div className="flex justify-between items-center mx-10 gap-6 text-sm text-purple-400">
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
            {profileData.own_profile ? (
              <Link href="/settings" className="flex items-center gap-2 border rounded-lg px-2 text-sm bg-blue-500 text-white cursor-pointer">
                <span className="text-sm">✏️</span>
                Edit Profile
              </Link>
            ) : isBlockedByTarget ? (
              <span className="text-sm text-red-400">You are blocked</span>
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
          {blockActionError ? <p className="text-red-400 text-sm mx-10 mt-2">{blockActionError}</p> : null}
        </section>

        <section className="flex justify-start gap-8 ml-5">
          <div className="flex flex-col items-center">
            <h1 className="text-4xl text-purple-100">{userPosts.length}</h1>
            <span className="text-purple-400">Posts</span>
          </div>
          {canShowFollowLists ? (
            <>
              <div className="flex flex-col items-center">
                <h1 className="text-4xl text-purple-100">{followersCount}</h1>
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
          <button type="button" onClick={() => setActiveTab("posts")} className={`cursor-pointer transition ${activeTab === "posts" ? "text-purple-200 font-semibold" : "text-purple-400 hover:text-purple-300"}`}>
            Posts({userPosts.length})
          </button>
          <button type="button" onClick={() => setActiveTab("about")} className={`cursor-pointer transition ${activeTab === "about" ? "text-purple-200 font-semibold" : "text-purple-400 hover:text-purple-300"}`}>
            About
          </button>
          {canShowFollowLists ? (
            <>
              <button type="button" onClick={() => setActiveTab("followers")} className={`cursor-pointer transition ${activeTab === "followers" ? "text-purple-200 font-semibold" : "text-purple-400 hover:text-purple-300"}`}>
                Followers({followersCount})
              </button>
              <button type="button" onClick={() => setActiveTab("following")} className={`cursor-pointer transition ${activeTab === "following" ? "text-purple-200 font-semibold" : "text-purple-400 hover:text-purple-300"}`}>
                Following({following.length})
              </button>
            </>
          ) : null}
        </section>
      </main>

      {activeTab === "posts" ? (
        userPosts.length === 0 ? (
          <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-300 w-full p-5">
            No posts yet.
          </article>
        ) : (
          userPosts.map((post) => {
            const postDateLabel = formatFriendlyDateTime(post.created_at_time || post.created_at);
            return (
              <article key={post.id} className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-100 w-full p-5 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition">
                <div className="flex items-center gap-2">
                  <Avatar
                    src={post.author_profile_picture || profileData.profile_picture}
                    name={fullName}
                    size={30}
                  />
                  <Link
                    href={post.user_id ? `/profile/${post.user_id}` : (targetUserId ? `/profile/${targetUserId}` : "/profile")}
                    className="font-bold text-lg text-purple-200 hover:text-purple-100"
                  >
                    {fullName}
                  </Link>
                </div>
                {postDateLabel ? <span className="text-sm text-purple-400 ml-4 mb-2">{postDateLabel}</span> : null}
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
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl mb-1 text-purple-100">User Information</h1>
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
          {followers.length === 0 ? (
            <p className="text-sm text-purple-300">No followers yet.</p>
          ) : (
            <ul className="flex flex-col gap-2.5">
              {followers.map((follower) => (
                <li key={follower.id} className="flex items-center gap-3 rounded-md border border-purple-500/20 bg-[#0d0d1a] px-3 py-2 hover:bg-purple-900/20 transition">
                  <Avatar
                    src={follower.profile_picture}
                    name={`${follower.first_name || ""} ${follower.last_name || ""}`.trim()}
                    size={24}
                    className="h-6 w-6"
                  />
                  <span className="flex-1 min-w-0">
                    <span className="block truncate text-sm font-semibold text-purple-200">{`${follower.first_name || ""} ${follower.last_name || ""}`.trim() || "Unknown User"}</span>
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
                  <span className="text-xs bg-purple-900/30 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1">Follower</span>
                </li>
              ))}
            </ul>
          )}
        </article>
      ) : null}

      {!isLoading && !error && canShowFollowLists && activeTab === "following" ? (
        <article className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-white w-full p-5">
          <h1 className="font-bold text-2xl text-purple-200 mb-3">Following ({following.length})</h1>
          {following.length === 0 ? (
            <p className="text-sm text-purple-300">Not following anyone yet.</p>
          ) : (
            <ul className="flex flex-col gap-2.5">
              {following.map((followedUser) => (
                <li key={followedUser.id} className="flex items-center gap-3 rounded-md border border-purple-500/20 bg-[#0d0d1a] px-3 py-2 hover:bg-purple-900/20 transition">
                  <Avatar
                    src={followedUser.profile_picture}
                    name={`${followedUser.first_name || ""} ${followedUser.last_name || ""}`.trim()}
                    size={24}
                    className="h-6 w-6"
                  />
                  <span className="flex-1 min-w-0">
                    <span className="block truncate text-sm font-semibold text-purple-200">{`${followedUser.first_name || ""} ${followedUser.last_name || ""}`.trim() || "Unknown User"}</span>
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
                  <span className="text-xs bg-purple-900/30 text-purple-300 border border-purple-500/30 rounded-md px-3 py-1">Following</span>
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
