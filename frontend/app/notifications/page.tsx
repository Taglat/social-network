"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useSocket } from "@/lib/socket";
import type { Notification } from "@/lib/types";

export default function NotificationsPage() {
  const { notifications, refreshNotifications, markAllNotificationsRead } =
    useSocket();
  // Ids of notifications already responded to (hides their action buttons).
  const [responded, setResponded] = useState<Set<string>>(new Set());

  // Mark everything read when the page is opened (clears the badge), but keep
  // actionable items usable.
  useEffect(() => {
    markAllNotificationsRead();
  }, [markAllNotificationsRead]);

  const respond = async (n: Notification, action: "accept" | "decline") => {
    try {
      if (n.type === "follow_request" && n.actorId) {
        await api.post(`/api/follow-requests/${n.actorId}/respond`, { action });
      } else if (n.type === "group_invite" && n.entityId) {
        await api.post(`/api/groups/${n.entityId}/invitation/respond`, { action });
      } else if (n.type === "group_join_request" && n.entityId && n.actorId) {
        await api.post(
          `/api/groups/${n.entityId}/requests/${n.actorId}/respond`,
          { action }
        );
      }
      setResponded((prev) => new Set(prev).add(n.id));
    } catch (e) {
      // e.g. the request was already handled elsewhere (stale notification).
      alert(e instanceof Error ? e.message : "Action failed");
      setResponded((prev) => new Set(prev).add(n.id));
    } finally {
      refreshNotifications();
    }
  };

  return (
    <div className="container">
      <h2>Notifications</h2>
      {notifications.length === 0 && (
        <div className="card muted">No notifications yet.</div>
      )}
      {notifications.map((n) => (
        <div
          key={n.id}
          className="card"
          style={{ borderLeft: n.isRead ? "none" : "3px solid var(--primary)" }}
        >
          <div className="spread">
            <div>
              <div>{n.message}</div>
              <div className="muted">
                {new Date(n.createdAt).toLocaleString()}
              </div>
            </div>
            {n.actionable && !responded.has(n.id) && (
              <div className="row">
                <button className="small" onClick={() => respond(n, "accept")}>
                  Accept
                </button>
                <button
                  className="small secondary"
                  onClick={() => respond(n, "decline")}
                >
                  Decline
                </button>
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
