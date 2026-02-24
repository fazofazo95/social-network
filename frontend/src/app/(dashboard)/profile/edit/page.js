"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import FormContainer from "src/components/ui/FormContainer";
import Button from "src/components/ui/Button";
import Logo from "src/components/ui/Logo";
import Input from "src/components/ui/Input";
import { fetchUserData } from "src/lib/services/user";

const EditProfilePage = () => {
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState("");
	const [formValues, setFormValues] = useState({
		first_name: "",
		last_name: "",
		nickname: "",
		birthday_date: "",
		relationship_status: "",
		employed_at: "",
		phone_number: "",
		location: "",
		about_me: "",
	});

	function handleChange(event) {
		const { name, value } = event.target;
		setFormValues((prev) => ({
			...prev,
			[name]: value,
		}));
	}

	async function loadProfile() {
		setIsLoading(true);
		setError("");
		try {
			const profile = await fetchUserData("me");
			setFormValues({
				first_name: profile?.first_name || "",
				last_name: profile?.last_name || "",
				nickname: profile?.nickname || "",
				birthday_date: profile?.birthday_date || "",
				relationship_status: profile?.relationship_status || "",
				employed_at: profile?.employed_at || "",
				phone_number: profile?.phone_number || "",
				location: profile?.location || "",
				about_me: profile?.about_me || "",
			});
		} catch (loadError) {
			console.error("Failed to load profile for editing:", loadError);
			setError(loadError?.message || "Failed to load profile data.");
		} finally {
			setIsLoading(false);
		}
	}

	function handleSubmit(event) {
		event.preventDefault();
		alert("Profile update endpoint is not implemented in backend yet.");
	}

	useEffect(() => {
		loadProfile();
	}, []);

	if (isLoading) {
		return <div className="text-black">Loading profile form...</div>;
	}

	return (
		<FormContainer onSubmit={handleSubmit} className="max-w-2xl">
			<Logo
				title="Edit Profile"
				subtitle="Update your profile information"
				variant="blur"
			/>

			{error ? <p className="text-red-600 mb-4">{error}</p> : null}

			<div className="mb-6 flex justify-center gap-4 relative mt-8">
				<Input
					label="First Name"
					icon="/name_icon.svg"
					id="first_name"
					name="first_name"
					type="text"
					placeholder="Your First Name"
					value={formValues.first_name}
					onChange={handleChange}
					required
				/>
				<Input
					label="Last Name"
					icon="/name_icon.svg"
					id="last_name"
					name="last_name"
					type="text"
					placeholder="Your Last Name"
					value={formValues.last_name}
					onChange={handleChange}
					required
				/>
			</div>

			<Input
				label="Nickname"
				icon="/nickname_icon.svg"
				id="nickname"
				name="nickname"
				type="text"
				placeholder="Your Nickname"
				value={formValues.nickname}
				onChange={handleChange}
				optional
				className="mb-8"
			/>

			<div className="mb-6 flex justify-center gap-4 relative mt-8">
				<Input
					label="Date of Birth"
					icon="/calendar_icon.svg"
					id="birthday_date"
					name="birthday_date"
					type="date"
					value={formValues.birthday_date}
					onChange={handleChange}
					optional
				/>
				<Input
					label="Phone Number"
					icon="/name_icon.svg"
					id="phone_number"
					name="phone_number"
					type="text"
					placeholder="Your Phone Number"
					value={formValues.phone_number}
					onChange={handleChange}
					optional
				/>
			</div>

			<div className="mb-6 flex justify-center gap-4 relative mt-8">
				<Input
					label="Location"
					icon="/location_icon.svg"
					id="location"
					name="location"
					type="text"
					placeholder="Your Location"
					value={formValues.location}
					onChange={handleChange}
					optional
				/>
				<Input
					label="Relationship Status"
					icon="/profile_status_icon.svg"
					id="relationship_status"
					name="relationship_status"
					type="text"
					placeholder="Single, Married, etc."
					value={formValues.relationship_status}
					onChange={handleChange}
					optional
				/>
			</div>

			<Input
				label="Employed At"
				icon="/profile_status_icon.svg"
				id="employed_at"
				name="employed_at"
				type="text"
				placeholder="Company or Workplace"
				value={formValues.employed_at}
				onChange={handleChange}
				optional
				className="mb-14"
			/>

			<div className="mb-20 flex justify-left relative gap-4">
				<label className="absolute left-0 bottom-1 text-sm mb-1" htmlFor="avatar">
					Avatar/Profile Picture <span className="text-gray-500">(Optional)</span>
				</label>

				<Image
					src="/image_icon.svg"
					alt="Image Icon"
					width={40}
					height={40}
					className="absolute left-2 top-0 bg-gray-200 py-2 px-2 rounded-(--rounded-full)"
				/>

				<input
					className="absolute top-2 left-16 bg-gray-200 w-3/4 rounded-xl text-sm pl-1.5 text-black"
					id="avatar"
					name="avatar"
					type="file"
					accept="image/*"
				/>
			</div>

			<div className="mb-14 relative">
				<label className="label-custom h-16" htmlFor="about_me">
					About Me <span className="text-gray-500">(Optional)</span>
				</label>

				<Image
					src="/aboutme_icon.svg"
					alt="About Me Icon"
					width={20}
					height={20}
					className="absolute left-2 top-3"
				/>
				<input
					className="border rounded-md w-full py-2 pb-12 pl-8 pr-2 bg-white text-gray-600"
					id="about_me"
					name="about_me"
					type="text"
					placeholder="Tell us about yourself..."
					value={formValues.about_me}
					onChange={handleChange}
				/>
			</div>

			<Button type="submit">Save Changes</Button>
		</FormContainer>
	);
};

export default EditProfilePage;
