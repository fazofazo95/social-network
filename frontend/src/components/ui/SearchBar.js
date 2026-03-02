import Image from "next/image";
import Link from "next/link";
import { parseProfileImage } from "src/lib/utils/profileImage";

const SearchBar = ({ placeholder = "Search...", icon = "/search_icon.svg", onChange, className = "", results = [], onResultClick, isLoading = false, value = "" }) => {
  const showResults = value.trim().length > 0 && results.length > 0;

  return (
    <div className={`relative ${className}`}>
      <Image
        src={icon}
        alt="Search Icon"
        width={20}
        height={20}
        className="absolute left-4 top-1.75 w-2.5 pointer-events-none"
      />
      <input
        type="text"
        placeholder={placeholder}
        onChange={onChange}
        value={value}
        className="rounded-md bg-gray-200 ml-2 pl-6 py-1 text-black outline-none focus:ring-2 focus:ring-purple-400"
      />
      
      {showResults && (
        <div className="absolute top-full left-0 mt-1 w-56 bg-gray-200 rounded-md border border-gray-300 shadow-lg z-50 max-h-64 overflow-y-auto">
          {isLoading ? (
            <div className="px-3 py-2 text-gray-500 text-sm">Loading...</div>
          ) : (
            results.slice(0, 6).map((user) => (
              <Link
                key={user.id}
                href={`/profile/${user.id}`}
                onClick={onResultClick}
                className="flex items-center gap-3 px-3 py-2 hover:bg-gray-100 border-b border-gray-200 last:border-b-0 cursor-pointer"
              >
                <img
                  src={parseProfileImage(user.profile_picture)}
                  alt={user.display_name || `${user.first_name} ${user.last_name}`}
                  className="w-8 h-8 rounded-full object-cover shrink-0"
                />
                <div className="text-sm font-medium text-black">
                  {user.display_name || `${user.first_name || ""} ${user.last_name || ""}`.trim() || "User"}
                </div>
              </Link>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default SearchBar;
