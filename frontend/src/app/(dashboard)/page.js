"use client";

import FormContainer from "../../components/ui/FormContainer";
import SideBar from "../../components/ui/SideBar";
import SuggestedFriends from "../../components/ui/Suggested_Friends";
import Input from "../../components/ui/Input";
import Image from "next/image";

//import LoginForm from "@/components/LoginForm";

// import { useState, useEffect } from "react";

export default function App() {
  // const [session, setSession] = useState(null);

  // useEffect(() => {
  //   // Η απο το session storage ή local storage
  //  const userSession = localStorage.getItem("session");
  //  setSession(userSession ? JSON.parse(userSession) : null);
  // }, []);

  // return session ? ( <p></p> ) : ( <LoginForm /> );
  return (
    <main className="w-full max-w-2xl flex flex-col gap-20">
      <form
        encType="multipart/form-data"
        className="bg-white w-full rounded-lg shadow-custom p-4 sticky top-16 z-10"
      >
        <div className="flex items-start gap-4 mb-4">
          <Image
            src="/profil_icon.svg"
            alt="Profile Icon"
            width={25}
            height={25}
          />
          <textarea
            className="border rounded border-gray-200 text-black w-full h-20 focus:outline-none pl-2 resize-none"
            placeholder="What's on your mind?"
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
                className="font-medium cursor-pointer text-black hidden"
              />
              <span
                className="font-medium cursor-pointer text-black"
              >
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
            <button type="button" className="text-white cursor-pointer">
              Post
            </button>
          </li>
        </ul>
      </form>
      
      <article className="border border-gray-200 rounded-lg bg-white text-black  w-full p-5">
        <h1>User 1</h1>
        <p>This is a sample post content.</p>
      </article>
    </main>
  );
}
