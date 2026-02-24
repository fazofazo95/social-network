"use client";



import { useState } from "react";
import { createGroup } from "src/lib/services/group";
import Button from "src/components/ui/Button";
import FormContainer from "src/components/ui/FormContainer";
import Logo from "src/components/ui/Logo";
import Input from "src/components/ui/Input";
const CreateGroupPage = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleCreateGroup = async (e) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    const formData = new FormData(e.target);
    if (!formData.get("name") || !formData.get("description")) {
      setError("Please fill in all required fields");
      setLoading(false);
      return;
    }
    try {
      await createGroup(formData);
      window.location.href = "/(dashboard)/groups";
    } catch (err) {
      setError(err?.message || "Failed to create group");
    } finally {
      setLoading(false);
    }
  };

  return (
<FormContainer onSubmit={handleCreateGroup} encType="multipart/form-data">
      <Logo
        title="Create your account"
        subtitle="Join our community today!"
        variant="blur"
      />

      <Input
        label="Group Name"
        icon="/purple_groups_icon.svg"
        id="name"
        name="name"
        type="text"
        placeholder="Enter group name"
        required
        className="mb-8"
      />



      <div className="mb-8">
        <label htmlFor="group_picture" className="block text-purple-200 font-semibold mb-1">Group Picture (optional)</label>
        <input
          className="w-full rounded-lg p-2 bg-[#2a0055] text-white border border-purple-700"
          id="group_picture"
          name="group_picture"
          type="file"
          accept="image/*"
        />
      </div>



      <Input
        label="Description"
        icon="/aboutme_icon.svg"
        id="description"
        name="description"
        type="text"
        placeholder="Tell us about the group..."
        required
        textarea
        rows={3}
        className="mb-8"
      />

      <div className="flex flex-row gap-4 mb-8">
        <div className="flex-1">
          <label htmlFor="visibility" className="block text-purple-200 font-semibold mb-1">Visibility</label>
          <select name="visibility" id="visibility" className="w-full rounded-lg p-2 bg-[#2a0055] text-white border border-purple-700">
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
        </div>
        <div className="flex-1">
          <label htmlFor="join_mode" className="block text-purple-200 font-semibold mb-1">Join Mode</label>
          <select name="join_mode" id="join_mode" className="w-full rounded-lg p-2 bg-[#2a0055] text-white border border-purple-700">
            <option value="auto">Auto</option>
            <option value="request">Request</option>
            <option value="invite">Invite</option>
            <option value="request_and_invite">Request & Invite</option>
          </select>
        </div>
      </div>

      <Button type="submit">Create Group</Button>
    </FormContainer>
  );
};

export default CreateGroupPage;
