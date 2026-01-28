"use client";


import FormContainer from "../../../components/ui/FormContainer";
import Button from "../../../components/ui/Button";
import Logo from "../../../components/ui/Logo";
import Input from "../../../components/ui/Input";
import Image from "next/image";


const RegisterPage = () => {

const handleRegister = async (e) => {
  e.preventDefault();
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;
  const confirmPassword = document.getElementById("confirm-password").value;
  const firstname = document.getElementById("firstname").value;
  const lastname = document.getElementById("lastname").value;
  const dateofbirth = document.getElementById("dateofbirth").value;
  const nickname = document.getElementById("nickname").value;
  const aboutme = document.getElementById("aboutme").value;
  const avatarInput = document.getElementById("avatar");
  const avatarFile = avatarInput.files[0];
  
    if (!email || !password || !confirmPassword || !firstname || !lastname || !dateofbirth) {
    alert("Please fill in all required fields");
    return;
  }

  if (password !== confirmPassword) {
    alert("Passwords do not match!");
    return;
  }

  const userData = {
    email: email,
    first_name: firstname,
    password: password,
    last_name: lastname,
    date_of_birth: dateofbirth,
    nickname: nickname,
    about_me: aboutme,
    avatar: avatarFile ? avatarFile.name : null,
  };
  
  try {
    const response = await fetch("http://localhost:8080/api/signup", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(userData),
    });

    const data = await response.json();
    
    if (response.ok) {
      console.log("Registration successful:", data);
    
      window.location.href = "/login";
    } else {
      console.error("Registration failed:", data);
      alert(data.error || "Registration failed");
    }
  } catch (error) {
    console.error("Error:", error);
    alert("Failed to connect to server");
  }
}
  return (
    <FormContainer onSubmit={handleRegister} encType="multipart/form-data">
      <Logo
        title="Create your account"
        subtitle="Join our community today!"
        variant="blur"
      />

      <Input
        label="Email"
        icon="/email_icon.svg"
        id="email"
        type="email"
        placeholder="your.email@example.com"
        required
        className="mb-4"
      />

      <div className="mb-6 flex justify-center gap-4 relative mt-8">
        <Input
          label="Password"
          icon="/lock_icon.svg"
          id="password"
          type="password"
          placeholder="******************"
          required
        />
        <Input
          label="Confirm Password"
          icon="/lock_icon.svg"
          id="confirm-password"
          type="password"
          placeholder="******************"
          required
        />
      </div>

      <div className="mb-10 flex justify-center gap-4 relative mt-8">
        <Input
          label="First Name"
          icon="/name_icon.svg"
          id="firstname"
          type="text"
          placeholder="Your First Name"
          required
        />
        <Input
          label="Last Name"
          icon="/name_icon.svg"
          id="lastname"
          type="text"
          placeholder="Your Last Name"
          required
        />
      </div>

      <Input
        label="Date of Birth"
        icon="/calendar_icon.svg"
        id="dateofbirth"
        type="date"
        required
        className="mb-14"
      />

      <div className="mb-20 flex justify-left relative gap-4">
        <label
          className="absolute left-0 bottom-1 text-sm mb-1"
          htmlFor="avatar"
        >
          Avatar/Profile Picture{" "}
          <span className="text-gray-500">(Optional)</span>
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
          type="file"
        />
      </div>

      <Input
        label="Nickname"
        icon="/nickname_icon.svg"
        id="nickname"
        type="text"
        placeholder="Your Nickname"
        optional
        className="mb-8"
      />

      <div className="mb-14 relative">
        <label className="label-custom h-16" htmlFor="aboutme">
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
          id="aboutme"
          type="text"
          placeholder="Tell us about yourself..."
        />
      </div>

      <Button type="submit">Create Account</Button>
    </FormContainer>
  );
};

export default RegisterPage;
