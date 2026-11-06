"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import type { Post, Profile, User } from "@/lib/types";
import Avatar from "./Avatar";

interface Props {
  groupId?: string; // when set, creates a post inside a group
  onCreated: (post: Post) => void;
}

// CreatePost is the post composer. Outside a group it offers the three privacy
// levels (and, for "private", a clear checkbox picker of which followers may
// see the post); inside a group privacy is implicit.
export default function CreatePost({ groupId, onCreated }: Props) {
  const { user } = useAuth();
  const [content, setContent] = useState("");
  const [privacy, setPrivacy] = useState("public");
  const [image, setImage] = useState<File | null>(null);
  const [allowed, setAllowed] = useState<string[]>([]);
  const [followers, setFollowers] = useState<User[]>([]);
  const [loadedFollowers, setLoadedFollowers] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // Load the current user's followers once, when "private" is first chosen.
  useEffect(() => {
    if (!groupId && privacy === "private" && !loadedFollowers && user) {
      api
        .get<Profile>(`/api/users/${user.id}`)
        .then((p) => setFollowers(p.followers ?? []))
        .catch(() => setFollowers([]))
        .finally(() => setLoadedFollowers(true));
    }
  }, [privacy, groupId, loadedFollowers, user]);

  const toggle = (id: string) =>
    setAllowed((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );

  const isPrivate = !groupId && privacy === "private";

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;
    if (isPrivate && allowed.length === 0) {
      setError("Select at least one follower who can see this private post.");
      return;
    }
    setError("");
    setLoading(true);
    const form = new FormData();
    form.set("content", content);
    if (groupId) form.set("groupId", groupId);
    else form.set("privacy", privacy);
    if (image) form.set("image", image);
    if (isPrivate) allowed.forEach((id) => form.append("allowedUsers", id));

    try {
      const post = await api.form<Post>("/api/posts", form);
      onCreated(post);
      setContent("");
      setImage(null);
      setAllowed([]);
      setPrivacy("public");
      (e.target as HTMLFormElement).reset();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to post");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <form onSubmit={submit}>
        <textarea
          placeholder={groupId ? "Share something with the group…" : "What's on your mind?"}
          value={content}
          onChange={(e) => setContent(e.target.value)}
        />
        <div className="row" style={{ marginTop: 10, flexWrap: "wrap" }}>
          <input
            type="file"
            accept="image/jpeg,image/png,image/gif"
            style={{ width: "auto" }}
            onChange={(e) => setImage(e.target.files?.[0] ?? null)}
          />
          {!groupId && (
            <select
              value={privacy}
              onChange={(e) => setPrivacy(e.target.value)}
              style={{ width: "auto" }}
            >
              <option value="public">Public — everyone</option>
              <option value="almost_private">Followers only</option>
              <option value="private">Selected followers…</option>
            </select>
          )}
          <button type="submit" disabled={loading} style={{ marginLeft: "auto" }}>
            Post
          </button>
        </div>

        {isPrivate && (
          <div className="audience">
            <div className="spread" style={{ marginBottom: 6 }}>
              <label style={{ margin: 0 }}>Who can see this post?</label>
              <span className="muted">Selected: {allowed.length}</span>
            </div>

            {!loadedFollowers ? (
              <div className="muted">Loading followers…</div>
            ) : followers.length === 0 ? (
              <div className="muted">
                You have no followers yet — only people who follow you can be
                chosen for a private post.
              </div>
            ) : (
              <div className="audience-list">
                {followers.map((f) => {
                  const checked = allowed.includes(f.id);
                  return (
                    <label
                      key={f.id}
                      className={"audience-item" + (checked ? " checked" : "")}
                    >
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => toggle(f.id)}
                        style={{ width: "auto" }}
                      />
                      <Avatar
                        name={`${f.firstName} ${f.lastName}`}
                        avatar={f.avatar}
                        size={28}
                      />
                      <span>
                        {f.firstName} {f.lastName}
                      </span>
                      {checked && <span className="check">✓</span>}
                    </label>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {error && <div className="error">{error}</div>}
      </form>
    </div>
  );
}
