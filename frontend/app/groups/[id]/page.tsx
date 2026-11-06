"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useSocket } from "@/lib/socket";
import type { EventItem, GroupDetail, Message } from "@/lib/types";
import Avatar from "@/components/Avatar";
import CreatePost from "@/components/CreatePost";
import PostCard from "@/components/PostCard";
import ChatBox from "@/components/ChatBox";
import ConfirmDialog from "@/components/ConfirmDialog";

export default function GroupDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { sendGroup, setActiveConversation } = useSocket();
  const [data, setData] = useState<GroupDetail | null>(null);
  const [error, setError] = useState("");
  const [confirmLeave, setConfirmLeave] = useState(false);

  const load = useCallback(() => {
    api
      .get<GroupDetail>(`/api/groups/${id}`)
      .then(setData)
      .catch((e) => setError(e.message));
  }, [id]);

  useEffect(load, [load]);

  const loadHistory = useCallback(
    () => api.get<Message[]>(`/api/groups/${id}/messages`),
    [id]
  );
  const acceptMsg = useCallback((m: Message) => m.groupId === id, [id]);

  // While a member is on the group page its chat is visible, so messages of
  // this group must not count as unread.
  const memberOnPage = data?.group.membership === "member";
  useEffect(() => {
    if (!memberOnPage) return;
    setActiveConversation(`group:${id}`);
    return () => setActiveConversation(null);
  }, [memberOnPage, id, setActiveConversation]);

  if (error) return <div className="container card error">{error}</div>;
  if (!data) return <div className="container muted">Loading…</div>;

  const { group, members, posts, events, pendingRequests, nonMembers } = data;
  const isMember = group.membership === "member";
  const isCreator = group.creatorId === user?.id;

  // run wraps an action: it reports any error (e.g. stale state -> 404) instead
  // of letting the promise reject unhandled, then refreshes the group data.
  const run = async (fn: () => Promise<unknown>) => {
    try {
      await fn();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Action failed");
    } finally {
      load();
    }
  };

  const invite = (userId: string) =>
    run(() => api.post(`/api/groups/${id}/invite`, { userId }));
  const respondInvite = (action: "accept" | "decline") =>
    run(() => api.post(`/api/groups/${id}/invitation/respond`, { action }));
  const requestJoin = () => run(() => api.post(`/api/groups/${id}/request`));
  const respondRequest = (userId: string, action: "accept" | "decline") =>
    run(() => api.post(`/api/groups/${id}/requests/${userId}/respond`, { action }));
  const leave = () => {
    setConfirmLeave(false);
    run(() => api.post(`/api/groups/${id}/leave`));
  };

  return (
    <div className="container">
      <div className="card">
        <h2 style={{ margin: 0 }}>{group.title}</h2>
        <p>{group.description}</p>
        <div className="muted">{group.memberCount} members</div>

        {group.membership === "invited" && (
          <div className="row" style={{ marginTop: 10 }}>
            <span className="tag">You're invited</span>
            <button className="small" onClick={() => respondInvite("accept")}>
              Accept
            </button>
            <button className="small secondary" onClick={() => respondInvite("decline")}>
              Decline
            </button>
          </div>
        )}
        {group.membership === "" && (
          <button onClick={requestJoin} style={{ marginTop: 10 }}>
            Request to join
          </button>
        )}
        {group.membership === "requested" && (
          <span className="tag">Join request pending</span>
        )}
        {isMember && !isCreator && (
          <button
            className="secondary small"
            style={{ marginTop: 10 }}
            onClick={() => setConfirmLeave(true)}
          >
            Leave group
          </button>
        )}
      </div>

      {confirmLeave && (
        <ConfirmDialog
          title="Leave group"
          message={`Leave "${group.title}"? You will lose access to its posts, events and chat.`}
          confirmLabel="Leave"
          onConfirm={leave}
          onCancel={() => setConfirmLeave(false)}
        />
      )}

      {/* Members */}
      <div className="card">
        <h3 style={{ marginTop: 0 }}>Members</h3>
        <div className="col" style={{ gap: 6 }}>
          {members.map((m) => (
            <div key={m.id} className="row">
              <Avatar name={`${m.firstName} ${m.lastName}`} avatar={m.avatar} size={30} />
              <span>
                {m.firstName} {m.lastName}
                {m.id === group.creatorId && <span className="tag" style={{ marginLeft: 6 }}>creator</span>}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Creator: pending join requests */}
      {isCreator && pendingRequests.length > 0 && (
        <div className="card">
          <h3 style={{ marginTop: 0 }}>Join requests</h3>
          {pendingRequests.map((u) => (
            <div key={u.id} className="spread" style={{ marginBottom: 8 }}>
              <span>
                {u.firstName} {u.lastName}
              </span>
              <div className="row">
                <button className="small" onClick={() => respondRequest(u.id, "accept")}>
                  Accept
                </button>
                <button className="small secondary" onClick={() => respondRequest(u.id, "decline")}>
                  Decline
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {isMember && (
        <>
          {/* Invite members */}
          <div className="card">
            <h3 style={{ marginTop: 0 }}>Invite a user</h3>
            {nonMembers.length === 0 ? (
              <div className="muted">Everyone is already in the group or invited.</div>
            ) : (
              <div className="col" style={{ gap: 6 }}>
                {nonMembers.map((u) => (
                  <div key={u.id} className="spread">
                    <span>
                      {u.firstName} {u.lastName}
                    </span>
                    <button className="small" onClick={() => invite(u.id)}>
                      Invite
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Events */}
          <EventsSection groupId={id} events={events} onChange={load} />

          {/* Group chat */}
          <div className="card" style={{ padding: 0, height: 420 }}>
            <ChatBox
              title={`${group.title} chat`}
              loadHistory={loadHistory}
              send={(content) => sendGroup(id, content)}
              accept={acceptMsg}
            />
          </div>

          {/* Group posts */}
          <h3>Group posts</h3>
          <CreatePost groupId={id} onCreated={load} />
          {posts.length === 0 && <div className="card muted">No posts yet.</div>}
          {posts.map((p) => (
            <PostCard key={p.id} post={p} />
          ))}
        </>
      )}
    </div>
  );
}

function EventsSection({
  groupId,
  events,
  onChange,
}: {
  groupId: string;
  events: EventItem[];
  onChange: () => void;
}) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventTime, setEventTime] = useState("");
  const [open, setOpen] = useState(false);
  const [creating, setCreating] = useState(false);

  const create = async (e: React.FormEvent) => {
    e.preventDefault();
    // Guard against double-submit: a second click while the request is in
    // flight would create the event (and its notifications) twice.
    if (creating) return;
    if (!title.trim() || !description.trim() || !eventTime) return;
    setCreating(true);
    try {
      await api.post(`/api/groups/${groupId}/events`, {
        title,
        description,
        eventTime,
      });
      setTitle("");
      setDescription("");
      setEventTime("");
      setOpen(false);
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to create event");
    } finally {
      setCreating(false);
      onChange();
    }
  };

  const respond = async (eventId: string, response: "going" | "not_going") => {
    try {
      await api.post(`/api/events/${eventId}/respond`, { response });
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed to respond");
    } finally {
      onChange();
    }
  };

  return (
    <div className="card">
      <div className="spread">
        <h3 style={{ margin: 0 }}>Events</h3>
        <button className="small secondary" onClick={() => setOpen((o) => !o)}>
          {open ? "Cancel" : "Create event"}
        </button>
      </div>

      {open && (
        <form onSubmit={create} className="col" style={{ marginTop: 10 }}>
          <input placeholder="Title" value={title} onChange={(e) => setTitle(e.target.value)} />
          <textarea
            placeholder="Description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
          <input
            type="datetime-local"
            value={eventTime}
            onChange={(e) => setEventTime(e.target.value)}
          />
          <button type="submit" disabled={creating} style={{ alignSelf: "flex-start" }}>
            {creating ? "Creating…" : "Create"}
          </button>
        </form>
      )}

      <div className="col" style={{ marginTop: 10 }}>
        {events.length === 0 && <div className="muted">No events yet.</div>}
        {events.map((ev) => (
          <div key={ev.id} className="card" style={{ marginBottom: 0 }}>
            <div className="spread">
              <strong>{ev.title}</strong>
              <span className="muted">{new Date(ev.eventTime).toLocaleString()}</span>
            </div>
            <p style={{ margin: "6px 0" }}>{ev.description}</p>
            <div className="row">
              <button
                className={"small" + (ev.myResponse === "going" ? "" : " secondary")}
                onClick={() => respond(ev.id, "going")}
              >
                Going ({ev.goingCount})
              </button>
              <button
                className={"small" + (ev.myResponse === "not_going" ? "" : " secondary")}
                onClick={() => respond(ev.id, "not_going")}
              >
                Not going ({ev.notGoingCount})
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
