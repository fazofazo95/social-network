"use client";
import Image from "next/image";
import Follow_Bottom from "./Follow_Button";
import { useEffect, useState } from "react";
import { getApiBaseUrl } from "src/lib/apiClient";
import { getDiscoveredUsers } from "src/lib/services/discover";

const FALLBACK_PROFILE_IMAGE = "/profil2_icon.svg";
const BACKEND_BASE_URL = getApiBaseUrl();

function parseProfileImage(profilePicture) {
  if (!profilePicture || typeof profilePicture !== "string") {
    return FALLBACK_PROFILE_IMAGE;
  }

  const trimmed = profilePicture.trim();
  if (!trimmed) {
    return FALLBACK_PROFILE_IMAGE;
  }

  if (trimmed.startsWith("http://") || trimmed.startsWith("https://") || trimmed.startsWith("data:")) {
    return trimmed;
  }

  if (trimmed.startsWith("/uploads/")) {
    return `${BACKEND_BASE_URL}${trimmed}`;
  }

  return FALLBACK_PROFILE_IMAGE;
}

const SuggestedFriends = () => {
  const [suggestedFriends, setSuggestedFriends] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  async function getSuggestedFriends() {
    try {
      const users = await getDiscoveredUsers();
      setSuggestedFriends(users);
    } catch (error) {
      console.error('Error fetching suggested friends:', error);
      setSuggestedFriends([]);
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    getSuggestedFriends();
  }, []);

  return (
    <section className="border rounded-xl w-64 border-purple-500 px-1  shadow-custom">
      <h1 className="custom-pink-text pt-2 pb-1">Suggested Friends</h1>
      <ul className="px-3 py-2 flex flex-col gap-3">
          {isLoading ? (
            <p>Loading...</p>
          ) : (
            suggestedFriends.map(friend => (
              <li key={friend.id} className="flex gap-1">
                <Image
                  src={parseProfileImage(friend.profile_picture)}
                  alt={`${friend.first_name} ${friend.last_name}'s Profile Picture`}
                  width={20}
                  height={20}
                  unoptimized
                />
                <span>{friend.first_name} {friend.last_name}</span>
                <Follow_Bottom status={friend.status} targetUserId={friend.id} />
              </li>
            ))
          )}
      </ul>
    </section>
  );
};

export default SuggestedFriends;
