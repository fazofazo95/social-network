import Image from "next/image";

const IconButton = ({ icon, emoji, alt, iconSize = 20, onClick, className = "" }) => {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-6 h-6 flex items-center justify-center rounded-full bg-transparent hover:bg-purple-900/30 transition cursor-pointer ${className}`}
    >
      {emoji ? (
        <span className="text-sm">{emoji}</span>
      ) : (
        <Image
          src={icon}
          alt={alt}
          width={iconSize}
          height={iconSize}
        />
      )}
    </button>
  );
};

export default IconButton;
