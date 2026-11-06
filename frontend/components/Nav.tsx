"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { useSocket } from "@/lib/socket";
import Avatar from "./Avatar";

// Nav is the persistent navigation bar shown on every authenticated page. It
// surfaces the notification badge and a *separate* message badge (the audit
// requires new notifications to be displayed differently from new messages).
export default function Nav() {
  const { user, logout } = useAuth();
  const { unreadNotifications, unreadMessages, connected } = useSocket();
  const router = useRouter();

  if (!user) return null;

  const handleLogout = async () => {
    await logout();
    router.push("/login");
  };

  return (
    <nav className="nav">
      <div className="nav-inner">
        <Link href="/" className="brand">
          Social
        </Link>
        <div className="links">
          <Link href="/">Feed</Link>
          <Link href="/groups">Groups</Link>
          <Link href="/chat">
            Chat
            {unreadMessages > 0 && <span className="badge">{unreadMessages}</span>}
          </Link>
          <Link href="/notifications">
            Notifications
            {unreadNotifications > 0 && (
              <span className="badge">{unreadNotifications}</span>
            )}
          </Link>
        </div>
        <span
          className={connected ? "dot-online" : "dot-offline"}
          title={connected ? "Realtime connected" : "Reconnecting…"}
        />
        <Link href={`/profile/${user.id}`} className="row" style={{ gap: 8 }}>
          <Avatar name={`${user.firstName} ${user.lastName}`} avatar={user.avatar} size={32} />
        </Link>
        <button className="secondary small" onClick={handleLogout}>
          Logout
        </button>
      </div>
    </nav>
  );
}
