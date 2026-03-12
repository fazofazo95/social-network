"use client";

import FormContainer from "../../../components/ui/FormContainer";
import Button from "../../../components/ui/Button";
import Logo from "../../../components/ui/Logo";
import Input from "../../../components/ui/Input";
import { signupUser } from "src/lib/services/auth";
import { useToast } from "src/components/ui/Toast";

const RegisterPage = () => {
  const toast = useToast();

  const handleRegister = async (e) => {
    e.preventDefault();

    const formData = new FormData(e.target);
  const password = formData.get("password");
  const confirmPassword = formData.get("confirm-password");


  if (
    !formData.get("email") ||
    !password ||
    !confirmPassword ||
    !formData.get("firstname") ||
    !formData.get("lastname") ||
    !formData.get("date_of_birth")
  ) {
    toast.warning("Please fill in all required fields");
    return;
  }

  if (password !== confirmPassword) {
    toast.warning("Passwords do not match!");
    return;
  }


  formData.delete("confirm-password");
 
  formData.append("username", formData.get("firstname") + " " + formData.get("lastname"));

    try {
      const data = await signupUser(formData);
      console.log("Registration successful:", data);
      window.location.href = "/login";
    } catch (error) {
      console.error("Registration failed:", error);
      toast.error(error?.message || "Registration failed");
    }
  };
  return (
    <FormContainer onSubmit={handleRegister} encType="multipart/form-data">
      <Logo
        title="Create your account"
        subtitle="Join our community today!"
        variant="blur"
      />

      <Input
        label="Email"
        emoji="📧"
        id="email"
        name="email"
        type="email"
        placeholder="your.email@example.com"
        required
        className="mb-4"
      />

      <div className="mb-6 flex justify-center gap-4 relative mt-8">
        <Input
          label="Password"
          emoji="🔒"
          id="password"
          name="password"
          type="password"
          placeholder="******************"
          required
        />
        <Input
          label="Confirm Password"
          emoji="🔒"
          id="confirm-password"
          name="confirm-password"
          type="password"
          placeholder="******************"
          required
        />
      </div>

      <div className="mb-10 flex justify-center gap-4 relative mt-8">
        <Input
          label="First Name"
          emoji="👤"
          id="firstname"
          name="firstname"
          type="text"
          placeholder="Your First Name"
          required
        />
        <Input
          label="Last Name"
          emoji="👤"
          id="lastname"
          name="lastname"
          type="text"
          placeholder="Your Last Name"
          required
        />
      </div>

      <Input
        label="Date of Birth"
        emoji="📅"
        id="date_of_birth"
        name="date_of_birth"
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
          <span className="text-purple-400">(Optional)</span>
        </label>

        <span className="absolute left-2.5 top-2 text-2xl leading-none">🖼️</span>

        <input
          className="absolute top-2 left-14 bg-[#0d0d1a] border border-purple-500/30 w-3/4 rounded-xl text-sm pl-2 text-purple-200 file:bg-transparent file:text-purple-300 file:border-0 file:rounded-md file:px-2 file:py-0.5 file:mr-2 file:cursor-pointer"
          id="avatar"
          name="avatar"
          type="file"
        />
      </div>

      <Input
        label="Nickname"
        emoji="🏷️"
        id="nickname"
        name="nickname"
        type="text"
        placeholder="Your Nickname"
        optional
        className="mb-8"
      />

      <div className="mb-14 relative">
        <label className="label-custom h-16" htmlFor="about_me">
          About Me <span className="text-purple-400">(Optional)</span>
        </label>

        <span className="absolute left-2.5 bottom-[1.1rem] text-lg leading-none">📝</span>
        <input
          className="border border-purple-500/30 rounded-md w-full py-2 pb-12 pl-10 pr-2 bg-[#0d0d1a] text-purple-200 placeholder-purple-400/40 focus:border-purple-500/50 focus:outline-none"
          id="about_me"
          name="about_me"
          type="text"
          placeholder="Tell us about yourself..."
        />
      </div>

      <Button type="submit">Create Account</Button>
    </FormContainer>
  );
};

export default RegisterPage;
