"use client";


import Image from "next/image";
import { useState, useEffect } from "react";
import { getDiscoveredGroups } from "src/lib/services/discoverGroups";
import { parseProfileImage } from "src/lib/utils/profileImage";
import Link from "next/link";

const GroupsPage = () => {
    const currentUser = "John Doe";
    const [activeSection, setActiveSection] = useState("my-groups");
    const [searchQuery, setSearchQuery] = useState("");
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [groups, setGroups] = useState([
        { id: 1, name: "Web Development", owner: "John Doe", content: "A community for web developers to share knowledge, resources, and collaborate on projects.", members: 15420, privacy: "public" },
        { id: 2, name: "React Enthusiasts", owner: "Alison Johnson", content: "Everything about React, Next.js, and modern front end development. Share tips, ask questions, grow together!", members: 8932, privacy: "public" },
    ]);

    const [invitations, setInvitations] = useState([
        { id: 3, name: "Photography Club", owner: "Bob Smith", content: "Share your best shots and learn photography techniques from professionals.", members: 5621, privacy: "private", from: "Bob Smith", status: "pending" },
    ]);

    const [discoverGroups, setDiscoverGroups] = useState([
        { id: 4, name: "Tech Startups", owner: "Carol White", content: "Discuss startup ideas, funding, and entrepreneurship.", members: 3200, privacy: "public", hasPendingRequest: false, hasJoined: false },
        { id: 5, name: "Gaming Community", owner: "Dave Brown", content: "Connect with gamers, share tips, and organize gaming sessions.", members: 1500, privacy: "private", hasPendingRequest: false, hasJoined: false },
    ]);

    const [groupsLoading, setGroupsLoading] = useState(false);
    const [groupsError, setGroupsError] = useState(null);

    const myGroups = groups;
    const pendingInvitations = invitations.filter(inv => inv.status === "pending");

    useEffect(() => {
        // Fetch discovered groups when user opens Discover tab
        if (activeSection === "discover") {
            setGroupsLoading(true);
            setGroupsError(null);
            getDiscoveredGroups()
                .then((res) => {
                    // prefer API result if available, otherwise keep local mock
                    if (Array.isArray(res) && res.length > 0) setDiscoverGroups(res);
                })
                .catch((err) => setGroupsError(err?.message || "Failed to load groups"))
                .finally(() => setGroupsLoading(false));
        }
    }, [activeSection]);

    const handleJoinRequest = (groupId) => {
        const group = discoverGroups.find(g => g.id === groupId);
        if (!group) return;
        if (group.privacy === "public") {
            setDiscoverGroups(discoverGroups.map(g => 
                g.id === groupId ? { ...g, hasJoined: true, members: (g.members || 0) + 1 } : g
            ));
        } else {
            setDiscoverGroups(discoverGroups.map(g => 
                g.id === groupId ? { ...g, hasPendingRequest: true } : g
            ));
        }
    };

    const handleAcceptInvitation = (groupId) => {
        const invitation = invitations.find(inv => inv.id === groupId);
        if (invitation) {
            setGroups([...groups, { id: invitation.id, name: invitation.name, owner: invitation.owner, content: invitation.content, members: invitation.members || 0, privacy: invitation.privacy || 'private' }]);
            setInvitations(invitations.filter(inv => inv.id !== groupId));
        }
    };

    const handleDeclineInvitation = (groupId) => {
        setInvitations(invitations.filter(inv => inv.id !== groupId));
    };

    const handleLeaveGroup = (groupId) => {
        setGroups(groups.filter(g => g.id !== groupId));
    };

    const filterGroups = (groupList) => {
        if (!searchQuery) return groupList;
        return groupList.filter(g => 
            (g.name || "").toLowerCase().includes(searchQuery.toLowerCase()) ||
            (g.content || "").toLowerCase().includes(searchQuery.toLowerCase())
        );
    };

    const formatMembers = (count) => {
        if (!count && count !== 0) return "0";
        if (count >= 1000) {
            return (count / 1000).toFixed(count % 1000 === 0 ? 0 : 1) + "k";
        }
        return count.toString();
    };

    return (
        <main className="flex flex-col w-full max-w-3xl gap-4 p-4">
            {/* Header */}
            <header className="flex flex-row justify-between items-center w-full">  
                <h1 className="text-3xl font-bold text-white">Groups</h1>
                <button 
                    id="createGroupBtn" 
                    onClick={() => setShowCreateModal(true)} 
                    className="bg-blue-500 hover:bg-blue-600 rounded-md py-2 px-4 text-white flex items-center gap-2 cursor-pointer transition text-sm"
                >
                    <span className="text-lg">+</span>
                    Create Group
                </button>
            </header>

            {/* Search Bar */}
            <div className="relative">
                <Image 
                    src="/search_icon.svg" 
                    alt="Search" 
                    width={16} 
                    height={16} 
                    className="absolute left-3 top-1/2 transform -translate-y-1/2 opacity-60"
                />
                <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search groups..."
                    className="w-full pl-10 pr-4 py-2 bg-[#2a2a3e] border border-purple-500/30 rounded-md focus:outline-none focus:ring-1 focus:ring-purple-500 text-white placeholder-purple-300/50 text-sm"
                />
            </div>

            {/* Tabs */}
            <div className="flex flex-row gap-3 w-full">
                <button 
                    id="myGroupsBtn" 
                    onClick={() => setActiveSection("my-groups")} 
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                        activeSection === "my-groups"
                            ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)] shadow-purple-500/40"
                            : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                >
                    My Groups ({myGroups.length})
                </button>
                <button 
                    id="invitationsBtn" 
                    onClick={() => setActiveSection("invitations")} 
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md relative ${
                        activeSection === "invitations"
                            ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)] shadow-purple-500/40"
                            : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                >
                    Invitations
                    {pendingInvitations.length > 0 && (
                        <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center shadow-[0_0_8px_rgba(239,68,68,0.6)]">
                            {pendingInvitations.length}
                        </span>
                    )}
                </button>
                <button 
                    id="discoverBtn" 
                    onClick={() => setActiveSection("discover")} 
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                        activeSection === "discover"
                            ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)] shadow-purple-500/40"
                            : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                >
                    Discover
                </button>
               
            </div>

            {/* My Groups Section */}
            <section id="my-groups-section" className={activeSection === "my-groups" ? "flex flex-col gap-4" : "hidden"}>
                {filterGroups(myGroups).length > 0 ? (
                    filterGroups(myGroups).map(group => (
                        <article key={group.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-4">
                                {/* Group Avatar */}
                                <div className="w-12 h-12 rounded-md bg-purple-600 flex items-center justify-center text-white text-xl font-bold flex-shrink-0 shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                                    {group.name[0]}
                                </div>
                                
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 flex-wrap">
                                        <h2 className="text-lg font-bold text-purple-100">{group.name}</h2>
                                        {group.privacy === "private" ? (
                                            <Image src="/lock_icon.svg" alt="Private" width={14} height={14} />
                                        ) : (
                                            <Image src="/globe_icon.svg" alt="Public" width={14} height={14} />
                                        )}
                                        {group.owner === currentUser && (
                                            <span className="bg-green-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(34,197,94,0.4)]">Creator</span>
                                        )}
                                    </div>
                                    <p className="text-purple-300/70 text-sm mt-1 line-clamp-2">{group.content}</p>
                                    <div className="flex items-center gap-4 text-xs text-purple-400 mt-2">
                                        <span className="flex items-center gap-1">
                                            <Image src="/groups_icon.svg" alt="Members" width={14} height={14} className="opacity-60" />
                                            {formatMembers(group.members)} members
                                        </span>
                                        <span>Created by {group.owner}</span>
                                    </div>
                                </div>
                                
                                {/* Action Buttons */}
                                <div className="flex gap-2 flex-shrink-0">
                                    <Link href={`/groups/${group.id}`}>
                                        <button className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] hover:shadow-[0_0_15px_rgba(168,85,247,0.5)]">
                                            View group
                                        </button>
                                    </Link>
                                    {group.owner !== currentUser && (
                                        <button 
                                            onClick={() => handleLeaveGroup(group.id)}
                                            className="px-3 py-1.5 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                                        >
                                            Leave
                                        </button>
                                    )}
                                </div>
                            </div>
                        </article>
                    ))
                ) : (
                    <EmptyState 
                        message={searchQuery ? "No groups found" : "No groups joined yet"}
                        subMessage={searchQuery ? "Try a different search term" : "Join or create a group to get started"}
                    />
                )}
            </section>

            {/* Invitations Section */}
            <section id="invitations-section" className={activeSection === "invitations" ? "flex flex-col gap-4" : "hidden"}>
                {filterGroups(pendingInvitations).length > 0 ? (
                    filterGroups(pendingInvitations).map(group => (
                        <article key={group.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-4">
                                <div className="w-12 h-12 rounded-md bg-purple-600 flex items-center justify-center text-white text-xl font-bold flex-shrink-0 shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                                    {group.name[0]}
                                </div>
                                
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 flex-wrap">
                                        <h2 className="text-lg font-bold text-purple-100">{group.name}</h2>
                                        {group.privacy === "private" ? (
                                            <Image src="/lock_icon.svg" alt="Private" width={14} height={14} />
                                        ) : (
                                            <Image src="/globe_icon.svg" alt="Public" width={14} height={14} />
                                        )}
                                    </div>
                                    <p className="text-purple-300/70 text-sm mt-1 line-clamp-2">{group.content}</p>
                                    <div className="flex items-center gap-4 text-xs text-purple-400 mt-2">
                                        <span className="flex items-center gap-1">
                                            <Image src="/groups_icon.svg" alt="Members" width={14} height={14} className="opacity-60" />
                                            {formatMembers(group.members)} members
                                        </span>
                                        <span>Invited by {group.from}</span>
                                    </div>
                                </div>
                                
                                <div className="flex gap-2 flex-shrink-0">
                                    <button
                                        onClick={() => handleAcceptInvitation(group.id)}
                                        className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] hover:shadow-[0_0_15px_rgba(168,85,247,0.5)]"
                                    >
                                        Accept
                                    </button>
                                    <button
                                        onClick={() => handleDeclineInvitation(group.id)}
                                        className="px-3 py-1.5 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                                    >
                                        Decline
                                    </button>
                                </div>
                            </div>
                        </article>
                    ))
                ) : (
                    <EmptyState 
                        message="No pending invitations"
                        subMessage="You have no pending group invitations"
                    />
                )}
            </section>

            {/* Discover Section */}
            <section id="discover-section" className={activeSection === "discover" ? "flex flex-col gap-4" : "hidden"}>
                {groupsLoading && <p>Loading groups...</p>}
                {groupsError && <p className="text-red-500">{groupsError}</p>}
                {filterGroups(discoverGroups).length > 0 ? (
                    filterGroups(discoverGroups).map(group => (
                        <article key={group.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-4">
                                <div className="w-12 h-12 rounded-md bg-purple-600 flex items-center justify-center text-white text-xl font-bold flex-shrink-0 shadow-[0_0_10px_rgba(168,85,247,0.3)] overflow-hidden">
                                    {/* Prefer group picture when available */}
                                    {group.group_picture ? (
                                        <Image
                                            src={parseProfileImage(group.group_picture)}
                                            alt={group.name}
                                            width={48}
                                            height={48}
                                            className="object-cover w-full h-full"
                                            onError={e => { e.target.onerror = null; e.target.src = "/groups_icon.svg"; }}
                                        />
                                    ) : (
                                        <span className="text-xl">{group.name[0]}</span>
                                    )}
                                </div>
                                
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center gap-2 flex-wrap">
                                        <h2 className="text-lg font-bold text-purple-100">{group.name}</h2>
                                        {group.privacy === "private" ? (
                                            <Image src="/lock_icon.svg" alt="Private" width={14} height={14} />
                                        ) : (
                                            <Image src="/globe_icon.svg" alt="Public" width={14} height={14} />
                                        )}
                                    </div>
                                    <p className="text-purple-300/70 text-sm mt-1 line-clamp-2">{group.content || group.description || "No description provided."}</p>
                                    <div className="flex items-center gap-4 text-xs text-purple-400 mt-2">
                                        <span className="flex items-center gap-1">
                                            <Image src="/groups_icon.svg" alt="Members" width={14} height={14} className="opacity-60" />
                                            {formatMembers(group.members || group.members_count || 0)} members
                                        </span>
                                        <span>Created by {group.owner_name || group.owner || "Unknown"}</span>
                                    </div>
                                </div>
                                
                                <div className="flex gap-2 flex-shrink-0">
                                    {group.hasJoined ? (
                                        <Link href={`/groups/${group.id}`}>
                                            <button className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] hover:shadow-[0_0_15px_rgba(168,85,247,0.5)]">
                                                View group
                                            </button>
                                        </Link>
                                    ) : group.hasPendingRequest ? (
                                        <button
                                            disabled
                                            className="px-3 py-1.5 bg-purple-900/30 text-purple-400 border border-purple-500/30 rounded-md cursor-not-allowed text-sm"
                                        >
                                            Pending
                                        </button>
                                    ) : (
                                        <button
                                            onClick={() => handleJoinRequest(group.id)}
                                            className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] hover:shadow-[0_0_15px_rgba(168,85,247,0.5)]"
                                        >
                                            {group.privacy === "private" ? "Request" : "Join"}
                                        </button>
                                    )}
                                </div>
                            </div>
                        </article>
                    ))
                ) : (
                    <EmptyState 
                        message={searchQuery ? "No groups found" : "No groups to discover"}
                        subMessage={searchQuery ? "Try a different search term" : "Check back later for new groups"}
                    />
                )}
            </section>

            {/* Create Group Modal */}
            {showCreateModal && (
                <CreateGroupModal onClose={() => setShowCreateModal(false)} />
            )}
        </main>
    );
};

// Empty State Component
const EmptyState = ({ message, subMessage }) => (
    <div className="text-center py-12 bg-[#1a1a2e] rounded-lg border border-purple-500/30 shadow-[0_0_20px_rgba(168,85,247,0.1)]">
        <Image src="/groups_icon.svg" alt="Groups" width={48} height={48} className="mx-auto mb-4 opacity-50" />
        <h3 className="text-lg font-semibold text-purple-200 mb-2">{message}</h3>
        <p className="text-purple-400 text-sm">{subMessage}</p>
    </div>
);

// Create Group Modal Component
const CreateGroupModal = ({ onClose }) => {
    const [formData, setFormData] = useState({
        title: "",
        description: "",
        privacy: "public",
    });
    const [errors, setErrors] = useState({});

    const handleSubmit = (e) => {
        e.preventDefault();
        
        const newErrors = {};
        if (!formData.title.trim()) newErrors.title = "Title is required";
        if (!formData.description.trim()) newErrors.description = "Description is required";
        
        if (Object.keys(newErrors).length > 0) {
            setErrors(newErrors);
            return;
        }

        console.log("Creating group:", formData);
        alert("Group created successfully!");
        onClose();
    };

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
                <h2 className="text-xl font-bold mb-4 text-purple-100">Create New Group</h2>
                
                <form onSubmit={handleSubmit} className="space-y-4">
                    {/* Title */}
                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-1">
                            Group Title <span className="text-red-500">*</span>
                        </label>
                        <input
                            type="text"
                            value={formData.title}
                            onChange={(e) => {
                                setFormData({ ...formData, title: e.target.value });
                                setErrors({ ...errors, title: "" });
                            }}
                            className={`w-full px-3 py-2 bg-[#0d0d1a] border rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:shadow-[0_0_10px_rgba(168,85,247,0.3)] text-purple-100 placeholder-purple-400/50 text-sm ${
                                errors.title ? "border-red-500" : "border-purple-500/30"
                            }`}
                            placeholder="Enter group title..."
                        />
                        {errors.title && <p className="text-red-500 text-xs mt-1">{errors.title}</p>}
                    </div>

                    {/* Description */}
                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-1">
                            Description <span className="text-red-500">*</span>
                        </label>
                        <textarea
                            value={formData.description}
                            onChange={(e) => {
                                setFormData({ ...formData, description: e.target.value });
                                setErrors({ ...errors, description: "" });
                            }}
                            rows={3}
                            className={`w-full px-3 py-2 bg-[#0d0d1a] border rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:shadow-[0_0_10px_rgba(168,85,247,0.3)] text-purple-100 placeholder-purple-400/50 text-sm resize-none ${
                                errors.description ? "border-red-500" : "border-purple-500/30"
                            }`}
                            placeholder="Describe your group..."
                        />
                        {errors.description && <p className="text-red-500 text-xs mt-1">{errors.description}</p>}
                    </div>

                    {/* Privacy */}
                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-2">
                            Privacy
                        </label>
                        <div className="flex gap-4">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="privacy"
                                    value="public"
                                    checked={formData.privacy === "public"}
                                    onChange={() => setFormData({ ...formData, privacy: "public" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Public</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="privacy"
                                    value="private"
                                    checked={formData.privacy === "private"}
                                    onChange={() => setFormData({ ...formData, privacy: "private" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Private</span>
                            </label>
                        </div>
                    </div>

                    {/* Buttons */}
                    <div className="flex gap-3 justify-end pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] hover:shadow-[0_0_20px_rgba(168,85,247,0.6)]"
                        >
                            Create Group
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default GroupsPage;