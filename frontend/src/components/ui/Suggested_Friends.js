"use client";
import Image from "next/image";
import Follow_Bottom from "../Follow_Button";
import { useEffect, useState } from "react";
//fetch('/api/discover', { method: 'GET', credentials: 'include' })
//{ "id": number, "first_name": string, "last_name": string, "profile_picture": string, "status": string }

const FALLBACK_PROFILE_IMAGE = "/profil2_icon.svg";
const BACKEND_BASE_URL = "http://localhost:8080";

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
      const response = await fetch('http://localhost:8080/api/discover', { method: 'GET', credentials: 'include' });
      if (!response.ok) {
        console.error('Failed to fetch suggested friends:', response.statusText);
      }
      const resp = await response.json();
      setSuggestedFriends(Array.isArray(resp?.data) ? resp.data : []);
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
                <Follow_Bottom status={friend.status} />
              </li>
            ))
          )}
      </ul>
    </section>
  );
};

export default SuggestedFriends;
