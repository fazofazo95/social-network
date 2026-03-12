"use client";
import Image from "next/image";
import Link from "next/link";
import Follow_Bottom from "./Follow_Button";
import { useEffect, useState } from "react";
import { getDiscoveredUsers } from "src/lib/services/discover";
import { parseProfileImage } from "src/lib/utils/profileImage";

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
    <section className="bg-[#1a1a2e] border border-purple-500/30 rounded-xl px-1 shadow-custom">
      <h1 className="custom-pink-text pt-2 pb-1 px-3 text-sm font-semibold">Suggested Friends</h1>
      <ul className="px-3 py-2 flex flex-col gap-3">
          {isLoading ? (
            <p className="text-purple-400 text-sm">Loading...</p>
          ) : (
            suggestedFriends.map(friend => (
              <li key={friend.id} className="flex items-center gap-2 hover:bg-purple-900/20 rounded-lg px-1 py-1 transition">
                <Image
                  src={parseProfileImage(friend.profile_picture)}
                  alt={`${friend.first_name} ${friend.last_name}'s Profile Picture`}
                  width={20}
                  height={20}
                  className="rounded-full"
                />
                <Link href={`/profile/${friend.id}`} className="text-purple-200 text-sm hover:text-purple-100 transition truncate">
                  {friend.first_name} {friend.last_name}
                </Link>
                <Follow_Bottom status={friend.status} targetUserId={friend.id} />
              </li>
            ))
          )}
      </ul>
    </section>
  );
};

export default SuggestedFriends;
