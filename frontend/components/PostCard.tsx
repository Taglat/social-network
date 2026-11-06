"use client";

import { useState } from "react";
import Link from "next/link";
import { api, mediaUrl } from "@/lib/api";
import type { Comment, Post } from "@/lib/types";
import Avatar from "./Avatar";

const PRIVACY_LABELS: Record<string, string> = {
  public: "Public",
  almost_private: "Followers",
  private: "Private",
  group: "Group",
};

export default function PostCard({ post }: { post: Post }) {
  const [comments, setComments] = useState<Comment[] | null>(null);
  const [showComments, setShowComments] = useState(false);
  const [text, setText] = useState("");
  const [image, setImage] = useState<File | null>(null);
  const [count, setCount] = useState(post.commentCount);

  const toggleComments = async () => {
    const next = !showComments;
    setShowComments(next);
    if (next && comments === null) {
      try {
        setComments(await api.get<Comment[]>(`/api/posts/${post.id}/comments`));
      } catch {
        setComments([]);
      }
    }
  };

  const addComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim() && !image) return;
    const form = new FormData();
    form.set("content", text);
    if (image) form.set("image", image);
    try {
      const updated = await api.form<Comment[]>(
        `/api/posts/${post.id}/comments`,
        form
      );
      setComments(updated);
      setCount(updated.length);
      setText("");
      setImage(null);
      (e.target as HTMLFormElement).reset();
    } catch {
      /* ignore */
    }
  };

  return (
    <div className="card">
      <div className="row">
        <Avatar name={post.authorName} avatar={post.authorAvatar} />
        <div>
          <Link href={`/profile/${post.userId}`} style={{ fontWeight: 600, color: "var(--text)" }}>
            {post.authorName}
          </Link>
          <div className="muted">
            {new Date(post.createdAt).toLocaleString()} ·{" "}
            <span className="tag">{PRIVACY_LABELS[post.privacy] ?? post.privacy}</span>
          </div>
        </div>
      </div>
      <p style={{ whiteSpace: "pre-wrap" }}>{post.content}</p>
      {post.image && (
        // eslint-disable-next-line @next/next/no-img-element
        <img className="post-image" src={mediaUrl(post.image)} alt="post" />
      )}
      <div style={{ marginTop: 8 }}>
        <button className="secondary small" onClick={toggleComments}>
          💬 {count} comment{count === 1 ? "" : "s"}
        </button>
      </div>

      {showComments && (
        <div style={{ marginTop: 12 }}>
          {comments?.map((c) => (
            <div key={c.id} className="row" style={{ alignItems: "flex-start", marginBottom: 10 }}>
              <Avatar name={c.authorName} avatar={c.authorAvatar} size={30} />
              <div style={{ background: "#f0f2f5", borderRadius: 12, padding: "6px 10px", flex: 1 }}>
                <div style={{ fontWeight: 600, fontSize: 13 }}>{c.authorName}</div>
                {c.content && <div>{c.content}</div>}
                {c.image && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img className="post-image" src={mediaUrl(c.image)} alt="comment" style={{ maxWidth: 200 }} />
                )}
              </div>
            </div>
          ))}
          <form onSubmit={addComment} className="col" style={{ gap: 6 }}>
            <input
              placeholder="Write a comment…"
              value={text}
              onChange={(e) => setText(e.target.value)}
            />
            <div className="row">
              <input
                type="file"
                accept="image/jpeg,image/png,image/gif"
                style={{ width: "auto" }}
                onChange={(e) => setImage(e.target.files?.[0] ?? null)}
              />
              <button type="submit" className="small">
                Comment
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
