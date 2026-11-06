"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useSocket } from "@/lib/socket";
import type { Message, User } from "@/lib/types";
import Avatar from "@/components/Avatar";
import ChatBox from "@/components/ChatBox";

export default function ChatPage() {
  const { user } = useAuth();
  const { isOnline, sendPrivate, clearMessageBadge, setActiveConversation } =
    useSocket();
  const [partners, setPartners] = useState<User[]>([]);
  const [active, setActive] = useState<User | null>(null);

  useEffect(() => {
    api.get<User[]>("/api/chat/partners").then(setPartners).catch(() => {});
    clearMessageBadge();
  }, [clearMessageBadge]);

  // Mark the opened conversation as "being viewed" so its incoming messages do
  // not count as unread, and clear the badge when switching conversations.
  useEffect(() => {
    setActiveConversation(active ? `user:${active.id}` : null);
    if (active) clearMessageBadge();
    return () => setActiveConversation(null);
  }, [active, setActiveConversation, clearMessageBadge]);

  const loadHistory = useCallback(
    () => api.get<Message[]>(`/api/messages/${active?.id}`),
    [active?.id]
  );
  const accept = useCallback(
    (m: Message) =>
      m.groupId === null &&
      ((m.senderId === active?.id && m.receiverId === user?.id) ||
        (m.senderId === user?.id && m.receiverId === active?.id)),
    [active?.id, user?.id]
  );

  return (
    <div className="container-wide">
      <div className="chat-wrap">
        <div className="chat-list card" style={{ padding: 8 }}>
          <h3 style={{ padding: "0 8px" }}>Chats</h3>
          {partners.length === 0 && (
            <div className="muted" style={{ padding: 8 }}>
              Follow people (or be followed) to start chatting.
            </div>
          )}
          {partners.map((p) => (
            <div
              key={p.id}
              className={"list-item" + (active?.id === p.id ? " active" : "")}
              onClick={() => setActive(p)}
            >
              <Avatar name={`${p.firstName} ${p.lastName}`} avatar={p.avatar} size={36} />
              <div style={{ flex: 1 }}>
                <div>
                  {p.firstName} {p.lastName}
                </div>
                <div className="muted">
                  {isOnline(p.id) ? "Online" : "Offline"}
                </div>
              </div>
              <span className={isOnline(p.id) ? "dot-online" : "dot-offline"} />
            </div>
          ))}
        </div>

        {active ? (
          <ChatBox
            key={active.id}
            title={`${active.firstName} ${active.lastName}`}
            loadHistory={loadHistory}
            send={(content) => sendPrivate(active.id, content)}
            accept={accept}
          />
        ) : (
          <div className="chat-main center-page muted">
            Select a conversation to start chatting.
          </div>
        )}
      </div>
    </div>
  );
}
