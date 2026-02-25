"use client";

import Image from "next/image";
import Link from "next/link";
import { parseProfileImage } from "src/lib/utils/profileImage";
import { formatFriendlyDateTime } from "src/lib/utils/dateTime";

const CommentThread = ({
  postId,
  currentUserId,
  currentUserProfilePicture,
  commentValue,
  onCommentChange,
  onCommentImageChange,
  onSubmit,
  isCommentSubmitting,
  commentError,
  comments,
  isCommentsLoading,
  editingCommentId,
  editingCommentContent,
  onStartEdit,
  onEditContentChange,
  onSaveEdit,
  onCancelEdit,
  onDelete,
  commentActionLoadingById,
  buildAuthorLink,
  toUploadUrl,
}) => {
  const echoPhotoUploadId = `echo-photo-upload-${postId}`;

  const getAuthorHref = (userId) => {
    if (typeof buildAuthorLink === "function") {
      return buildAuthorLink(userId);
    }
    if (!userId) return "";
    return `/profile/${userId}`;
  };

  return (
    <>
      <form onSubmit={onSubmit} className="flex items-center gap-2 w-full">
        <Image
          src={parseProfileImage(currentUserProfilePicture)}
          alt="Profile Icon"
          width={25}
          height={25}
        />
        <div className="flex justify-between bg-gray-100 text-black w-full rounded-lg resize-none h-10">
          <input
            type="text"
            placeholder="Write a comment..."
            className="focus:outline-none w-full pl-1 bg-transparent"
            value={commentValue}
            onChange={(event) => onCommentChange(event.target.value)}
            disabled={isCommentSubmitting}
          />

          <label
            htmlFor={echoPhotoUploadId}
            className="flex items-center gap-1 cursor-pointer px-1"
          >
            <Image
              src="/photo_icon.svg"
              alt="Share Icon"
              width={20}
              height={20}
            />
            <input
              id={echoPhotoUploadId}
              type="file"
              className="font-medium cursor-pointer text-black hidden"
              accept="image/*"
              onChange={(event) => onCommentImageChange(event.target.files?.[0] || null)}
              disabled={isCommentSubmitting}
            />
          </label>
        </div>
        <button
          type="submit"
          className="text-sm px-2 py-1 rounded bg-blue-500 text-white disabled:opacity-50"
          disabled={isCommentSubmitting}
        >
          {isCommentSubmitting ? "Sending..." : "Send"}
        </button>
      </form>

      {commentError ? <p className="text-red-600 text-sm">{commentError}</p> : null}

      <div className="flex flex-col gap-2">
        {isCommentsLoading ? (
          <p className="text-sm text-gray-500">Loading echoes...</p>
        ) : comments.length === 0 ? (
          <p className="text-sm text-gray-500">No echoes yet.</p>
        ) : (
          comments.map((comment) => {
            const authorHref = getAuthorHref(comment.user_id);
            const authorName = `${comment.author_first_name || ""} ${comment.author_last_name || ""}`.trim() || "Unknown User";
            const createdLabel = formatFriendlyDateTime(comment.created_at_time || comment.created_at);
            const isEditing = editingCommentId === comment.id;
            const isOwnComment = comment.user_id === currentUserId;
            return (
              <div key={comment.id} className="bg-gray-50 rounded p-2">
                <div className="flex items-start justify-between gap-2 mb-1">
                  <div className="flex items-start gap-2">
                    <div className="pt-0.5">
                      <Image
                        src={parseProfileImage(comment.author_profile_picture)}
                        alt="Comment author"
                        width={20}
                        height={20}
                      />
                    </div>
                    <div className="flex flex-col leading-tight">
                      {authorHref ? (
                        <Link href={authorHref} className="text-sm font-medium">
                          {authorName}
                        </Link>
                      ) : (
                        <span className="text-sm font-medium">{authorName}</span>
                      )}
                      {createdLabel ? (
                        <span className="text-xs text-gray-500 mt-0.5">{createdLabel}</span>
                      ) : null}
                    </div>
                  </div>
                  {isOwnComment ? (
                    <div className="flex gap-2">
                      <button
                        type="button"
                        className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
                        onClick={() => onStartEdit(comment)}
                        disabled={!!commentActionLoadingById[comment.id]}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        className="text-xs bg-purple-900 hover:bg-purple-800 text-white rounded-lg px-3 py-1 disabled:opacity-50"
                        onClick={() => onDelete(comment.id)}
                        disabled={!!commentActionLoadingById[comment.id]}
                      >
                        Delete
                      </button>
                    </div>
                  ) : null}
                </div>
                {isEditing ? (
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      className="border rounded px-2 py-1 text-sm flex-1"
                      value={editingCommentContent}
                      onChange={(event) => onEditContentChange(event.target.value)}
                    />
                    <button
                      type="button"
                      className="text-xs px-2 py-1 rounded bg-blue-500 text-white disabled:opacity-50"
                      onClick={() => onSaveEdit(comment.id)}
                      disabled={!!commentActionLoadingById[comment.id]}
                    >
                      Save
                    </button>
                    <button
                      type="button"
                      className="text-xs px-2 py-1 rounded bg-gray-300 text-black"
                      onClick={onCancelEdit}
                    >
                      Cancel
                    </button>
                  </div>
                ) : (
                  <p className="text-sm">{comment.content}</p>
                )}
                {comment.image ? (
                  <div className="mt-2">
                    <Image
                      src={toUploadUrl(comment.image)}
                      alt="Comment image"
                      width={300}
                      height={180}
                      className="rounded w-full h-auto"
                    />
                  </div>
                ) : null}
              </div>
            );
          })
        )}
      </div>
    </>
  );
};

export default CommentThread;
