"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import type { Profile, User } from "@/lib/types";
import Avatar from "@/components/Avatar";
import PostCard from "@/components/PostCard";
import ConfirmDialog from "@/components/ConfirmDialog";

export default function ProfilePage() {
  const { id } = useParams<{ id: string }>();
  const { setUser } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [error, setError] = useState("");
  const [confirm, setConfirm] = useState<null | "unfollow" | "privacy">(null);

  const load = useCallback(() => {
    api
      .get<Profile>(`/api/users/${id}`)
      .then(setProfile)
      .catch((e) => setError(e.message));
  }, [id]);

  useEffect(load, [load]);

  if (error) return <div className="container card error">{error}</div>;
  if (!profile) return <div className="container muted">Loading…</div>;

  const { user } = profile;
  const fullName = `${user.firstName} ${user.lastName}`;
  // Defensive defaults: the lists should always be arrays, but guard against null.
  const followers = profile.followers ?? [];
  const following = profile.following ?? [];
  const posts = profile.posts ?? [];

  const follow = async () => {
    try {
      await api.post(`/api/users/${user.id}/follow`);
    } catch (e) {
      alert(e instanceof Error ? e.message : "Action failed");
    } finally {
      load();
    }
  };
  const doUnfollow = async () => {
    setConfirm(null);
    try {
      await api.post(`/api/users/${user.id}/unfollow`);
    } catch (e) {
      alert(e instanceof Error ? e.message : "Action failed");
    } finally {
      load();
    }
  };
  const togglePrivacy = async () => {
    setConfirm(null);
    try {
      const updated = await api.put<User>("/api/profile/privacy", {
        isPublic: !user.isPublic,
      });
      setUser(updated);
    } catch (e) {
      alert(e instanceof Error ? e.message : "Action failed");
    } finally {
      load();
    }
  };

  return (
    <div className="container">
      <div className="card">
        <div className="row" style={{ alignItems: "flex-start" }}>
          <Avatar name={fullName} avatar={user.avatar} size={80} />
          <div style={{ flex: 1 }}>
            <div className="spread">
              <h2 style={{ margin: 0 }}>
                {fullName}{" "}
                <span className={profile.online ? "dot-online" : "dot-offline"} />
              </h2>
              <div className="row">
                {profile.isSelf ? (
                  <button className="secondary" onClick={() => setConfirm("privacy")}>
                    Make {user.isPublic ? "private" : "public"}
                  </button>
                ) : profile.isFollowing ? (
                  <button className="secondary" onClick={() => setConfirm("unfollow")}>
                    Unfollow
                  </button>
                ) : profile.requestSent ? (
                  <button className="secondary" disabled>
                    Request sent
                  </button>
                ) : (
                  <button onClick={follow}>
                    {user.isPublic ? "Follow" : "Request to follow"}
                  </button>
                )}
              </div>
            </div>
            {user.nickname && <div className="muted">@{user.nickname}</div>}
            <div className="row" style={{ marginTop: 6 }}>
              <span className="tag">{user.isPublic ? "Public" : "Private"} profile</span>
            </div>
          </div>
        </div>

        {profile.canViewFull ? (
          <div style={{ marginTop: 12 }}>
            <div className="muted">{user.email}</div>
            <div className="muted">Born {user.dateOfBirth}</div>
            {user.aboutMe && <p>{user.aboutMe}</p>}
            <div className="row" style={{ gap: 24, marginTop: 8 }}>
              <span>
                <strong>{followers.length}</strong> followers
              </span>
              <span>
                <strong>{following.length}</strong> following
              </span>
            </div>
          </div>
        ) : (
          <p className="muted" style={{ marginTop: 12 }}>
            This profile is private. Follow this user to see their information,
            posts and connections.
          </p>
        )}
      </div>

      {profile.canViewFull && (
        <>
          <div className="card">
            <h3 style={{ marginTop: 0 }}>Followers</h3>
            <UserList users={followers} empty="No followers yet." />
            <h3>Following</h3>
            <UserList users={following} empty="Not following anyone yet." />
          </div>

          <h3>Posts</h3>
          {posts.length === 0 && (
            <div className="card muted">No posts to show.</div>
          )}
          {posts.map((p) => (
            <PostCard key={p.id} post={p} />
          ))}
        </>
      )}

      {confirm === "unfollow" && (
        <ConfirmDialog
          title="Unfollow"
          message={`Stop following ${fullName}?`}
          confirmLabel="Unfollow"
          onConfirm={doUnfollow}
          onCancel={() => setConfirm(null)}
        />
      )}
      {confirm === "privacy" && (
        <ConfirmDialog
          title="Change profile visibility"
          message={`Make your profile ${user.isPublic ? "private" : "public"}?`}
          confirmLabel="Confirm"
          onConfirm={togglePrivacy}
          onCancel={() => setConfirm(null)}
        />
      )}
    </div>
  );
}

function UserList({ users, empty }: { users: User[]; empty: string }) {
  if (users.length === 0) return <div className="muted">{empty}</div>;
  return (
    <div className="col" style={{ gap: 6 }}>
      {users.map((u) => (
        <Link
          key={u.id}
          href={`/profile/${u.id}`}
          className="row"
          style={{ color: "var(--text)" }}
        >
          <Avatar name={`${u.firstName} ${u.lastName}`} avatar={u.avatar} size={32} />
          <span>
            {u.firstName} {u.lastName}
          </span>
        </Link>
      ))}
    </div>
  );
}
