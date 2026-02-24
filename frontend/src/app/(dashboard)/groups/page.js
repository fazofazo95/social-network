"use client";

import SearchBar from "src/components/ui/SearchBar";
import Image from "next/image";
import { useState, useEffect } from "react";
import { getDiscoveredGroups } from "src/lib/services/discoverGroups";
import { parseProfileImage } from "src/lib/utils/profileImage";

const GroupsPage = () => {
    const currentUser = "User 1";
    const [activeSection, setActiveSection] = useState("my-groups");
    const user = {
        Groups: [
            { id: 1, name: "this is the title of Group 1", owner: "User 1", content: "this is the content of Group 1" , members: ["User 1", "User 3", "User 4"]},
            { id: 2, name: "this is the title of Group 2", owner: "User 2", content: "this is the content of Group 2" , members: ["User 2", "User 5"]},
        ],
        invitations: [
            { id: 1, from: "User A" , status: "pending"},
            { id: 2, from: "User B" , status: "pending"},
        ],
    };

    const [discoveredGroups, setDiscoveredGroups] = useState([]);
    const [groupsLoading, setGroupsLoading] = useState(false);
    const [groupsError, setGroupsError] = useState(null);

    useEffect(() => {
        if (activeSection === "discover") {
            setGroupsLoading(true);
            setGroupsError(null);
            getDiscoveredGroups()
                .then(setDiscoveredGroups)
                .catch((err) => setGroupsError(err.message || "Failed to load groups"))
                .finally(() => setGroupsLoading(false));
        }
    }, [activeSection]);

    return (
        <main className="flex flex-col items-left w-full max-w-2xl gap-6 p-4">
        <header className="flex flex-row justify-between items-center w-full mb-4">  
        <h1 className="text-5xl font-bold p-4">Groups</h1>
        <button id="createGroupBtn" onClick={() => setActiveSection("create-group")} className="bg-blue-500 rounded-lg  p-0.5 w-1/5 hover:bg-blue-600 text-white pl-4 relative cursor-pointer">
        <Image src="/group_plus.svg" alt="Create Group Icon" width={20} height={20} className="inline-block -mt-1 mr-2 absolute top-2 left-2.5"/>Create Group</button>
        </header>
        <ul className="flex flex-row justify-end gap-4 border-b border-purple-500 pb-4 w-full">
            <li className="font-bold  hover:border border-purple-500 rounded-lg px-2">
                <button id="myGroupsBtn" onClick={() => setActiveSection("my-groups")} className="cursor-pointer">My Groups ({user.Groups.length})</button>
            </li>
            <li className="font-bold hover:border border-purple-500 rounded-lg px-2">
                <button id="invitationsBtn" onClick={() => setActiveSection("invitations")} className="cursor-pointer relative">Invitations<span id="invitations-notification" className="rounded-full bg-red-600 text-xsm absolute  top-0 -right-2 p-0.5">{user.invitations?.filter(invitation => invitation.status === "pending").length || 0}</span></button>
            </li>
            <li className="font-bold hover:border border-purple-500 rounded-lg  px-2 mr-auto">
                <button id="discoverBtn" onClick={() => setActiveSection("discover")} className="cursor-pointer">Discover</button>
            </li>
            <li>
                  <SearchBar />
            </li>
        </ul>

        <section id="create-group-section" className={activeSection === "create-group" ? "flex flex-col border border-purple-500 rounded-lg p-4" : "hidden"}>
            <p>Create Group</p>
        </section>
        
<section id="my-groups-section" className={activeSection === "my-groups" ? "flex flex-col gap-4" : "hidden"}>
        {user.Groups.map(group => (
        <article key={group.id} className="flex flex-col border  border-purple-500 rounded-lg">
        <header className="flex flex-row items-end p-4">
        <Image
          src="/profil_icon.svg"
          alt="Groups Banner"
            width={50}
            height={50}
            className="bg-purple-500 rounded-full"
          
        />
        <h1 className="font-bold text-2xl">{group.name}</h1>
        <Image
          src="/profile_status_icon.svg"
          alt="Group Owner"
          width={30}
          height={30}
          className=" rounded-full ml-4"
        />
        {group.owner == currentUser && (
        <span className="bg-purple-500 rounded py-1 px-3 ml-2">Creator</span>
        )}

        <button className="ml-auto mr-2 cursor-pointer bg-purple-500 rounded py-1 px-3">View Group</button>
        {group.owner == currentUser && (
        <button className="ml-2 border border-purple-500 rounded cursor-pointer p-1">
            <Image src="/settings_icon.svg" alt="Settings Icon" width={20} height={20} />
        </button>
        )}
        {group.owner != currentUser && (
        <button className="ml-2 border border-purple-400 rounded cursor-pointer px-2">
           Leave
        </button>
        )}
        </header>
        <main className="ml-19 mb-5">
        <p>{group.content}</p>
        </main>
        <footer className="ml-12 mb-4 text-purple-500">
            <span>
                <Image src="/groups_icon.svg" alt="Members Icon" width={15} height={15} className="inline-block mr-1" />
                {group.members.length} members</span>
            <span className="ml-4">Created by {group.owner}</span>
        </footer>
        </article>
        ))}
        </section>
        <section id="discover-section" className={activeSection === "discover" ? "flex flex-col gap-6" : "hidden"}>
            {groupsLoading && <p>Loading groups...</p>}
            {groupsError && <p className="text-red-500">{groupsError}</p>}
            {!groupsLoading && !groupsError && discoveredGroups.length === 0 && (
                <p>No groups to discover.</p>
            )}
            <div className="flex flex-col gap-6">
                {discoveredGroups.map((group) => (
                    <article
                        key={group.id}
                        className="bg-[#1a0033] border border-purple-900 rounded-xl shadow-lg p-5 flex flex-row items-start gap-4 max-w-xl"
                    >
                      
                        <div className="flex items-center justify-center w-14 h-14 rounded-lg bg-purple-700 text-white text-2xl font-bold overflow-hidden">
                            {group.group_picture || group.name ? (
                                <Image
                                    src={parseProfileImage(group.group_picture)}
                                    alt={group.name}
                                    width={56}
                                    height={56}
                                    className="rounded-lg object-cover"
                                    onError={e => { e.target.onerror = null; e.target.src = "/groups_icon.svg"; }}
                                />
                            ) : (
                                <span>{group.name?.[0]?.toUpperCase() || "G"}</span>
                            )}
                        </div>
                      
                        <div className="flex-1 flex flex-col gap-1">
                            <div className="flex items-center gap-2">
                                <span className="font-bold text-xl text-white">{group.name}</span>
                            </div>
                            <p className="text-purple-100 text-sm mb-1">
                                {group.description || "No description provided."}
                            </p>
                            <div className="flex items-center gap-6 text-purple-300 text-sm mt-2">
                                <span className="flex items-center gap-1">
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-4 h-4 inline-block"><path strokeLinecap="round" strokeLinejoin="round" d="M17.982 18.725A7.488 7.488 0 0012 16.5a7.488 7.488 0 00-5.982 2.225M15 9a3 3 0 11-6 0 3 3 0 016 0z" /></svg>
                                    {group.members_count ? group.members_count.toLocaleString() : "-"} members
                                </span>
                                <span>Created by <span className="text-purple-200 font-medium">{group.owner_name || "Unknown"}</span></span>
                            </div>
                        </div>
                    
                        <div className="flex flex-col items-end ml-4">
                            {group.request_pending ? (
                                <button className="bg-purple-800 text-purple-300 rounded-md px-4 py-2 font-semibold cursor-not-allowed" disabled>Request Pending</button>
                            ) : (
                                <button className="bg-purple-500 hover:bg-purple-600 text-white rounded-md px-4 py-2 font-semibold">Join group</button>
                            )}
                        </div>
                    </article>
                ))}
            </div>
        </section>
                    <section id="invitations-section" className={activeSection === "invitations" ? "flex flex-col border border-purple-500 rounded-lg p-4" : "hidden"}>
            <p>Invitations</p>
        </section>

        </main>
     );
}
 
export default GroupsPage;