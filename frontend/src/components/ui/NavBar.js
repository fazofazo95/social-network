"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import IconButton from "./IconButton";
import SearchBar from "./SearchBar";
import { logoutUser } from "src/lib/services/auth";

const NavBar = () => {
  const router = useRouter();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const menuRef = useRef(null);

  useEffect(() => {
    const handleOutsideClick = (event) => {
      if (menuRef.current && !menuRef.current.contains(event.target)) {
        setIsMenuOpen(false);
      }
    };

    document.addEventListener("mousedown", handleOutsideClick);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
    };
  }, []);

  const handleLogout = async () => {
    try {
      setIsLoggingOut(true);
      await logoutUser();
    } catch (error) {
      console.error("Logout failed:", error?.message || error);
    } finally {
      setIsLoggingOut(false);
      setIsMenuOpen(false);
      router.replace("/login");
    }
  };

  return (
    <nav className="border w-full h-8 bg-white relative flex flex-row justify-between">
      <div className="flex items-center relative">
        <Image src="/logo_icon.svg" alt="Logo" width={25} height={25} />
        <div className="relative">
          <h1 className="text-purple-500 text-2xl font-semibold relative">
            Pulse
          </h1>
          <h1 className="absolute top-0 left-0 text-2xl text-purple-500 neon-glow">
            Pulse
          </h1>
        </div>
        
        <SearchBar placeholder="Search..." />
      </div>

      <div className="flex items-center gap-3 pr-4">
        <IconButton 
          icon="/notif-icon.svg" 
          alt="Notification Icon" 
          iconSize={17}
        />

        <div className="relative" ref={menuRef}>
          <IconButton
            icon="/profil_icon.svg"
            alt="Profile Icon"
            iconSize={20}
            onClick={() => setIsMenuOpen((prev) => !prev)}
          />

          {isMenuOpen && (
            <div className="absolute right-0 mt-2 w-36 rounded-lg border border-gray-200 bg-white shadow-custom z-50">
              <Link
                href="/profile"
                className="block px-3 py-2 text-sm text-black hover:bg-gray-100 rounded-t-lg"
                onClick={() => setIsMenuOpen(false)}
              >
                Profile
              </Link>
              <button
                type="button"
                className="w-full text-left px-3 py-2 text-sm text-black hover:bg-gray-100 rounded-b-lg disabled:opacity-50"
                onClick={handleLogout}
                disabled={isLoggingOut}
              >
                {isLoggingOut ? "Logging out..." : "Logout"}
              </button>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
};

export default NavBar;
