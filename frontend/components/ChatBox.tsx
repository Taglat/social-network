"use client";

import { useEffect, useRef, useState } from "react";
import { useAuth } from "@/lib/auth";
import { useSocket } from "@/lib/socket";
import type { Message } from "@/lib/types";
import EmojiPicker from "./EmojiPicker";

interface Props {
  title: string;
  loadHistory: () => Promise<Message[]>;
  send: (content: string) => void;
  // accept decides whether an incoming live message belongs to this conversation.
  accept: (msg: Message) => boolean;
}

// ChatBox is the shared conversation view: message history, live updates over
// the websocket, an emoji picker and an interactive (as-you-type) search box.
export default function ChatBox({ title, loadHistory, send, accept }: Props) {
  const { user } = useAuth();
  const { onMessage } = useSocket();
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState("");
  const [search, setSearch] = useState("");
  const [showEmoji, setShowEmoji] = useState(false);
  const endRef = useRef<HTMLDivElement>(null);

  // Load history whenever the conversation changes.
  useEffect(() => {
    let active = true;
    loadHistory()
      .then((m) => active && setMessages(m || []))
      .catch(() => active && setMessages([]));
    return () => {
      active = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [title]);

  // Subscribe to live messages for this conversation.
  useEffect(() => {
    return onMessage((msg) => {
      if (accept(msg)) setMessages((prev) => [...prev, msg]);
    });
  }, [onMessage, accept]);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    const content = text.trim();
    if (!content) return;
    send(content);
    setText("");
    setShowEmoji(false);
  };

  const visible = search
    ? messages.filter((m) =>
        m.content.toLowerCase().includes(search.toLowerCase())
      )
    : messages;

  return (
    <div className="chat-main">
      <div className="spread" style={{ padding: "10px 14px", borderBottom: "1px solid var(--border)" }}>
        <strong>{title}</strong>
        <input
          placeholder="Search messages…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{ width: 200 }}
        />
      </div>
      <div className="chat-messages">
        {visible.length === 0 && (
          <div className="muted">
            {search ? "No messages match your search." : "No messages yet."}
          </div>
        )}
        {visible.map((m) => (
          <div
            key={m.id}
            className={"bubble" + (m.senderId === user?.id ? " mine" : "")}
          >
            {m.senderId !== user?.id && (
              <div style={{ fontSize: 12, fontWeight: 700, opacity: 0.8 }}>
                {m.senderName}
              </div>
            )}
            {m.content}
          </div>
        ))}
        <div ref={endRef} />
      </div>
      <form className="chat-input" onSubmit={submit}>
        <button
          type="button"
          className="secondary"
          onClick={() => setShowEmoji((s) => !s)}
        >
          😊
        </button>
        {showEmoji && (
          <EmojiPicker onSelect={(e) => setText((t) => t + e)} />
        )}
        <input
          placeholder="Type a message…"
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
        <button type="submit">Send</button>
      </form>
    </div>
  );
}
