"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { verifySession } from "src/lib/services/auth";

const AuthGate = ({ children, requireAuth = true }) => {
  const router = useRouter();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    let isMounted = true;

    const checkAuth = async () => {
      try {
        await verifySession();

        if (!requireAuth) {
          router.replace("/");
          return;
        }
      } catch (_error) {
        if (requireAuth) {
          router.replace("/login");
          return;
        }
      } finally {
        if (isMounted) {
          setIsChecking(false);
        }
      }
    };

    checkAuth();

    return () => {
      isMounted = false;
    };
  }, [requireAuth, router]);

  if (isChecking) {
    return <div className="text-center mt-10">Loading...</div>;
  }

  return children;
};

export default AuthGate;
