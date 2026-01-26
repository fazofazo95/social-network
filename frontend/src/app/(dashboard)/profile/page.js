"use client";

import Image from "next/image";
import {useState} from "react";

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

    const handleStatusChange = () => {
        if (status === "Public") {
            console.log("Changing status to Private");
            setStatus("Private");
        } else {
            setStatus("Public");
        }
    };
  return (
    <>
      <main className="flex flex-col w-full min-w-xl max-w-2xl bg-white rounded-lg overflow-hidden gap-1">
        <div className="bg-[url('/example_cover.png')] w-full bg-size-[100%_100%] h-36 relative">
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
      </main>
    </>
  );
};

export default Profile;
