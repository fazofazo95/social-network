import Image from "next/image";
import { parseProfileImage } from "src/lib/utils/profileImage";

function getInitials(name, type) {
  if (!name || typeof name !== "string") return type === "group" ? "G" : "?";
  const parts = name.trim().split(/\s+/);
  if (type === "group") return parts[0]?.[0]?.toUpperCase() || "G";
  if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
  return parts[0]?.[0]?.toUpperCase() || "?";
}

const Avatar = ({ src, name, size = 40, className = "", type = "user" }) => {
  const imageUrl = parseProfileImage(src);

  if (imageUrl) {
    return (
      <Image
        src={imageUrl}
        alt={name || "Avatar"}
        width={size}
        height={size}
        className={`rounded-full ${className}`}
      />
    );
  }

  const initials = getInitials(name, type);
  const fontSize = size <= 20 ? "text-[10px]" : size <= 28 ? "text-xs" : size <= 40 ? "text-sm" : "text-base";

  return (
    <div
      className={`rounded-full bg-purple-600 flex items-center justify-center text-white font-bold shadow-[0_0_10px_rgba(168,85,247,0.3)] shrink-0 ${fontSize} ${className}`}
      style={{ width: size, height: size, minWidth: size }}
    >
      {initials}
    </div>
  );
};

export default Avatar;
