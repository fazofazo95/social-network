"use client";

import Image from "next/image";
import {useState} from "react";
import Echo_Button from "src/components/ui/Echo_Button";
import Ripple_Button from "src/components/ui/Ripple_Button";
 let currentImage = "/example_cover.png";
const Profile = () => {
    const Profile = {
        firstname: "John",
        lastname: "Doe",
        username: "johndoe",
        location: "New York, USA",
        joinDate: "March 2023",
        status: "Public",
    };
    const UserPosts = [1, 2, 3, 4, 5];
    const Followers = [1, 2, 3, 4, 5, 6, 7, 8];
    const Following = [1, 2, 3, 4];

    const [status, setStatus] = useState(Profile.status);

   
    const [coverImage, setCoverImage] = useState(currentImage);
    const handleChangeCover = (event) => {
        const file = event.target.files[0];
        console.log(file);
        if (file) {
            const reader = new FileReader();
            console.log(reader);
            reader.onload = () => {
                setCoverImage(reader.result);
            };
            currentImage = reader.result;
            reader.readAsDataURL(file);
        }
    };
    const handleStatusChange = () => {
        if (status === "Public") {
            console.log("Changing status to Private");
            setStatus("Private");
        } else {
            setStatus("Public");
        }
    };

    const Toggle = (event) => {
        const
         btnId = event.target.id;
          const userPostsSection = document.getElementById("userPosts");
          const aboutSection = document.getElementById("aboutSection");
          const followersSection = document.getElementById("followersSection");
          const followingSection = document.getElementById("followingSection");
        if (btnId === "userPostsBtn") {
            userPostsSection.classList.remove("hidden");
            aboutSection.classList.add("hidden");
            followersSection.classList.add("hidden");
            followingSection.classList.add("hidden");
        } else if (btnId === "aboutBtn") {
            aboutSection.classList.remove("hidden");
            userPostsSection.classList.add("hidden");
            followersSection.classList.add("hidden");
            followingSection.classList.add("hidden");
        } else if (btnId === "followersBtn") {
            followersSection.classList.remove("hidden");
            userPostsSection.classList.add("hidden");
            aboutSection.classList.add("hidden");
            userPostsSection.classList.add("hidden");
        } else if (btnId === "followingBtn") {
            followingSection.classList.remove("hidden");
            aboutSection.classList.add("hidden");
            userPostsSection.classList.add("hidden");
            followersSection.classList.add("hidden");
        }
    };

      

  return (
<div className="flex flex-col gap-10">
      <main className="flex flex-col w-full min-w-xl max-w-2xl bg-white rounded-lg overflow-hidden gap-2">
        <div 
          className="w-full h-36 relative"
          style={{
            backgroundImage: `url('${coverImage}')`,
            backgroundSize: '100% 100%'
          }}
        >
          <label
            htmlFor="cover-upload"
            className="flex items-center gap-1 cursor-pointer absolute bottom-2 right-2 bg-gray-200 bg-opacity-70 p-1 rounded"
          >
            <Image
              src="/cover_icon.svg"
              alt="Share Icon"
              width={20}
              height={20}
            />
            <span className="text-sm text-black">Change Cover</span>
            <input
              id="cover-upload"
              type="file"
              accept="image/*"
              onChange={handleChangeCover}
              className="font-medium cursor-pointer text-black hidden"
            />
          </label>
        </div>

        <section className="border-b border-gray-200 pb-4 mb-2">
            <div className="flex items-center gap-2 justify-start">
          <div className="flex items-center gap-2 pl-5 pt-5">
            <Image
              src="/profil_icon.svg"
              alt="Profile Picture"
              width={50}
              height={50}
              className="rounded-full border-white"
            />

            <div className="mb-4 ">
              <h1 className="text-3xl font-black text-black">{Profile.firstname} {Profile.lastname}</h1>
              <span className="text-gray-400 text-sm">{Profile.username}</span>
            </div>
          </div>
          </div>

            <div className="flex justify-between mx-10 gap-6 text-sm text-gray-400">
              <span className="flex items-center gap-2">
                <Image src="/location_icon.svg" alt="Location Icon" width={15} height={15} />
                {Profile.location}</span>
              <span className="flex items-center gap-2 p-1">
                <Image src="/calendar_icon.svg" alt="Calendar Icon" width={15} height={15} />
                Joined {Profile.joinDate}</span>
              <span className="flex items-center gap-2 p-1">
                <Image src="/profile_status_icon.svg" alt="Public Icon" width={15} height={15} />
                {status}
                </span>
                
              <button onClick={handleStatusChange} className=" flex items-center gap-2 bg-white border rounded-lg px-2 text-sm border-blue-500 text-blue-500 cursor-pointer">  <Image src="/profile_status2_icon.svg" alt="Public Icon" width={15} height={15} />Make {status === "Public" ? "Private" : "Public"}</button>
              <button className=" flex items-center gap-2 border rounded-lg px-2 text-sm bg-blue-500 text-white cursor-pointer  ">  <Image src="/edit_profile_icon.svg" alt="Public Icon" width={15} height={15} />Edit Profile</button>
            </div>
        </section>
        <section className="flex justify-start gap-8 ml-5">
            <div className="flex flex-col items-center">
            <h1 className="text-4xl text-black">{UserPosts.length}</h1>
            <span className="text-gray-400">Posts</span>  
            </div>
            <div className="flex flex-col items-center">
            <h1 className="text-4xl text-black">{Followers.length}</h1>
            <span className="text-gray-400">Followers</span>  
            </div>
            <div className="flex flex-col items-center">
            <h1 className="text-4xl text-black">{Following.length}</h1>
            <span className="text-gray-400">Following</span>  
            </div>
        </section>
        <section className="text-gray-400 flex justify-around border-t border-gray-200 mt-4 pt-2 pb-2">
            <button id="userPostsBtn" onClick={Toggle} className="text-gray-400">Posts({UserPosts.length})</button>
            <button id="aboutBtn" onClick={Toggle} className="text-gray-400">About</button>
            <button id="followersBtn" onClick={Toggle} className="text-gray-400">Followers({Followers.length})</button>
            <button id="followingBtn" onClick={Toggle} className="text-gray-400">Following({Following.length})</button>
        </section>

      </main>
      
        <article id="userPosts" className="hidden border border-gray-200 rounded-lg bg-white text-black  w-full p-5">
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
         <Echo_Button />
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
        <article id="aboutSection" className="hidden border border-gray-200 rounded-lg bg-white text-black  w-full p-5">
        <h1 className="font-bold text-lg mb-2">About {Profile.firstname}</h1>
        <p>This is the about section.</p>
        </article>
        <article id="followersSection" className="hidden border border-gray-200 rounded-lg bg-white text-black  w-full p-5">
        <h1 className="font-bold text-lg mb-2">Followers</h1>
        <ul>
            {Followers.map((follower, index) => (
                <li key={index} className="mb-1">Follower {follower}</li>
            ))}
        </ul>
        </article>
        <article id="followingSection" className="hidden border border-gray-200 rounded-lg bg-white text-black  w-full p-5">
        <h1 className="font-bold text-lg mb-2">Following</h1>
        <ul>
            {Following.map((followed, index) => (
                <li key={index} className="mb-1">Following {followed}</li>
            ))}
        </ul>
        </article>
    </div>

  );
};

export default Profile;
