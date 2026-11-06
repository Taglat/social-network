"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Post } from "@/lib/types";
import CreatePost from "@/components/CreatePost";
import PostCard from "@/components/PostCard";

export default function FeedPage() {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .get<Post[]>("/api/posts")
      .then((p) => setPosts(p || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="container">
      <CreatePost onCreated={(p) => setPosts((prev) => [p, ...prev])} />
      {loading && <p className="muted">Loading feed…</p>}
      {!loading && posts.length === 0 && (
        <div className="card muted">
          No posts yet. Create the first one, or follow people to see theirs.
        </div>
      )}
      {posts.map((p) => (
        <PostCard key={p.id} post={p} />
      ))}
    </div>
  );
}
