"use client";


import Image from "next/image";
import { useState, useEffect } from "react";
import { getDiscoveredGroups } from "src/lib/services/discoverGroups";
import { 
    createGroup,
    getActiveGroups, 
    getPendingGroupInvites, 
    acceptGroupInvite, 
    rejectGroupInvite,
    requestToJoinGroup,
    leaveGroup
} from "src/lib/services/group";
import { parseProfileImage } from "src/lib/utils/profileImage";
import Link from "next/link";

const GroupsPage = () => {
    const currentUser = "John Doe";
    const [activeSection, setActiveSection] = useState("my-groups");
    const [searchQuery, setSearchQuery] = useState("");
    const [showCreateModal, setShowCreateModal] = useState(false);
    
    // State for groups data
    const [groups, setGroups] = useState([]);
    const [groupsLoading, setGroupsLoading] = useState(true);
    const [groupsError, setGroupsError] = useState(null);

    // State for invitations data
    const [invitations, setInvitations] = useState([]);
    const [invitationsLoading, setInvitationsLoading] = useState(false);
    const [invitationsError, setInvitationsError] = useState(null);

    // State for discover groups
    const [discoverGroups, setDiscoverGroups] = useState([]);
    const [discoverLoading, setDiscoverLoading] = useState(false);
    const [discoverError, setDiscoverError] = useState(null);

    useEffect(() => {
        // Fetch active groups and invitations when component mounts
        setGroupsLoading(true);
        setGroupsError(null);
        getActiveGroups()
            .then((res) => {
                setGroups(Array.isArray(res) ? res : []);
            })
            .catch((err) => setGroupsError(err?.message || "Failed to load your groups"))
            .finally(() => setGroupsLoading(false));
        
        // Fetch invitations immediately to show badge right away
        getPendingGroupInvites()
            .then((res) => {
                setInvitations(Array.isArray(res) ? res : []);
            })
            .catch((err) => console.error("Failed to load invitations badge:", err));
        
        // Set up polling for new invitations every 3 seconds to show badge in real-time
        const invitationInterval = setInterval(() => {
            getPendingGroupInvites()
                .then((res) => {
                    setInvitations(Array.isArray(res) ? res : []);
                })
                .catch((err) => console.error("Failed to refresh invitations badge:", err));
        }, 3000);
        
        return () => clearInterval(invitationInterval);
    }, []);

    useEffect(() => {
        // Load full invitation list when user clicks on Invitations tab
        if (activeSection === "invitations") {
            setInvitationsLoading(true);
            setInvitationsError(null);
            getPendingGroupInvites()
                .then((res) => {
                    setInvitations(Array.isArray(res) ? res : []);
                })
                .catch((err) => setInvitationsError(err?.message || "Failed to load invitations"))
                .finally(() => setInvitationsLoading(false));
        }
    }, [activeSection]);

    useEffect(() => {
        // Fetch discovered groups when user opens Discover tab
        if (activeSection === "discover") {
            setDiscoverLoading(true);
            setDiscoverError(null);
            getDiscoveredGroups()
                .then((res) => {
                    if (Array.isArray(res) && res.length > 0) setDiscoverGroups(res);
                })
                .catch((err) => setDiscoverError(err?.message || "Failed to load groups"))
                .finally(() => setDiscoverLoading(false));
        }
    }, [activeSection]);

    const handleJoinRequest = async (groupId) => {
        const group = discoverGroups.find(g => g.id === groupId);
        if (!group) return;

        try {
            await requestToJoinGroup(groupId);
            // Update local state based on join_mode (stored in group.type)
            // "auto" = direct join, anything else = pending request
            if (group.type === "auto") {
                setDiscoverGroups(discoverGroups.map(g => 
                    g.id === groupId ? { ...g, hasJoined: true, members: (g.members || 0) + 1 } : g
                ));
                // Add to my groups
                setGroups([...groups, group]);
            } else {
                setDiscoverGroups(discoverGroups.map(g => 
                    g.id === groupId ? { ...g, hasPendingRequest: true } : g
                ));
            }
        } catch (error) {
            console.error("Failed to join group:", error);
            alert(error?.message || "Failed to join group");
        }
    };

    const handleAcceptInvitation = async (groupId) => {
        try {
            await acceptGroupInvite(groupId);
            const invitation = invitations.find(inv => inv.id === groupId);
            if (invitation) {
                // Add to groups
                setGroups([...groups, { 
                    id: invitation.id, 
                    name: invitation.name, 
                    owner: invitation.owner, 
                    content: invitation.content || invitation.description,
                    members: invitation.members || 0, 
                    privacy: invitation.privacy || 'private' 
                }]);
                // Remove from invitations
                setInvitations(invitations.filter(inv => inv.id !== groupId));
            }
        } catch (error) {
            console.error("Failed to accept invitation:", error);
            alert(error?.message || "Failed to accept invitation");
        }
    };

    const handleDeclineInvitation = async (groupId) => {
        try {
            await rejectGroupInvite(groupId);
            setInvitations(invitations.filter(inv => inv.id !== groupId));
        } catch (error) {
            console.error("Failed to decline invitation:", error);
            alert(error?.message || "Failed to decline invitation");
        }
    };

    const handleLeaveGroup = async (groupId) => {
        try {
            await leaveGroup(groupId);
            setGroups(groups.filter(g => g.id !== groupId));
            alert("You have left the group");
        } catch (error) {
            console.error("Failed to leave group:", error);
            alert(error?.message || "Failed to leave group");
        }
    };

    const filterGroups = (groupList) => {
        if (!searchQuery) return groupList;
        const query = searchQuery.toLowerCase();
        return groupList.filter(g => 
            (g.name || "").toLowerCase().includes(query) ||
            (g.content || "").toLowerCase().includes(query) ||
            (g.description || "").toLowerCase().includes(query)
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
                    My Groups ({groups.length})
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
                    {invitations.length > 0 && (
                        <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center shadow-[0_0_8px_rgba(239,68,68,0.6)]">
                            {invitations.length}
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
                {groupsLoading && <p>Loading your groups...</p>}
                {groupsError && <p className="text-red-500">{groupsError}</p>}
                {!groupsLoading && filterGroups(groups).length > 0 ? (
                    filterGroups(groups).map(group => (
                        <article key={group.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-4">
                                {/* Group Avatar */}
                                <div className="w-12 h-12 rounded-md bg-purple-600 flex items-center justify-center text-white text-xl font-bold shrink-0 shadow-[0_0_10px_rgba(168,85,247,0.3)]">
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
                                    <p className="text-purple-300/70 text-sm mt-1 line-clamp-2">{group.content || group.description}</p>
                                    <div className="flex items-center gap-4 text-xs text-purple-400 mt-2">
                                        <span className="flex items-center gap-1">
                                            <Image src="/groups_icon.svg" alt="Members" width={14} height={14} className="opacity-60" />
                                            {formatMembers(group.members || group.group_members || 0)} members
                                        </span>
                                        <span>Created by {group.owner_first_name} {group.owner_last_name}</span>
                                    </div>
                                </div>
                                
                                {/* Action Buttons */}
                                <div className="flex gap-2 shrink-0">
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
                    !groupsLoading && <EmptyState 
                        message={searchQuery ? "No groups found" : "No groups joined yet"}
                        subMessage={searchQuery ? "Try a different search term" : "Join or create a group to get started"}
                    />
                )}
            </section>

            {/* Invitations Section */}
            <section id="invitations-section" className={activeSection === "invitations" ? "flex flex-col gap-4" : "hidden"}>
                {invitationsLoading && <p>Loading invitations...</p>}
                {invitationsError && <p className="text-red-500">{invitationsError}</p>}
                {!invitationsLoading && filterGroups(invitations).length > 0 ? (
                    filterGroups(invitations).map(group => (
                        <article key={group.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-4">
                                <div className="w-12 h-12 rounded-md bg-purple-600 flex items-center justify-center text-white text-xl font-bold shrink-0 shadow-[0_0_10px_rgba(168,85,247,0.3)]">
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
                                    <p className="text-purple-300/70 text-sm mt-1 line-clamp-2">{group.content || group.description}</p>
                                    <div className="flex items-center gap-4 text-xs text-purple-400 mt-2">
                                        <span className="flex items-center gap-1">
                                            <Image src="/groups_icon.svg" alt="Members" width={14} height={14} className="opacity-60" />
                                            {formatMembers(group.members || group.group_members || 0)} members
                                        </span>
                                        <span>Invited by {group.owner_first_name} {group.owner_last_name}</span>
                                    </div>
                                </div>
                                
                                <div className="flex gap-2 shrink-0">
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
                    !invitationsLoading && <EmptyState 
                        message="No pending invitations"
                        subMessage="You have no pending group invitations"
                    />
                )}
            </section>

            {/* Discover Section */}
            <section id="discover-section" className={activeSection === "discover" ? "flex flex-col gap-4" : "hidden"}>
                {discoverLoading && <p>Loading groups...</p>}
                {discoverError && <p className="text-red-500">{discoverError}</p>}
                {!discoverLoading && filterGroups(discoverGroups).length > 0 ? (
                    filterGroups(discoverGroups).map(group => (
                        <article key={group.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-4">
                                <div className="w-12 h-12 rounded-md bg-purple-600 flex items-center justify-center text-white text-xl font-bold  shrink-0 shadow-[0_0_10px_rgba(168,85,247,0.3)] overflow-hidden">
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
                                        <span>Created by {group.owner_first_name} {group.owner_last_name}</span>
                                    </div>
                                </div>
                                
                                <div className="flex gap-2 shrink-0">
                                    <Link href={`/groups/${group.id}`}>
                                        <button className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)] hover:shadow-[0_0_15px_rgba(168,85,247,0.5)]">
                                            View
                                        </button>
                                    </Link>
                                    {group.hasJoined ? null : group.hasPendingRequest ? (
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
                                            {group.type === "auto" ? "Join" : "Request"}
                                        </button>
                                    )}
                                </div>
                            </div>
                        </article>
                    ))
                ) : (
                    !discoverLoading && <EmptyState 
                        message={searchQuery ? "No groups found" : "No groups to discover"}
                        subMessage={searchQuery ? "Try a different search term" : "Check back later for new groups"}
                    />
                )}
            </section>

            {/* Create Group Modal */}
            {showCreateModal && (
                <CreateGroupModal 
                    onClose={() => setShowCreateModal(false)}
                    onGroupCreated={() => {
                        // Refresh groups list
                        getActiveGroups()
                            .then((res) => setGroups(Array.isArray(res) ? res : []))
                            .catch((err) => console.error("Failed to refresh groups:", err));
                    }}
                />
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
const CreateGroupModal = ({ onClose, onGroupCreated }) => {
    const [formData, setFormData] = useState({
        title: "",
        description: "",
        visibility: "public",
        join_mode: "auto",
    });
    const [errors, setErrors] = useState({});
    const [isSubmitting, setIsSubmitting] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        const newErrors = {};
        if (!formData.title.trim()) newErrors.title = "Title is required";
        if (!formData.description.trim()) newErrors.description = "Description is required";
        
        if (Object.keys(newErrors).length > 0) {
            setErrors(newErrors);
            return;
        }

        // Prepare FormData for API
        const formDataToSend = new FormData();
        formDataToSend.append("name", formData.title);
        formDataToSend.append("description", formData.description);
        formDataToSend.append("visibility", formData.visibility);
        formDataToSend.append("join_mode", formData.join_mode);

        setIsSubmitting(true);
        try {
            await createGroup(formDataToSend);
            alert("Group created successfully!");
            onGroupCreated?.();
            onClose();
        } catch (error) {
            console.error("Failed to create group:", error);
            setErrors({ submit: error?.message || "Failed to create group" });
        } finally {
            setIsSubmitting(false);
        }
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

                    {/* Visibility */}
                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-2">
                            Visibility
                        </label>
                        <div className="flex gap-4">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="visibility"
                                    value="public"
                                    checked={formData.visibility === "public"}
                                    onChange={() => setFormData({ ...formData, visibility: "public" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Public</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="visibility"
                                    value="private"
                                    checked={formData.visibility === "private"}
                                    onChange={() => setFormData({ ...formData, visibility: "private" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Private</span>
                            </label>
                        </div>
                    </div>

                    {/* Join Mode */}
                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-2">
                            Join Mode
                        </label>
                        <div className="flex flex-col gap-3">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="join_mode"
                                    value="auto"
                                    checked={formData.join_mode === "auto"}
                                    onChange={() => setFormData({ ...formData, join_mode: "auto" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Auto (Anyone can join)</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="join_mode"
                                    value="request"
                                    checked={formData.join_mode === "request"}
                                    onChange={() => setFormData({ ...formData, join_mode: "request" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Request (Needs approval)</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="join_mode"
                                    value="invite"
                                    checked={formData.join_mode === "invite"}
                                    onChange={() => setFormData({ ...formData, join_mode: "invite" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Invite Only</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input
                                    type="radio"
                                    name="join_mode"
                                    value="request_and_invite"
                                    checked={formData.join_mode === "request_and_invite"}
                                    onChange={() => setFormData({ ...formData, join_mode: "request_and_invite" })}
                                    className="w-4 h-4 text-purple-500 accent-purple-500"
                                />
                                <span className="text-sm text-purple-300">Both (Request or Invite)</span>
                            </label>
                        </div>
                    </div>

                    {/* Error message */}
                    {errors.submit && (
                        <div className="p-3 bg-red-500/20 border border-red-500/50 rounded-md text-red-200 text-sm">
                            {errors.submit}
                        </div>
                    )}

                    {/* Buttons */}
                    <div className="flex gap-3 justify-end pt-2">
                        <button
                            type="button"
                            onClick={onClose}
                            disabled={isSubmitting}
                            className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isSubmitting}
                            className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] hover:shadow-[0_0_20px_rgba(168,85,247,0.6)] disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {isSubmitting ? "Creating..." : "Create Group"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default GroupsPage;