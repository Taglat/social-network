"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/lib/auth";
import Nav from "./Nav";

const PUBLIC_ROUTES = ["/login", "/register"];

// AppShell renders the nav bar and guards protected routes: unauthenticated
// users are redirected to /login.
export default function AppShell({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const isPublic = PUBLIC_ROUTES.includes(pathname);

  useEffect(() => {
    if (!loading && !user && !isPublic) {
      router.replace("/login");
    }
    if (!loading && user && isPublic) {
      router.replace("/");
    }
  }, [loading, user, isPublic, router]);

  if (loading) {
    return <div className="center-page muted">Loading…</div>;
  }
  if (!user && !isPublic) {
    return <div className="center-page muted">Redirecting…</div>;
  }

  return (
    <>
      <Nav />
      {children}
    </>
  );
}
