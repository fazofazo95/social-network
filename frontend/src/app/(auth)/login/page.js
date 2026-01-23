"use client";


import FormContainer from "../../../components/ui/FormContainer";
import Link from "next/link";
import Button from "../../../components/ui/Button";
import Logo from "../../../components/ui/Logo";
import Input from "../../../components/ui/Input";

const LoginPage = () => {
  return (

      <FormContainer>
        <Logo
          title="Welcome back!"
          subtitle="Login to continue to your account"
        />

        <Input
          label="Email Address"
          icon="/email_icon.svg"
          id="username"
          type="email"
          placeholder="your.email@example.com"
          className="mb-4"
        />

        <Input
          label="Password"
          icon="/lock_icon.svg"
          id="password"
          type="password"
          placeholder="******************"
          className="mb-6 mt-8"
        />

        <div className="mb-4 flex items-center justify-between">
          <label className="gap-2">
            <input type="checkbox" className="mr-2" />
            Remember Me
          </label>

          <Link
            href="/forgot-password"
            className="inline-block align-baseline font-bold text-sm text-pink-400 hover:text-pink-700"
          >
            Forgot Password?
          </Link>
        </div>

        <Button type="button">Sign In</Button>

        <div className="flex items-center justify-center mt-4 gap-6">
          <p className="border-t border-gray-400 w-1/2"></p>
          <p className="text-gray-500">Or</p>
          <p className="border-t border-gray-400 w-1/2"></p>
        </div>

        <div className="mt-5 text-center">
          <p className="text-gray-500 text-lg">
            Don't have an account?{" "}
            <Link
              href="/register"
              className="font-bold text-pink-400 hover:text-pink-700"
            >
              Sign Up
            </Link>
          </p>
        </div>
      </FormContainer>

  );
};

export default LoginPage;
