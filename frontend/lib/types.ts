// TypeScript shapes mirroring the JSON returned by the Go backend.

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  dateOfBirth: string;
  avatar: string | null;
  nickname: string | null;
  aboutMe: string | null;
  isPublic: boolean;
  createdAt: string;
  online?: boolean;
}

export interface Post {
  id: string;
  userId: string;
  groupId: string | null;
  content: string;
  image: string | null;
  privacy: string;
  createdAt: string;
  authorName: string;
  authorAvatar: string | null;
  commentCount: number;
}

export interface Comment {
  id: string;
  postId: string;
  userId: string;
  content: string;
  image: string | null;
  createdAt: string;
  authorName: string;
  authorAvatar: string | null;
}

export interface Group {
  id: string;
  creatorId: string;
  title: string;
  description: string;
  createdAt: string;
  memberCount: number;
  membership: string; // "", "member", "invited", "requested"
}

export interface GroupDetail {
  group: Group;
  members: User[];
  posts: Post[];
  events: EventItem[];
  pendingRequests: User[];
  nonMembers: User[];
}

export interface EventItem {
  id: string;
  groupId: string;
  creatorId: string;
  title: string;
  description: string;
  eventTime: string;
  createdAt: string;
  goingCount: number;
  notGoingCount: number;
  myResponse: string | null;
}

export interface Message {
  id: string;
  senderId: string;
  receiverId: string | null;
  groupId: string | null;
  content: string;
  createdAt: string;
  senderName: string;
}

export interface Notification {
  id: string;
  userId: string;
  type: string;
  actorId: string | null;
  entityId: string | null;
  message: string;
  isRead: boolean;
  createdAt: string;
  // true when the accept/decline actions for this notification would still
  // succeed (pending follow request / group invite / group join request).
  actionable: boolean;
}

export interface Profile {
  user: User;
  posts: Post[];
  followers: User[];
  following: User[];
  isSelf: boolean;
  isFollowing: boolean;
  requestSent: boolean;
  canViewFull: boolean;
  online: boolean;
}
