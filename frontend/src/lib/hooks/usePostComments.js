"use client";

import { useState } from "react";
import { createComment, deleteComment, getPostComments, updateComment } from "src/lib/services/comment";
import { deletePost, getPostById, updatePost } from "src/lib/services/post";

export function usePostComments({ setPosts } = {}) {
  const [commentsByPost, setCommentsByPost] = useState({});
  const [commentsLoadingByPost, setCommentsLoadingByPost] = useState({});
  const [commentInputByPost, setCommentInputByPost] = useState({});
  const [commentImageByPost, setCommentImageByPost] = useState({});
  const [commentSubmittingByPost, setCommentSubmittingByPost] = useState({});
  const [commentErrorByPost, setCommentErrorByPost] = useState({});
  const [editingCommentIdByPost, setEditingCommentIdByPost] = useState({});
  const [editingCommentContentByPost, setEditingCommentContentByPost] = useState({});
  const [commentActionLoadingById, setCommentActionLoadingById] = useState({});
  const [editingPostId, setEditingPostId] = useState(null);
  const [editingPostContent, setEditingPostContent] = useState("");
  const [postActionLoadingById, setPostActionLoadingById] = useState({});
  const [postActionErrorById, setPostActionErrorById] = useState({});

  function updatePostInList(postId, updateFn) {
    if (!setPosts) return;
    setPosts((prev) => prev.map((post) => (post.id === postId ? updateFn(post) : post)));
  }

  function removePostFromList(postId) {
    if (!setPosts) return;
    setPosts((prev) => prev.filter((post) => post.id !== postId));
  }

  async function loadComments(postId) {
    setCommentsLoadingByPost((prev) => ({ ...prev, [postId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      const comments = await getPostComments(postId);
      setCommentsByPost((prev) => ({ ...prev, [postId]: comments }));
    } catch (error) {
      console.error("Error loading comments:", error);
      setCommentsByPost((prev) => ({ ...prev, [postId]: [] }));
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to load echoes.",
      }));
    } finally {
      setCommentsLoadingByPost((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleCommentSubmit(event, postId) {
    event.preventDefault();

    const content = (commentInputByPost[postId] || "").trim();
    const image = commentImageByPost[postId] || null;

    if (!content) {
      setCommentErrorByPost((prev) => ({ ...prev, [postId]: "Comment content is required." }));
      return;
    }

    setCommentSubmittingByPost((prev) => ({ ...prev, [postId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));

    const formData = new FormData();
    formData.append("content", content);
    formData.append("parent_type", "post");
    formData.append("parent_id", String(postId));
    if (image) {
      formData.append("avatar", image);
    }

    try {
      await createComment(formData);
      setCommentInputByPost((prev) => ({ ...prev, [postId]: "" }));
      setCommentImageByPost((prev) => ({ ...prev, [postId]: null }));
      await loadComments(postId);
    } catch (error) {
      console.error("Error creating comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to create echo.",
      }));
    } finally {
      setCommentSubmittingByPost((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleDeleteComment(postId, commentId) {
    if (!commentId) return;

    setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      await deleteComment(commentId);
      await loadComments(postId);
    } catch (error) {
      console.error("Error deleting comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to delete echo.",
      }));
    } finally {
      setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: false }));
    }
  }

  async function handleSaveCommentEdit(postId, commentId) {
    const content = (editingCommentContentByPost[postId] || "").trim();
    if (!content) {
      setCommentErrorByPost((prev) => ({ ...prev, [postId]: "Comment content is required." }));
      return;
    }

    setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: true }));
    setCommentErrorByPost((prev) => ({ ...prev, [postId]: "" }));
    try {
      await updateComment(commentId, content);
      setEditingCommentIdByPost((prev) => ({ ...prev, [postId]: null }));
      setEditingCommentContentByPost((prev) => ({ ...prev, [postId]: "" }));
      await loadComments(postId);
    } catch (error) {
      console.error("Error updating comment:", error);
      setCommentErrorByPost((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to update echo.",
      }));
    } finally {
      setCommentActionLoadingById((prev) => ({ ...prev, [commentId]: false }));
    }
  }

  async function handleStartEditPost(postId) {
    if (!postId) return;

    setPostActionErrorById((prev) => ({ ...prev, [postId]: "" }));
    setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
    try {
      const post = await getPostById(postId);
      setEditingPostId(postId);
      setEditingPostContent(post?.content || "");
    } catch (error) {
      console.error("Error loading post for edit:", error);
      setPostActionErrorById((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to load post.",
      }));
    } finally {
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleSavePostEdit(postId) {
    const content = editingPostContent.trim();
    if (!content) {
      setPostActionErrorById((prev) => ({ ...prev, [postId]: "Post content is required." }));
      return;
    }

    setPostActionErrorById((prev) => ({ ...prev, [postId]: "" }));
    setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
    try {
      await updatePost(postId, content);
      updatePostInList(postId, (post) => ({ ...post, content }));
      setEditingPostId(null);
      setEditingPostContent("");
    } catch (error) {
      console.error("Error updating post:", error);
      setPostActionErrorById((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to update post.",
      }));
    } finally {
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
    }
  }

  async function handleDeletePost(postId) {
    if (!postId) return;

    setPostActionErrorById((prev) => ({ ...prev, [postId]: "" }));
    setPostActionLoadingById((prev) => ({ ...prev, [postId]: true }));
    try {
      await deletePost(postId);
      removePostFromList(postId);
      if (editingPostId === postId) {
        setEditingPostId(null);
        setEditingPostContent("");
      }
    } catch (error) {
      console.error("Error deleting post:", error);
      setPostActionErrorById((prev) => ({
        ...prev,
        [postId]: error?.message || "Failed to delete post.",
      }));
    } finally {
      setPostActionLoadingById((prev) => ({ ...prev, [postId]: false }));
    }
  }

  return {
    commentsByPost,
    setCommentsByPost,
    commentsLoadingByPost,
    commentInputByPost,
    setCommentInputByPost,
    commentImageByPost,
    setCommentImageByPost,
    commentSubmittingByPost,
    commentErrorByPost,
    setCommentErrorByPost,
    editingCommentIdByPost,
    setEditingCommentIdByPost,
    editingCommentContentByPost,
    setEditingCommentContentByPost,
    commentActionLoadingById,
    editingPostId,
    setEditingPostId,
    editingPostContent,
    setEditingPostContent,
    postActionLoadingById,
    postActionErrorById,
    setPostActionErrorById,
    loadComments,
    handleCommentSubmit,
    handleDeleteComment,
    handleSaveCommentEdit,
    handleStartEditPost,
    handleSavePostEdit,
    handleDeletePost,
  };
}
