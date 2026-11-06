"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { api, wsUrl } from "./api";
import { useAuth } from "./auth";
import type { Message, Notification } from "./types";

type MessageListener = (msg: Message) => void;

interface SocketState {
  connected: boolean;
  onlineUsers: Set<string>;
  isOnline: (id: string) => boolean;
  notifications: Notification[];
  unreadNotifications: number;
  unreadMessages: number;
  clearMessageBadge: () => void;
  // setActiveConversation tells the provider which conversation the user is
  // currently looking at ("user:<id>" | "group:<id>" | null); messages arriving
  // for that conversation are visible, so they must not count as unread.
  setActiveConversation: (key: string | null) => void;
  refreshNotifications: () => Promise<void>;
  markAllNotificationsRead: () => Promise<void>;
  sendPrivate: (to: string, content: string) => void;
  sendGroup: (groupId: string, content: string) => void;
  onMessage: (cb: MessageListener) => () => void;
}

const SocketContext = createContext<SocketState | undefined>(undefined);

export function SocketProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const [connected, setConnected] = useState(false);
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set());
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadMessages, setUnreadMessages] = useState(0);

  const wsRef = useRef<WebSocket | null>(null);
  const listenersRef = useRef<Set<MessageListener>>(new Set());
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const activeConvRef = useRef<string | null>(null);

  const setActiveConversation = useCallback((key: string | null) => {
    activeConvRef.current = key;
  }, []);

  const refreshNotifications = useCallback(async () => {
    try {
      const notes = await api.get<Notification[]>("/api/notifications");
      setNotifications(notes || []);
    } catch {
      /* ignore */
    }
  }, []);

  const markAllNotificationsRead = useCallback(async () => {
    await api.post("/api/notifications/read-all");
    setNotifications((prev) => prev.map((n) => ({ ...n, isRead: true })));
  }, []);

  const onMessage = useCallback((cb: MessageListener) => {
    listenersRef.current.add(cb);
    return () => {
      listenersRef.current.delete(cb);
    };
  }, []);

  const send = useCallback((payload: unknown) => {
    const ws = wsRef.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(payload));
    }
  }, []);

  const sendPrivate = useCallback(
    (to: string, content: string) =>
      send({ type: "private_message", to, content }),
    [send]
  );
  const sendGroup = useCallback(
    (groupId: string, content: string) =>
      send({ type: "group_message", groupId, content }),
    [send]
  );

  const clearMessageBadge = useCallback(() => setUnreadMessages(0), []);

  // Establish (and auto-reconnect) the websocket while logged in.
  useEffect(() => {
    if (!user) {
      wsRef.current?.close();
      wsRef.current = null;
      setConnected(false);
      return;
    }

    refreshNotifications();
    let closed = false;

    const connect = () => {
      const ws = new WebSocket(wsUrl());
      wsRef.current = ws;

      ws.onopen = () => setConnected(true);
      ws.onclose = () => {
        setConnected(false);
        if (!closed) {
          reconnectRef.current = setTimeout(connect, 2000);
        }
      };
      ws.onmessage = (event) => {
        let data: any;
        try {
          data = JSON.parse(event.data);
        } catch {
          return;
        }
        switch (data.type) {
          case "online_users":
            setOnlineUsers(new Set<string>(data.userIds || []));
            break;
          case "presence":
            setOnlineUsers((prev) => {
              const next = new Set(prev);
              if (data.online) next.add(data.userId);
              else next.delete(data.userId);
              return next;
            });
            break;
          case "message": {
            const msg: Message = data.message;
            listenersRef.current.forEach((cb) => cb(msg));
            // The Chat badge counts only *private* messages (the subject asks
            // for an indicator of new private messages; group chats live on
            // their group pages). Skip own messages and the conversation the
            // user is currently looking at.
            if (
              !msg.groupId &&
              msg.senderId !== user.id &&
              activeConvRef.current !== `user:${msg.senderId}`
            ) {
              setUnreadMessages((n) => n + 1);
            }
            break;
          }
          case "notification":
            setNotifications((prev) => [data.notification, ...prev]);
            break;
        }
      };
    };

    connect();
    return () => {
      closed = true;
      if (reconnectRef.current) clearTimeout(reconnectRef.current);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [user, refreshNotifications]);

  const isOnline = useCallback(
    (id: string) => onlineUsers.has(id),
    [onlineUsers]
  );
  const unreadNotifications = notifications.filter((n) => !n.isRead).length;

  return (
    <SocketContext.Provider
      value={{
        connected,
        onlineUsers,
        isOnline,
        notifications,
        unreadNotifications,
        unreadMessages,
        clearMessageBadge,
        setActiveConversation,
        refreshNotifications,
        markAllNotificationsRead,
        sendPrivate,
        sendGroup,
        onMessage,
      }}
    >
      {children}
    </SocketContext.Provider>
  );
}

export function useSocket(): SocketState {
  const ctx = useContext(SocketContext);
  if (!ctx) throw new Error("useSocket must be used within SocketProvider");
  return ctx;
}
