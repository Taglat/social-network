"use client";

const EMOJIS = [
  "😀", "😁", "😂", "🤣", "😊", "😍", "😘", "😎",
  "🤔", "😴", "😢", "😭", "😡", "👍", "👎", "👏",
  "🙏", "💪", "🔥", "🎉", "❤️", "💔", "⭐", "✅",
  "❌", "💯", "🥳", "😅", "🤗", "😇", "🤩", "😋",
];

// EmojiPicker shows a small grid of emojis; clicking one inserts it.
export default function EmojiPicker({
  onSelect,
}: {
  onSelect: (emoji: string) => void;
}) {
  return (
    <div className="emoji-pop">
      {EMOJIS.map((e) => (
        <button key={e} type="button" onClick={() => onSelect(e)}>
          {e}
        </button>
      ))}
    </div>
  );
}
