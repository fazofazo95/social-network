"use client";
import { useState } from "react";

const Follow_Bottom = ({status : initialStatus}) => {
      const [status, setStatus] = useState(initialStatus);
    
      const handleFollow = () => {
        const isPrivate = true; // Simulate checking if the profile is private
        if (isPrivate) {
          setStatus("pending");
        } else {
          setStatus("following");
        }
      };
    return (  <button
                type="button"
                className="text-pink-500 ml-auto cursor-pointer hover:text-pink-400"
                onClick={handleFollow}
              >
                {status === "not_following" && "Follow"}
                {status === "pending" && "Pending"}
                {status === "following" && "Unfollow"}
              </button> );
}
 
export default Follow_Bottom;