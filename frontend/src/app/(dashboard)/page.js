"use client";

import FormContainer from "../../components/ui/FormContainer";
import SideBar from "../../components/ui/SideBar";
import SuggestedFriends from "../../components/ui/Suggested_Friends";
import Input from "../../components/ui/Input";
import Image from "next/image";
import Ripple_Button from "src/components/ui/Ripple_Button";

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
  const handleEchoBtn = () => {
    const echoSection = document.getElementById("echo-section");
    if (echoSection.classList.contains("hidden")) {
      echoSection.classList.remove("hidden");
      echoSection.classList.add("flex");
    } else {
      echoSection.classList.add("hidden");
      echoSection.classList.remove("flex");
    }
  };

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
            <button type="button" className="text-white cursor-pointer">
              Post
            </button>
          </li>
        </ul>
      </form>

      <article className="border border-gray-200 rounded-lg bg-white text-black  w-full p-5">
        <div className="flex items-center gap-2">
          <Image
            src="/profil_icon.svg"
            alt="Profile Icon"
            width={30}
            height={30}
          />
          <h1 className="font-bold text-lg">User 1</h1>
        </div>
        <span className="text-sm text-gray-500 ml-4 mb-2">Just now</span>
        <p>This is a sample post content.</p>
        <div className="flex justify-end gap-4 mt-2 border-b border-gray-200 pb-1">
          <span className="text-gray-500 text-sm mr-auto">10 Ripples</span>
          <span className="text-gray-500 text-sm">2 Echoes</span>
          <span className="text-gray-500 text-sm">1 Spreads</span>
        </div>
        <div className="flex justify-between gap-8 mt-2 mx-8">
          <Ripple_Button />
          <button onClick={handleEchoBtn} className="flex cursor-pointer gap-1">
            <Image
              src="/echo_icon.svg"
              alt="Echo Icon"
              width={20}
              height={20}
            />
            Echo
          </button>
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
          id="echo-section"
          className="border-t border-gray-200 rounded mt-2 pt-2 gap-1 hidden"
        >
          <Image
            src="/profil_icon.svg"
            alt="Profile Icon"
            width={25}
            height={25}
          />
          <div className=" flex justify-between bg-gray-100 text-black w-full rounded-lg  resize-none h-10">
            <input
              type="text"
              placeholder="Write a comment..."
              className="focus:outline-none w-full pl-1"
            />

            <label
              htmlFor="photo-upload"
              className="flex items-center gap-1 cursor-pointer"
            >
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
            </label>
          </div>
        </div>
      </article>
    </main>
  );
}
