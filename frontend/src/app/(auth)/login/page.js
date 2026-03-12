"use client";


import FormContainer from "../../../components/ui/FormContainer";
import Link from "next/link";
import Button from "../../../components/ui/Button";
import Logo from "../../../components/ui/Logo";
import Input from "../../../components/ui/Input";
import { useRouter } from "next/navigation";
import { loginUser } from "src/lib/services/auth";
import { useToast } from "src/components/ui/Toast";

const LoginPage = () => {
  const router = useRouter();
  const toast = useToast();
  
  const handleLogin = async () => {
  const email = document.getElementById("email").value;
  const password = document.getElementById("password").value;

  const userData = {
    email: email,
    password: password,
  };
  
  try {
    await loginUser(userData);
    toast.success("Login successful!");
    router.push("/");
  } catch (error) {
    console.error("Login failed:", error);
    toast.error(error?.message || "Login failed");
  }
}
  return (

      <FormContainer>
        <Logo
          title="Welcome back!"
          subtitle="Login to continue to your account"
        />

        <Input
          label="Email Address"
          emoji="📧"
          id="email"
          type="email"
          placeholder="your.email@example.com"
          className="mb-4"
        />

        <Input
          label="Password"
          emoji="🔒"
          id="password"
          type="password"
          placeholder="******************"
          className="mb-6 mt-8"
        />

        <Button type="button" onClick={handleLogin}>Sign In</Button>

        <div className="flex items-center justify-center mt-4 gap-6">
          <p className="border-t border-purple-500/30 w-1/2"></p>
          <p className="text-purple-400">Or</p>
          <p className="border-t border-purple-500/30 w-1/2"></p>
        </div>

        <div className="mt-5 text-center">
          <p className="text-purple-400 text-lg">
            Don't have an account?{" "}
            <Link
              href="/register"
              className="font-bold text-pink-400 hover:text-pink-300"
            >
              Sign Up
            </Link>
          </p>
        </div>
      </FormContainer>

  );
};

export default LoginPage;
