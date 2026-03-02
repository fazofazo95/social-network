"use client";

import Image from "next/image";
import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import {
    getGroupPage,
    getGroupMembers,
    getGroupPendingRequests,
    getGroupPendingInvites,
    acceptJoinRequest,
    rejectJoinRequest,
    removeInvite,
    kickMember,
    promoteMember,
    demoteMember,
    deleteGroup,
    inviteToGroup
} from "src/lib/services/group";
import { getDiscoveredUsers } from "src/lib/services/discover";

const GroupDetailPage = () => {
    const params = useParams();
    const groupId = params.id;
    const currentUser = "John Doe";
    
    const [activeTab, setActiveTab] = useState("posts");
    const [showInviteModal, setShowInviteModal] = useState(false);
    const [showCreateEventModal, setShowCreateEventModal] = useState(false);
    const [showCreatePostModal, setShowCreatePostModal] = useState(false);

    // Loading states
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [userRole, setUserRole] = useState(null); // owner, moderator, member, or null

    // API data
    const [group, setGroup] = useState(null);
    const [members, setMembers] = useState([]);
    const [pendingRequests, setPendingRequests] = useState([]);
    const [pendingInvites, setPendingInvites] = useState([]);

    // More UI state
    const [posts, setPosts] = useState([
        { id: 1, author: "Alice Johnson", content: "Just finished a great tutorial on React hooks! Anyone else working on React projects?", likes: 24, liked: false, comments: [], createdAt: "2 hours ago" },
        { id: 2, author: "Bob Smith", content: "Looking for collaborators on an open-source project. DM me if interested!", likes: 15, liked: false, comments: [], createdAt: "5 hours ago" },
        { id: 3, author: "John Doe", content: "Welcome to our new members! Feel free to introduce yourselves and share what you're working on.", likes: 42, liked: true, comments: [], createdAt: "1 day ago" },
    ]);

    const [expandedComments, setExpandedComments] = useState({});
    const [commentInputs, setCommentInputs] = useState({});

    const [events, setEvents] = useState([
        { id: 1, title: "Weekly Code Review", date: "March 1, 2026", time: "3:00 PM", attendees: 28, going: true },
        { id: 2, title: "React Workshop", date: "March 5, 2026", time: "2:00 PM", attendees: 45, going: false },
    ]);

    // Load group data on mount
    useEffect(() => {
        const loadGroupData = async () => {
            try {
                setLoading(true);
                setError(null);
                
                // Fetch group page (includes user's role)
                const groupData = await getGroupPage(groupId);
                if (groupData) {
                    setGroup(groupData);
                    setUserRole(groupData.role || null); // owner, moderator, member, or null
                }
                
                // If user is member, fetch members and moderator content
                if (groupData?.role) {
                    const membersData = await getGroupMembers(groupId);
                    setMembers(membersData);
                    
                    // If moderator/owner, fetch pending requests and invites
                    if (groupData.role === "owner" || groupData.role === "moderator") {
                        const requestsData = await getGroupPendingRequests(groupId);
                        setPendingRequests(requestsData);
                        
                        const invitesData = await getGroupPendingInvites(groupId);
                        setPendingInvites(invitesData);
                    }
                }
            } catch (err) {
                console.error("Failed to load group data:", err);
                setError(err?.message || "Failed to load group data");
            } finally {
                setLoading(false);
            }
        };

        if (groupId) {
            loadGroupData();
            
            // Set up polling to refresh pending requests and invites every 5 seconds
            const interval = setInterval(async () => {
                try {
                    if (userRole === "owner" || userRole === "moderator") {
                        const requestsData = await getGroupPendingRequests(groupId);
                        setPendingRequests(requestsData);
                        
                        const invitesData = await getGroupPendingInvites(groupId);
                        setPendingInvites(invitesData);
                    }
                } catch (err) {
                    console.error("Failed to refresh pending data:", err);
                }
            }, 5000);
            
            return () => clearInterval(interval);
        }
    }, [groupId, userRole]);

    const formatMembers = (count) => {
        if (count >= 1000) {
            return (count / 1000).toFixed(count % 1000 === 0 ? 0 : 1) + "k";
        }
        return count.toString();
    };

    // Moderator/Owner handlers
    const handleAcceptRequest = async (userId) => {
        try {
            await acceptJoinRequest(groupId, userId);
            setPendingRequests(pendingRequests.filter(r => r.id !== userId));
        } catch (error) {
            console.error("Failed to accept request:", error);
            alert(error?.message || "Failed to accept request");
        }
    };

    const handleDeclineRequest = async (userId) => {
        try {
            await rejectJoinRequest(groupId, userId);
            setPendingRequests(pendingRequests.filter(r => r.id !== userId));
        } catch (error) {
            console.error("Failed to reject request:", error);
            alert(error?.message || "Failed to reject request");
        }
    };

    const handleRemoveInvite = async (userId) => {
        try {
            await removeInvite(groupId, userId);
            setPendingInvites(pendingInvites.filter(i => i.id !== userId));
        } catch (error) {
            console.error("Failed to remove invite:", error);
            alert(error?.message || "Failed to remove invite");
        }
    };

    const handleKickMember = async (userId) => {
        if (!confirm("Are you sure you want to kick this member?")) return;
        try {
            await kickMember(groupId, userId);
            setMembers(members.filter(m => m.id !== userId));
        } catch (error) {
            console.error("Failed to kick member:", error);
            alert(error?.message || "Failed to kick member");
        }
    };

    const handlePromoteMember = async (userId) => {
        try {
            await promoteMember(groupId, userId);
            setMembers(members.map(m => 
                m.id === userId ? { ...m, role: "moderator" } : m
            ));
        } catch (error) {
            console.error("Failed to promote member:", error);
            alert(error?.message || "Failed to promote member");
        }
    };

    const handleDemoteMember = async (userId) => {
        try {
            await demoteMember(groupId, userId);
            setMembers(members.map(m => 
                m.id === userId ? { ...m, role: "member" } : m
            ));
        } catch (error) {
            console.error("Failed to demote member:", error);
            alert(error?.message || "Failed to demote member");
        }
    };

    const handleLike = (postId) => {
        setPosts(posts.map(post => {
            if (post.id === postId) {
                return {
                    ...post,
                    liked: !post.liked,
                    likes: post.liked ? post.likes - 1 : post.likes + 1
                };
            }
            return post;
        }));
    };

    const toggleComments = (postId) => {
        setExpandedComments(prev => ({
            ...prev,
            [postId]: !prev[postId]
        }));
    };

    const handleCommentInput = (postId, value) => {
        setCommentInputs(prev => ({
            ...prev,
            [postId]: value
        }));
    };

    const handleAddComment = (postId) => {
        const commentText = commentInputs[postId]?.trim();
        if (!commentText) return;

        setPosts(posts.map(post => {
            if (post.id === postId) {
                return {
                    ...post,
                    comments: [
                        ...post.comments,
                        {
                            id: Date.now(),
                            author: currentUser,
                            content: commentText,
                            createdAt: "Just now"
                        }
                    ]
                };
            }
            return post;
        }));

        setCommentInputs(prev => ({
            ...prev,
            [postId]: ""
        }));
    };

    const handleToggleAttend = (eventId) => {
        setEvents(prev => prev.map(ev => {
            if (ev.id === eventId) {
                const newGoing = !ev.going;
                const newAttendees = newGoing ? (ev.attendees || 0) + 1 : Math.max(0, (ev.attendees || 0) - 1);
                return { ...ev, going: newGoing, attendees: newAttendees };
            }
            return ev;
        }));
    };

    return (
        <main className="flex flex-col w-full max-w-4xl gap-6 p-4">
            {/* Back Button */}
            <Link href="/groups" className="flex items-center gap-2 text-purple-400 hover:text-purple-300 transition w-fit">
                <span>←</span>
                <span>Back to Groups</span>
            </Link>

            {loading && <p className="text-center text-purple-300">Loading group...</p>}
            {error && <p className="text-center text-red-500">{error}</p>}
            
            {!loading && !group && <p className="text-center text-purple-300">Group not found</p>}

            {!loading && group && (
            <>
            {/* Group Header */}
            <header className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 overflow-hidden">
                {/* Cover Image */}
                <div className="h-32 bg-linear-to-r from-purple-900 via-purple-600 to-purple-900 relative">
                    <div className="absolute inset-0 bg-[url('/grid-pattern.svg')] opacity-20"></div>
                </div>
                
                {/* Group Info */}
                <div className="p-6 -mt-8 relative">
                    <div className="flex items-end gap-4">
                        {/* Group Avatar */}
                        <div className="w-20 h-20 rounded-lg bg-purple-600 flex items-center justify-center text-white text-3xl font-bold shadow-[0_0_20px_rgba(168,85,247,0.4)] border-4 border-[#1a1a2e]">
                            {group?.name?.[0]}
                        </div>
                        
                        <div className="flex-1 pb-2">
                            <div className="flex items-center gap-3">
                                <h1 className="text-2xl font-bold text-purple-100">{group.name}</h1>
                                {group.visibility === "private" ? (
                                    <Image src="/lock_icon.svg" alt="Private" width={18} height={18} />
                                ) : (
                                    <Image src="/globe_icon.svg" alt="Public" width={18} height={18} />
                                )}
                                {userRole === "owner" && (
                                    <span className="bg-green-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(34,197,94,0.4)]">Owner</span>
                                )}
                                {userRole === "moderator" && (
                                    <span className="bg-blue-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(59,130,246,0.4)]">Moderator</span>
                                )}
                            </div>
                            <div className="flex items-center gap-4 text-sm text-purple-400 mt-1">
                                <span className="flex items-center gap-1">
                                    <Image src="/groups_icon.svg" alt="Members" width={14} height={14} className="opacity-60" />
                                    {formatMembers(group.group_members || group.members)} members
                                </span>
                                <span>Created {group.created_at || group.createdAt}</span>
                            </div>
                        </div>

                        {/* Action Buttons */}
                        <div className="flex gap-2">
                            {(userRole === "owner" || userRole === "moderator") && (
                                <button 
                                    onClick={() => setShowInviteModal(true)}
                                    className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition cursor-pointer text-sm flex items-center gap-2"
                                >
                                    <span>+</span> Invite
                                </button>
                            )}
                            {userRole === "owner" && (
                                <button className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm">
                                    <Image src="/settings_icon.svg" alt="Settings" width={16} height={16} />
                                </button>
                            )}
                        </div>
                    </div>

                    <p className="text-purple-300/80 mt-4">{group.description}</p>
                </div>
            </header>

            {/* Tabs */}
            <div className="flex flex-row gap-3 w-full flex-wrap">
                <button 
                    onClick={() => setActiveTab("posts")} 
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                        activeTab === "posts"
                            ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                            : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                >
                    Posts
                </button>
                <button 
                    onClick={() => setActiveTab("members")} 
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                        activeTab === "members"
                            ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                            : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                >
                    Members ({members.length})
                </button>
                <button 
                    onClick={() => setActiveTab("events")} 
                    className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md ${
                        activeTab === "events"
                            ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                            : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                    }`}
                >
                    Events ({events.length})
                </button>
                {(userRole === "owner" || userRole === "moderator") && (
                    <>
                        <button 
                            onClick={() => setActiveTab("requests")} 
                            className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md relative ${
                                activeTab === "requests"
                                    ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                                    : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                            }`}
                        >
                            Join Requests
                            {pendingRequests.length > 0 && (
                                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center shadow-[0_0_8px_rgba(239,68,68,0.6)]">
                                    {pendingRequests.length}
                                </span>
                            )}
                        </button>
                        <button 
                            onClick={() => setActiveTab("invites")} 
                            className={`py-2 px-4 text-sm font-medium transition-all cursor-pointer rounded-md relative ${
                                activeTab === "invites"
                                    ? "bg-purple-900/40 text-purple-200 border border-purple-500 shadow-[0_0_10px_rgba(168,85,247,0.4)]"
                                    : "text-purple-400 hover:text-purple-200 hover:bg-purple-900/20 border border-transparent"
                            }`}
                        >
                            Pending Invites
                            {pendingInvites.length > 0 && (
                                <span className="absolute -top-1 -right-1 bg-orange-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center shadow-[0_0_8px_rgba(249,115,22,0.6)]">
                                    {pendingInvites.length}
                                </span>
                            )}
                        </button>
                    </>
                )}
            </div>

            {/* Posts Section */}
            {activeTab === "posts" && (
                <section className="flex flex-col gap-4">
                    {/* Create Post */}
                    <div className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4">
                        <button 
                            onClick={() => setShowCreatePostModal(true)}
                            className="w-full text-left px-4 py-3 bg-[#0d0d1a] border border-purple-500/30 rounded-md text-purple-400/50 hover:border-purple-500/50 transition cursor-pointer"
                        >
                            Write something to the group...
                        </button>
                    </div>

                    {/* Posts List */}
                    {posts.map(post => (
                        <article key={post.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-start gap-3">
                                <div className="w-10 h-10 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                                    {post.author[0]}
                                </div>
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <span className="font-semibold text-purple-100">{post.author}</span>
                                        <span className="text-purple-400/60 text-sm">{post.createdAt}</span>
                                    </div>
                                    <p className="text-purple-300/80 mt-2">{post.content}</p>
                                    <div className="flex items-center gap-6 mt-4 text-sm text-purple-400">
                                        <button 
                                            onClick={() => handleLike(post.id)}
                                            className={`flex items-center gap-2 transition cursor-pointer ${post.liked ? 'text-purple-300' : 'hover:text-purple-300'}`}
                                        >
                                            <Image 
                                                src={post.liked ? "/ripples/Ripple_icon.svg" : "/ripples/Unripple_icon.svg"} 
                                                alt="Like" 
                                                width={18} 
                                                height={18} 
                                            />
                                            <span>{post.likes}</span>
                                        </button>
                                        <button 
                                            onClick={() => toggleComments(post.id)}
                                            className={`flex items-center gap-2 transition cursor-pointer ${expandedComments[post.id] ? 'text-purple-300' : 'hover:text-purple-300'}`}
                                        >
                                            <Image src="/echo_icon.svg" alt="Comment" width={18} height={18} />
                                            <span>{post.comments.length}</span>
                                        </button>
                                    </div>

                                    {/* Comments Section */}
                                    {expandedComments[post.id] && (
                                        <div className="mt-4 pt-4 border-t border-purple-500/20">
                                            {/* Comment Input */}
                                            <div className="flex gap-2 mb-4">
                                                <div className="w-8 h-8 rounded-full bg-purple-600 flex items-center justify-center text-white text-sm font-bold shadow-[0_0_8px_rgba(168,85,247,0.3)]">
                                                    {currentUser[0]}
                                                </div>
                                                <div className="flex-1 flex gap-2">
                                                    <input
                                                        type="text"
                                                        value={commentInputs[post.id] || ""}
                                                        onChange={(e) => handleCommentInput(post.id, e.target.value)}
                                                        onKeyDown={(e) => e.key === "Enter" && handleAddComment(post.id)}
                                                        placeholder="Write a comment..."
                                                        className="flex-1 px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 placeholder-purple-400/50 text-sm"
                                                    />
                                                    <button
                                                        onClick={() => handleAddComment(post.id)}
                                                        className="px-3 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                                                    >
                                                        Echo
                                                    </button>
                                                </div>
                                            </div>

                                            {/* Comments List */}
                                            {post.comments.length > 0 ? (
                                                <div className="space-y-3">
                                                    {post.comments.map(comment => (
                                                        <div key={comment.id} className="flex gap-2">
                                                            <div className="w-8 h-8 rounded-full bg-purple-700 flex items-center justify-center text-white text-sm font-bold">
                                                                {comment.author[0]}
                                                            </div>
                                                            <div className="flex-1 bg-[#0d0d1a] rounded-md p-3 border border-purple-500/20">
                                                                <div className="flex items-center gap-2">
                                                                    <span className="font-semibold text-purple-100 text-sm">{comment.author}</span>
                                                                    <span className="text-purple-400/50 text-xs">{comment.createdAt}</span>
                                                                </div>
                                                                <p className="text-purple-300/80 text-sm mt-1">{comment.content}</p>
                                                            </div>
                                                        </div>
                                                    ))}
                                                </div>
                                            ) : (
                                                <p className="text-purple-400/50 text-sm text-center py-2">No echoes yet. Be the first to comment!</p>
                                            )}
                                        </div>
                                    )}
                                </div>
                            </div>
                        </article>
                    ))}
                </section>
            )}

            {/* Members Section */}
            {activeTab === "members" && (
                <section className="flex flex-col gap-4">
                    {members.map(member => (
                        <article key={member.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-center gap-4">
                                <div className="w-12 h-12 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                                    {(member.first_name || "U")[0]}
                                </div>
                                <div className="flex-1">
                                    <div className="flex items-center gap-2">
                                        <span className="font-semibold text-purple-100">{member.first_name} {member.last_name}</span>
                                        {member.role === "creator" && (
                                            <span className="bg-green-500 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(34,197,94,0.4)]">Creator</span>
                                        )}
                                        {member.role === "moderator" && (
                                            <span className="bg-purple-600 text-white text-xs px-2 py-0.5 rounded shadow-[0_0_8px_rgba(168,85,247,0.4)]">Moderator</span>
                                        )}
                                    </div>
                                </div>
                                {userRole === "owner" && `${member.first_name} ${member.last_name}` !== currentUser && (
                                    <div className="flex gap-2">
                                        {member.role === "member" && (
                                            <button className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                                                Promote
                                            </button>
                                        )}
                                        <button className="px-3 py-1.5 bg-purple-900/30 hover:bg-red-900/30 text-purple-300 hover:text-red-300 border border-purple-500/30 hover:border-red-500/30 rounded-md transition cursor-pointer text-sm">
                                            Remove
                                        </button>
                                    </div>
                                )}
                            </div>
                        </article>
                    ))}
                </section>
            )}

            {/* Events Section */}
            {activeTab === "events" && (
                <section className="flex flex-col gap-4">
                    {(userRole === "moderator" || userRole === "owner") && (
                        <button 
                            onClick={() => setShowCreateEventModal(true)}
                            className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-md transition cursor-pointer text-sm flex items-center gap-2 w-fit"
                        >
                            <span>+</span> Create Event
                        </button>
                    )}
                    
                    {events.map(event => (
                        <article key={event.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                            <div className="flex items-center justify-between">
                                <div>
                                    <h3 className="font-semibold text-purple-100 text-lg">{event.title}</h3>
                                    <div className="flex items-center gap-4 text-sm text-purple-400 mt-1">
                                        <span>📅 {event.date}</span>
                                        <span>🕐 {event.time}</span>
                                        <span>👥 {event.attendees} attending</span>
                                    </div>
                                </div>
                                <button 
                                    onClick={() => handleToggleAttend(event.id)}
                                    className={`px-4 py-2 rounded-md transition cursor-pointer text-sm ${
                                        event.going 
                                            ? "bg-purple-900/30 text-purple-300 border border-purple-500/30 hover:bg-purple-900/50"
                                            : "bg-purple-600 hover:bg-purple-500 text-white shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                                    }`}
                                >
                                    {event.going ? "Going ✓" : "Attend"}
                                </button>
                            </div>
                        </article>
                    ))}
                </section>
            )}

            {/* Join Requests Section */}
            {activeTab === "requests" && (userRole === "owner" || userRole === "moderator") && (
                <section className="flex flex-col gap-4">
                    {pendingRequests.length > 0 ? (
                        pendingRequests.map(request => (
                            <article key={request.id} className="bg-[#1a1a2e] rounded-lg border border-purple-500/30 p-4 hover:border-purple-500/50 hover:shadow-[0_0_15px_rgba(168,85,247,0.15)] transition-all">
                                <div className="flex items-center gap-4">
                                    <div className="w-12 h-12 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(168,85,247,0.3)]">
                                        {(request.first_name || "U")[0]}{(request.last_name || "")?.[0] || ""}
                                    </div>
                                    <div className="flex-1">
                                        <span className="font-semibold text-purple-100">{request.first_name} {request.last_name}</span>
                                        <p className="text-purple-400/60 text-sm">Requested to join</p>
                                    </div>
                                    <div className="flex gap-2">
                                        <button 
                                            onClick={() => handleAcceptRequest(request.id)}
                                            className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_10px_rgba(168,85,247,0.3)]"
                                        >
                                            Accept
                                        </button>
                                        <button 
                                            onClick={() => handleDeclineRequest(request.id)}
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
                            message="No pending requests"
                            subMessage="All join requests have been handled"
                        />
                    )}
                </section>
            )}

            {/* Pending Invites Section */}
            {activeTab === "invites" && (userRole === "owner" || userRole === "moderator") && (
                <section className="flex flex-col gap-4">
                    {pendingInvites.length > 0 ? (
                        pendingInvites.map(invite => (
                            <article key={invite.id} className="bg-[#1a1a2e] rounded-lg border border-orange-500/30 p-4 hover:border-orange-500/50 hover:shadow-[0_0_15px_rgba(249,115,22,0.15)] transition-all">
                                <div className="flex items-center gap-4">
                                    <div className="w-12 h-12 rounded-full bg-orange-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(249,115,22,0.3)]">
                                        {(invite.first_name || "U")[0]}{(invite.last_name || "")?.[0] || ""}
                                    </div>
                                    <div className="flex-1">
                                        <span className="font-semibold text-purple-100">{invite.first_name} {invite.last_name}</span>
                                        <p className="text-purple-400/60 text-sm">Pending invite</p>
                                    </div>
                                    <div className="flex gap-2">
                                        <button 
                                            onClick={() => handleRemoveInvite(invite.id)}
                                            className="px-3 py-1.5 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                                        >
                                            Cancel
                                        </button>
                                    </div>
                                </div>
                            </article>
                        ))
                    ) : (
                        <EmptyState 
                            message="No pending invites"
                            subMessage="All invites have been sent or accepted"
                        />
                    )}
                </section>
            )}

            {/* Invite Modal */}
            {showInviteModal && (
                <InviteModal onClose={() => setShowInviteModal(false)} members={members} groupId={groupId} />
            )}

            {/* Create Post Modal */}
            {showCreatePostModal && (
                <CreatePostModal onClose={() => setShowCreatePostModal(false)} />
            )}

            {/* Create Event Modal */}
            {showCreateEventModal && (
                <CreateEventModal onClose={() => setShowCreateEventModal(false)} />
            )}
            </>
            )}
        </main>
    );
};

// Empty State Component
const EmptyState = ({ message, subMessage }) => (
    <div className="text-center py-12 bg-[#1a1a2e] rounded-lg border border-purple-500/30 shadow-[0_0_20px_rgba(168,85,247,0.1)]">
        <Image src="/groups_icon.svg" alt="Empty" width={48} height={48} className="mx-auto mb-4 opacity-50" />
        <h3 className="text-lg font-semibold text-purple-200 mb-2">{message}</h3>
        <p className="text-purple-400 text-sm">{subMessage}</p>
    </div>
);

// Invite Modal Component
const InviteModal = ({ onClose, members = [], groupId }) => {
    const [searchQuery, setSearchQuery] = useState("");
    const [invitedUsers, setInvitedUsers] = useState({});
    const [availableUsers, setAvailableUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [sendingInvites, setSendingInvites] = useState(false);

    // Fetch available users on mount
    useEffect(() => {
        const fetchUsers = async () => {
            try {
                setLoading(true);
                setError(null);
                const users = await getDiscoveredUsers();
                
                // Filter out members already in the group
                const memberIds = new Set(members.map(m => m.id));
                const filtered = users.filter(user => !memberIds.has(user.id));
                
                setAvailableUsers(filtered);
            } catch (err) {
                console.error("Failed to load users:", err);
                setError(err?.message || "Failed to load users");
            } finally {
                setLoading(false);
            }
        };
        
        fetchUsers();
    }, [members]);

    // Filter users based on search query
    const filteredUsers = availableUsers.filter(user => {
        const query = searchQuery.toLowerCase();
        return (
            user.first_name.toLowerCase().includes(query) ||
            user.last_name.toLowerCase().includes(query)
        );
    });

    const handleInviteUser = (userId) => {
        setInvitedUsers(prev => ({
            ...prev,
            [userId]: !prev[userId]
        }));
    };

    const handleSendInvites = async () => {
        const selectedUserIds = Object.keys(invitedUsers).filter(id => invitedUsers[id]);
        if (selectedUserIds.length === 0) {
            alert("Please select at least one user to invite");
            return;
        }

        setSendingInvites(true);
        try {
            // Send invites to each selected user
            await Promise.all(
                selectedUserIds.map(userId => inviteToGroup(groupId, userId))
            );
            alert(`Invitations sent to ${selectedUserIds.length} user(s)!`);
            onClose();
        } catch (err) {
            console.error("Failed to send invites:", err);
            alert(err?.message || "Failed to send invites");
        } finally {
            setSendingInvites(false);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
                <h2 className="text-xl font-bold mb-4 text-purple-100">Invite Members</h2>
                
                <div className="relative mb-4">
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
                        placeholder="Search users to invite..."
                        className="w-full pl-10 pr-4 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:shadow-[0_0_10px_rgba(168,85,247,0.3)] text-purple-100 placeholder-purple-400/50 text-sm"
                    />
                </div>

                {/* Users List */}
                <div className="max-h-80 overflow-y-auto mb-4 space-y-2">
                    {loading ? (
                        <p className="text-purple-400/60 text-sm text-center py-8">Loading users...</p>
                    ) : error ? (
                        <p className="text-red-400 text-sm text-center py-8">{error}</p>
                    ) : filteredUsers.length > 0 ? (
                        filteredUsers.map(user => (
                            <div key={user.id} className="flex items-center gap-3 p-3 bg-[#0d0d1a] border border-purple-500/20 rounded-md hover:border-purple-500/40 transition">
                                <div className="w-10 h-10 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shrink-0 shadow-[0_0_8px_rgba(168,85,247,0.3)]">
                                    {(user.first_name || "U")[0]}{(user.last_name || "")[0]}
                                </div>
                                <div className="flex-1">
                                    <p className="text-purple-100 font-medium text-sm">{user.first_name} {user.last_name}</p>
                                </div>
                                <input
                                    type="checkbox"
                                    checked={invitedUsers[user.id] || false}
                                    onChange={() => handleInviteUser(user.id)}
                                    disabled={sendingInvites}
                                    className="w-5 h-5 accent-purple-500 cursor-pointer disabled:opacity-50"
                                />
                            </div>
                        ))
                    ) : (
                        <p className="text-purple-400/60 text-sm text-center py-8">
                            {searchQuery ? "No users found matching your search" : "All available users are already in the group"}
                        </p>
                    )}
                </div>

                <div className="flex gap-3 justify-end pt-2">
                    <button
                        type="button"
                        onClick={onClose}
                        disabled={sendingInvites}
                        className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm disabled:opacity-50"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSendInvites}
                        disabled={Object.values(invitedUsers).every(v => !v) || sendingInvites}
                        className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)] disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {sendingInvites ? "Sending..." : "Send Invites"}
                    </button>
                </div>
            </div>
        </div>
    );
};

// Create Post Modal Component
const CreatePostModal = ({ onClose }) => {
    const [content, setContent] = useState("");

    const handleSubmit = (e) => {
        e.preventDefault();
        if (!content.trim()) return;
        console.log("Creating post:", content);
        alert("Post created!");
        onClose();
    };

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
                <h2 className="text-xl font-bold mb-4 text-purple-100">Create Post</h2>
                
                <form onSubmit={handleSubmit} className="space-y-4">
                    <textarea
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        rows={4}
                        className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 focus:shadow-[0_0_10px_rgba(168,85,247,0.3)] text-purple-100 placeholder-purple-400/50 text-sm resize-none"
                        placeholder="What's on your mind?"
                    />

                    <div className="flex gap-3 justify-end">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)]"
                        >
                            Post
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

// Create Event Modal Component
const CreateEventModal = ({ onClose }) => {
    const [formData, setFormData] = useState({
        title: "",
        date: "",
        time: "",
        description: "",
    });

    const handleSubmit = (e) => {
        e.preventDefault();
        if (!formData.title.trim()) return;
        console.log("Creating event:", formData);
        alert("Event created!");
        onClose();
    };

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
            <div className="bg-[#1a1a2e] rounded-lg max-w-lg w-full p-6 border border-purple-500/50 shadow-[0_0_30px_rgba(168,85,247,0.3)]">
                <h2 className="text-xl font-bold mb-4 text-purple-100">Create Event</h2>
                
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-1">Event Title</label>
                        <input
                            type="text"
                            value={formData.title}
                            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                            className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 placeholder-purple-400/50 text-sm"
                            placeholder="Enter event title..."
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-purple-300 mb-1">Date</label>
                            <input
                                type="date"
                                value={formData.date}
                                onChange={(e) => setFormData({ ...formData, date: e.target.value })}
                                className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 text-sm"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-purple-300 mb-1">Time</label>
                            <input
                                type="time"
                                value={formData.time}
                                onChange={(e) => setFormData({ ...formData, time: e.target.value })}
                                className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 text-sm"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-purple-300 mb-1">Description</label>
                        <textarea
                            value={formData.description}
                            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                            rows={3}
                            className="w-full px-3 py-2 bg-[#0d0d1a] border border-purple-500/30 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500 text-purple-100 placeholder-purple-400/50 text-sm resize-none"
                            placeholder="Describe your event..."
                        />
                    </div>

                    <div className="flex gap-3 justify-end">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 bg-purple-900/30 hover:bg-purple-900/50 text-purple-300 border border-purple-500/30 rounded-md transition cursor-pointer text-sm"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 bg-purple-600 hover:bg-purple-500 text-white rounded-md transition cursor-pointer text-sm shadow-[0_0_15px_rgba(168,85,247,0.4)]"
                        >
                            Create Event
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default GroupDetailPage;
