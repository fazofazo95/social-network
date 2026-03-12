"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import {
  fetchUserData,
  fetchContentSettings,
  fetchVisibilitySettings,
  updateContentSettings,
  updateUserAvatar,
  updateUserCover,
  updateVisibilitySettings,
} from "src/lib/services/user";
import Avatar from "src/components/ui/Avatar";
import { getApiBaseUrl } from "src/lib/apiClient";

const visibilityKeys = [
  "email_vis",
  "birthday_date_vis",
  "relationship_status_vis",
  "employed_at_vis",
  "phone_number_vis",
  "about_me_vis",
  "nickname_vis",
  "follow_vis",
];

const contentDefaults = {
  first_name: "",
  last_name: "",
  birthday_date: "",
  relationship_status: "",
  employed_at: "",
  location: "",
  phone_number: "",
  nickname: "",
  about_me: "",
};

const visibilityDefaults = {
  profile_type: "public",
  email_vis: "hidden",
  birthday_date_vis: "visible",
  relationship_status_vis: "visible",
  employed_at_vis: "visible",
  phone_number_vis: "hidden",
  about_me_vis: "visible",
  nickname_vis: "visible",
  follow_vis: "hidden",
};

function normalizeVisibility(value, fallback = "hidden") {
  const v = String(value || "").toLowerCase();
  if (v === "visible" || v === "hidden") {
    return v;
  }
  return fallback;
}

function normalizeProfileType(value) {
  return String(value || "").toLowerCase() === "private" ? "private" : "public";
}

export default function SettingsPage() {
  const [contentForm, setContentForm] = useState(contentDefaults);
  const [visibilityForm, setVisibilityForm] = useState(visibilityDefaults);

  const [isLoading, setIsLoading] = useState(true);
  const [loadingError, setLoadingError] = useState("");

  const [isSavingContent, setIsSavingContent] = useState(false);
  const [isSavingVisibility, setIsSavingVisibility] = useState(false);
  const [isSavingAvatar, setIsSavingAvatar] = useState(false);
  const [isSavingCover, setIsSavingCover] = useState(false);
  const [contentStatus, setContentStatus] = useState("");
  const [visibilityStatus, setVisibilityStatus] = useState("");
  const [avatarStatus, setAvatarStatus] = useState("");
  const [coverStatus, setCoverStatus] = useState("");
  const [profilePicture, setProfilePicture] = useState("");
  const [coverImage, setCoverImage] = useState("");
  const [avatarFile, setAvatarFile] = useState(null);
  const [coverFile, setCoverFile] = useState(null);

  function toCoverUrl(path) {
    if (!path) return "/example_cover.png";
    if (path.startsWith("http://") || path.startsWith("https://") || path.startsWith("data:")) {
      return path;
    }
    if (path.startsWith("/uploads/")) {
      return `${getApiBaseUrl()}${path}`;
    }
    return "/example_cover.png";
  }

  async function loadSettings() {
    setIsLoading(true);
    setLoadingError("");
    setContentStatus("");
    setVisibilityStatus("");
    setAvatarStatus("");
    setCoverStatus("");

    try {
      const [contentData, visibilityData, profileData] = await Promise.all([
        fetchContentSettings(),
        fetchVisibilitySettings(),
        fetchUserData("me").catch(() => null),
      ]);

      setContentForm({
        ...contentDefaults,
        ...(contentData || {}),
      });

      const safeVisibility = visibilityData || {};
      const nextVisibility = {
        profile_type: normalizeProfileType(safeVisibility.profile_type || visibilityDefaults.profile_type),
        email_vis: normalizeVisibility(safeVisibility.email_vis, visibilityDefaults.email_vis),
        birthday_date_vis: normalizeVisibility(safeVisibility.birthday_date_vis, visibilityDefaults.birthday_date_vis),
        relationship_status_vis: normalizeVisibility(
          safeVisibility.relationship_status_vis,
          visibilityDefaults.relationship_status_vis
        ),
        employed_at_vis: normalizeVisibility(safeVisibility.employed_at_vis, visibilityDefaults.employed_at_vis),
        phone_number_vis: normalizeVisibility(safeVisibility.phone_number_vis, visibilityDefaults.phone_number_vis),
        about_me_vis: normalizeVisibility(safeVisibility.about_me_vis, visibilityDefaults.about_me_vis),
        nickname_vis: normalizeVisibility(safeVisibility.nickname_vis, visibilityDefaults.nickname_vis),
        follow_vis: normalizeVisibility(safeVisibility.follow_vis, visibilityDefaults.follow_vis),
      };

      setVisibilityForm(nextVisibility);
      setProfilePicture(profileData?.profile_picture || "");
      setCoverImage(profileData?.cover_image || "");
    } catch (error) {
      console.error("Failed to load settings:", error);
      setLoadingError(error?.message || "Failed to load settings.");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    loadSettings();
  }, []);

  function handleContentChange(event) {
    const { name, value } = event.target;
    setContentForm((prev) => ({ ...prev, [name]: value }));
  }

  function handleVisibilityChange(event) {
    const { name, value } = event.target;
    setVisibilityForm((prev) => ({ ...prev, [name]: value }));
  }

  function handleAvatarFileChange(event) {
    setAvatarStatus("");
    setAvatarFile(event.target.files?.[0] || null);
  }

  function handleCoverFileChange(event) {
    setCoverStatus("");
    setCoverFile(event.target.files?.[0] || null);
  }

  async function handleSaveAvatar(event) {
    event.preventDefault();
    setAvatarStatus("");

    if (!avatarFile) {
      setAvatarStatus("Please choose an image first.");
      return;
    }

    const formData = new FormData();
    formData.append("avatar", avatarFile);

    try {
      setIsSavingAvatar(true);
      const updated = await updateUserAvatar(formData);
      setProfilePicture(updated?.profile_picture || "");
      setAvatarFile(null);
      setAvatarStatus("Profile image updated.");
    } catch (error) {
      console.error("Failed to update profile image:", error);
      setAvatarStatus(error?.message || "Failed to update profile image.");
    } finally {
      setIsSavingAvatar(false);
    }
  }

  async function handleSaveCover(event) {
    event.preventDefault();
    setCoverStatus("");

    if (!coverFile) {
      setCoverStatus("Please choose an image first.");
      return;
    }

    const formData = new FormData();
    formData.append("cover", coverFile);

    try {
      setIsSavingCover(true);
      const updated = await updateUserCover(formData);
      setCoverImage(updated?.cover_image || "");
      setCoverFile(null);
      setCoverStatus("Cover image updated.");
    } catch (error) {
      console.error("Failed to update cover image:", error);
      setCoverStatus(error?.message || "Failed to update cover image.");
    } finally {
      setIsSavingCover(false);
    }
  }

  async function handleSaveContent(event) {
    event.preventDefault();
    setIsSavingContent(true);
    setContentStatus("");

    const payload = Object.fromEntries(
      Object.entries(contentForm).map(([key, value]) => [key, typeof value === "string" ? value.trim() : value])
    );

    try {
      const updated = await updateContentSettings(payload);
      setContentForm({ ...contentDefaults, ...(updated || payload) });
      setContentStatus("Content settings saved.");
    } catch (error) {
      console.error("Failed to save content settings:", error);
      setContentStatus(error?.message || "Failed to save content settings.");
    } finally {
      setIsSavingContent(false);
    }
  }

  async function handleSaveVisibility(event) {
    event.preventDefault();
    setIsSavingVisibility(true);
    setVisibilityStatus("");

    const payload = {
      profile_type: normalizeProfileType(visibilityForm.profile_type),
    };

    visibilityKeys.forEach((key) => {
      payload[key] = normalizeVisibility(visibilityForm[key], visibilityDefaults[key]);
    });

    try {
      const updated = await updateVisibilitySettings(payload);
      setVisibilityForm((prev) => ({
        ...prev,
        ...(updated || payload),
        profile_type: normalizeProfileType((updated || payload).profile_type),
      }));
      setVisibilityStatus("Visibility settings saved.");
    } catch (error) {
      console.error("Failed to save visibility settings:", error);
      setVisibilityStatus(error?.message || "Failed to save visibility settings.");
    } finally {
      setIsSavingVisibility(false);
    }
  }

  return (
    <div className="w-full max-w-2xl flex flex-col gap-6">
      <section className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-100 w-full p-5">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-purple-400">Manage profile content and privacy visibility.</p>
      </section>

      {isLoading ? (
        <section className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-300 w-full p-5">
          Loading settings...
        </section>
      ) : null}

      {!isLoading && loadingError ? (
        <section className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-red-400 w-full p-5">
          {loadingError}
        </section>
      ) : null}

      {!isLoading && !loadingError ? (
        <form onSubmit={handleSaveCover} className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-100 w-full p-5 flex flex-col gap-4">
          <h2 className="text-xl font-bold">Cover Image</h2>

          <div className="flex flex-col gap-3">
            <Image
              src={toCoverUrl(coverImage)}
              alt="Current cover image"
              width={800}
              height={180}
              className="rounded-lg w-full h-36 object-cover"
            />
            <input type="file" accept="image/*" onChange={handleCoverFileChange} className="text-sm text-purple-300 file:bg-purple-900/30 file:text-purple-300 file:border file:border-purple-500/30 file:rounded-md file:px-3 file:py-1 file:mr-3 file:cursor-pointer" />
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-purple-400">{coverStatus}</span>
            <button
              type="submit"
              className="bg-purple-600 hover:bg-purple-500 text-white rounded-md px-4 py-2 disabled:opacity-50 shadow-[0_0_10px_rgba(168,85,247,0.3)] transition cursor-pointer"
              disabled={isSavingCover}
            >
              {isSavingCover ? "Saving..." : "Save Cover"}
            </button>
          </div>
        </form>
      ) : null}

      {!isLoading && !loadingError ? (
        <form onSubmit={handleSaveAvatar} className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-100 w-full p-5 flex flex-col gap-4">
          <h2 className="text-xl font-bold">Profile Image</h2>

          <div className="flex items-center gap-4">
            <Avatar
              src={profilePicture}
              name={`${contentForm.first_name || ""} ${contentForm.last_name || ""}`.trim()}
              size={64}
              className="shadow-[0_0_10px_rgba(168,85,247,0.3)]"
            />
            <input type="file" accept="image/*" onChange={handleAvatarFileChange} className="text-sm text-purple-300 file:bg-purple-900/30 file:text-purple-300 file:border file:border-purple-500/30 file:rounded-md file:px-3 file:py-1 file:mr-3 file:cursor-pointer" />
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-purple-400">{avatarStatus}</span>
            <button
              type="submit"
              className="bg-purple-600 hover:bg-purple-500 text-white rounded-md px-4 py-2 disabled:opacity-50 shadow-[0_0_10px_rgba(168,85,247,0.3)] transition cursor-pointer"
              disabled={isSavingAvatar}
            >
              {isSavingAvatar ? "Saving..." : "Save Image"}
            </button>
          </div>
        </form>
      ) : null}

      {!isLoading && !loadingError ? (
        <form onSubmit={handleSaveContent} className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-100 w-full p-5 flex flex-col gap-4">
          <h2 className="text-xl font-bold">Profile Content</h2>

          <div className="grid grid-cols-2 gap-4">
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              First Name
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="first_name"
                value={contentForm.first_name}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Last Name
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="last_name"
                value={contentForm.last_name}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Birthday Date
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="birthday_date"
                type="date"
                value={contentForm.birthday_date}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Relationship Status
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="relationship_status"
                value={contentForm.relationship_status}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Employed At
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="employed_at"
                value={contentForm.employed_at}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Location
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="location"
                value={contentForm.location}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Phone Number
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="phone_number"
                value={contentForm.phone_number}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 col-span-2 text-purple-300">
              Nickname
              <input
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50"
                name="nickname"
                value={contentForm.nickname}
                onChange={handleContentChange}
              />
            </label>
            <label className="flex flex-col text-sm gap-1 col-span-2 text-purple-300">
              About Me
              <textarea
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 min-h-24 text-purple-100 focus:outline-none focus:border-purple-500/50 resize-none"
                name="about_me"
                value={contentForm.about_me}
                onChange={handleContentChange}
              />
            </label>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-purple-400">{contentStatus}</span>
            <button
              type="submit"
              className="bg-purple-600 hover:bg-purple-500 text-white rounded-md px-4 py-2 disabled:opacity-50 shadow-[0_0_10px_rgba(168,85,247,0.3)] transition cursor-pointer"
              disabled={isSavingContent}
            >
              {isSavingContent ? "Saving..." : "Save Content"}
            </button>
          </div>
        </form>
      ) : null}

      {!isLoading && !loadingError ? (
        <form onSubmit={handleSaveVisibility} className="border border-purple-500/30 rounded-lg bg-[#1a1a2e] text-purple-100 w-full p-5 flex flex-col gap-4">
          <h2 className="text-xl font-bold">Visibility & Privacy</h2>

          <div className="grid grid-cols-2 gap-4">
            <label className="flex flex-col text-sm gap-1 col-span-2 text-purple-300">
              Profile Type
              <select
                className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer"
                name="profile_type"
                value={visibilityForm.profile_type}
                onChange={handleVisibilityChange}
              >
                <option value="public" className="bg-[#1a1a2e]">Public</option>
                <option value="private" className="bg-[#1a1a2e]">Private</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Email
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="email_vis" value={visibilityForm.email_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Birthday Date
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="birthday_date_vis" value={visibilityForm.birthday_date_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Relationship Status
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="relationship_status_vis" value={visibilityForm.relationship_status_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Employed At
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="employed_at_vis" value={visibilityForm.employed_at_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Phone Number
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="phone_number_vis" value={visibilityForm.phone_number_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              About Me
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="about_me_vis" value={visibilityForm.about_me_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Nickname
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="nickname_vis" value={visibilityForm.nickname_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>

            <label className="flex flex-col text-sm gap-1 text-purple-300">
              Follow List
              <select className="bg-[#0d0d1a] border border-purple-500/30 rounded-md px-3 py-2 text-purple-100 focus:outline-none focus:border-purple-500/50 cursor-pointer" name="follow_vis" value={visibilityForm.follow_vis} onChange={handleVisibilityChange}>
                <option value="visible" className="bg-[#1a1a2e]">Visible</option>
                <option value="hidden" className="bg-[#1a1a2e]">Hidden</option>
              </select>
            </label>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-sm text-purple-400">{visibilityStatus}</span>
            <button
              type="submit"
              className="bg-purple-600 hover:bg-purple-500 text-white rounded-md px-4 py-2 disabled:opacity-50 shadow-[0_0_10px_rgba(168,85,247,0.3)] transition cursor-pointer"
              disabled={isSavingVisibility}
            >
              {isSavingVisibility ? "Saving..." : "Save Visibility"}
            </button>
          </div>
        </form>
      ) : null}
    </div>
  );
}
