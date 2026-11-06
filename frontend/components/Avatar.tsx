import { mediaUrl } from "@/lib/api";

interface Props {
  name: string;
  avatar?: string | null;
  size?: number;
}

// Avatar renders a user's image, or their initials when no avatar is set.
export default function Avatar({ name, avatar, size = 40 }: Props) {
  const url = mediaUrl(avatar);
  if (url) {
    // eslint-disable-next-line @next/next/no-img-element
    return (
      <img
        className="avatar"
        src={url}
        alt={name}
        width={size}
        height={size}
        style={{
          width: size,
          height: size,
          minWidth: size,
          minHeight: size,
          flex: "0 0 auto",
          objectFit: "cover",
        }}
      />
    );
  }
  const initials = name
    .split(" ")
    .map((p) => p[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
  return (
    <span
      className="avatar-fallback"
      style={{
        width: size,
        height: size,
        minWidth: size,
        minHeight: size,
        flex: "0 0 auto",
        fontSize: size * 0.4,
      }}
    >
      {initials}
    </span>
  );
}
