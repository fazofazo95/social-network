import Image from "next/image";

const Logo = ({ title, subtitle, variant = "default" }) => {
  const glowVariants = {
    default: "neon-glow",
    blur: "blur-sm opacity-50"
  };

  return (
    <div className="flex flex-col items-center justify-center gap-0.5 mb-19">
      <div className="flex items-center relative">
        <Image src="/logo_icon.svg" alt="Logo" width={40} height={40} />
        <div className="relative">
          <h1 className="text-purple-500 text-4xl font-semibold relative z-10">
            Pulse
          </h1>
          <h1 className={`absolute top-0 left-0 text-4xl text-purple-400 ${glowVariants[variant]}`}>
            Pulse
          </h1>
        </div>
      </div>
      {title && (
        <p className="font-bold text-2xl text-center mb-2 pl-0.5">
          {title}
        </p>
      )}
      {subtitle && (
        <p className="text-gray-500 font-light text-md text-center">
          {subtitle}
        </p>
      )}
    </div>
  );
};

export default Logo;
