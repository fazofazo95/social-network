import Image from "next/image";

const IconButton = ({ icon, alt, iconSize = 20, onClick, className = "" }) => {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`w-6 h-6 flex items-center justify-center rounded-full bg-white hover:bg-(--color-customPurple) transition ${className}`}
    >
      <Image
        src={icon}
        alt={alt}
        width={iconSize}
        height={iconSize}
      />
    </button>
  );
};

export default IconButton;
