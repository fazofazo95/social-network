import Image from "next/image";
import IconButton from "./IconButton";
import SearchBar from "./SearchBar";

const NavBar = () => {
  return (
    <nav className="border w-full h-8 bg-white relative flex flex-row justify-between">
      <div className="flex items-center relative">
        <Image src="/logo_icon.svg" alt="Logo" width={25} height={25} />
        <div className="relative">
          <h1 className="text-purple-500 text-2xl font-semibold relative">
            Pulse
          </h1>
          <h1 className="absolute top-0 left-0 text-2xl text-purple-500 neon-glow">
            Pulse
          </h1>
        </div>
        
        <SearchBar placeholder="Search..." />
      </div>

      <div className="flex items-center gap-3 pr-4">
        <IconButton 
          icon="/notif-icon.svg" 
          alt="Notification Icon" 
          iconSize={17}
        />
        
        <IconButton 
          icon="/profil_icon.svg" 
          alt="Profile Icon" 
          iconSize={20}
        />
      </div>
    </nav>
  );
};

export default NavBar;
