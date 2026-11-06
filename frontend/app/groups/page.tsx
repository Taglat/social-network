"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { Group } from "@/lib/types";

export default function GroupsPage() {
  const [groups, setGroups] = useState<Group[]>([]);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");

  const load = () =>
    api.get<Group[]>("/api/groups").then(setGroups).catch(() => {});

  useEffect(() => {
    load();
  }, []);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !description.trim()) return;
    try {
      await api.post<Group>("/api/groups", { title, description });
      setTitle("");
      setDescription("");
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to create group");
    } finally {
      load();
    }
  };

  const requestJoin = async (id: string) => {
    try {
      await api.post(`/api/groups/${id}/request`);
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to send request");
    } finally {
      load();
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h3 style={{ marginTop: 0 }}>Create a group</h3>
        <form onSubmit={create} className="col">
          <input
            placeholder="Group title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />
          <textarea
            placeholder="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          <button type="submit" style={{ alignSelf: "flex-start" }}>
            Create group
          </button>
        </form>
      </div>

      <h2>Browse groups</h2>
      {groups.length === 0 && <div className="card muted">No groups yet.</div>}
      {groups.map((g) => (
        <div key={g.id} className="card spread">
          <div>
            <Link href={`/groups/${g.id}`} style={{ fontWeight: 600, fontSize: 16 }}>
              {g.title}
            </Link>
            <div className="muted">{g.description}</div>
            <div className="muted">{g.memberCount} members</div>
          </div>
          <div>
            {g.membership === "member" ? (
              <Link href={`/groups/${g.id}`}>
                <button className="small">Open</button>
              </Link>
            ) : g.membership === "requested" ? (
              <button className="small secondary" disabled>
                Requested
              </button>
            ) : g.membership === "invited" ? (
              <Link href={`/groups/${g.id}`}>
                <button className="small">View invite</button>
              </Link>
            ) : (
              <button className="small" onClick={() => requestJoin(g.id)}>
                Request to join
              </button>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
