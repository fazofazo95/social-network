"use client";

import SearchBar from "src/components/ui/SearchBar";
import Image from "next/image";

const GroupsPage = () => {
    const currentUser = "User 1";
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

    const Toggle = (event) => {
        const buttonId = event.target.id;
        const CreateGroupSection = document.getElementById("create-group-section");
        const MyGroupsSection = document.getElementById("my-groups-section");
        const InvitationsSection = document.getElementById("invitations-section");
        const DiscoverSection = document.getElementById("discover-section");
        if (buttonId === "createGroupBtn") {
          CreateGroupSection.classList.remove("hidden");
        MyGroupsSection.classList.add("hidden");
        InvitationsSection.classList.add("hidden");
        DiscoverSection.classList.add("hidden");

        } else if ( buttonId === "myGroupsBtn" ) {
        MyGroupsSection.classList.remove("hidden");
        CreateGroupSection.classList.add("hidden");
        InvitationsSection.classList.add("hidden");
        DiscoverSection.classList.add("hidden");
    } else if ( buttonId === "invitationsBtn" ){
            InvitationsSection.classList.remove("hidden");
            MyGroupsSection.classList.add("hidden");
            CreateGroupSection.classList.add("hidden");
            DiscoverSection.classList.add("hidden");        
        } else if ( buttonId === "discoverBtn" ){
            DiscoverSection.classList.remove("hidden");
            InvitationsSection.classList.add("hidden");
            MyGroupsSection.classList.add("hidden");
            CreateGroupSection.classList.add("hidden");        
        }
    };

    return (
        <main className="flex flex-col items-left w-full max-w-2xl gap-6 p-4">
        <header className="flex flex-row justify-between items-center w-full mb-4">  
        <h1 className="text-5xl font-bold p-4">Groups</h1>
        <button id="createGroupBtn" className="bg-blue-500 rounded-lg  p-0.5 w-1/5 hover:bg-blue-600 text-white pl-4 relative cursor-pointer">
        <Image src="/group_plus.svg" alt="Create Group Icon" width={20} height={20} className="inline-block -mt-1 mr-2 absolute top-2 left-2.5"/>Create Group</button>
        </header>
        <ul className="flex flex-row justify-end gap-4 border-b border-purple-500 pb-4 w-full">
            <li className="font-bold  hover:border border-purple-500 rounded-lg px-2">
                <button id="myGroupsBtn" onClick={Toggle} className="cursor-pointer">My Groups ({user.Groups.length})</button>
            </li>
            <li className="font-bold hover:border border-purple-500 rounded-lg px-2">
                <button id="invitationsBtn" onClick={Toggle} className="cursor-pointer relative">Invitations<span id="invitations-notification" className="rounded-full bg-red-600 text-xsm absolute  top-0 -right-2 p-0.5">{user.invitations?.filter(invitation => invitation.status === "pending").length || 0}</span></button>
            </li>
            <li className="font-bold hover:border border-purple-500 rounded-lg  px-2 mr-auto">
                <button id="discoverBtn" onClick={Toggle} className="cursor-pointer">Discover</button>
            </li>
            <li>
                  <SearchBar />
            </li>
        </ul>
        
<section id="my-groups-section" className="hidden">
        {user.Groups.map(group => (
        <article key={group.id} className="flex flex-col border border-purple-500 rounded-lg">
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
        <section id="discover-section" className="hidden flex-col border border-purple-500 rounded-lg p-4">
            <p>Discover</p>
        </section>
          <section id="invitations-section" className="hidden flex-col border border-purple-500 rounded-lg p-4">
            <p>Invitations</p>
        </section>

        </main>
     );
}
 
export default GroupsPage;