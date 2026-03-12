import Image from "next/image";
import Link from "next/link";
import { parseProfileImage } from "src/lib/utils/profileImage";

const SearchBar = ({ placeholder = "Search...", icon = "/search_icon.svg", onChange, className = "", users = [], groups = [], onResultClick, isLoading = false, value = "" }) => {
  const hasResults = value.trim().length > 0 && (users.length > 0 || groups.length > 0);

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
      
      {value.trim().length > 0 && (
        <div className="absolute top-full left-0 mt-1 w-64 bg-gray-200 rounded-md border border-gray-300 shadow-lg z-50 max-h-72 overflow-y-auto">
          {isLoading ? (
            <div className="px-3 py-2 text-gray-500 text-sm">Loading...</div>
          ) : !hasResults ? (
            <div className="px-3 py-2 text-gray-500 text-sm">No results found</div>
          ) : (
            <>
              {users.length > 0 && (
                <>
                  <div className="px-3 py-1 text-xs font-semibold text-gray-500 uppercase tracking-wide bg-gray-100">Users</div>
                  {users.slice(0, 4).map((user) => (
                    <Link
                      key={`user-${user.id}`}
                      href={`/profile/${user.id}`}
                      onClick={onResultClick}
                      className="flex items-center gap-3 px-3 py-2 hover:bg-gray-100 border-b border-gray-200 last:border-b-0 cursor-pointer"
                    >
                      <img
                        src={parseProfileImage(user.profile_picture)}
                        alt={`${user.first_name} ${user.last_name}`}
                        className="w-8 h-8 rounded-full object-cover shrink-0"
                      />
                      <div>
                        <div className="text-sm font-medium text-black">
                          {`${user.first_name || ""} ${user.last_name || ""}`.trim() || "User"}
                        </div>
                        {user.username && (
                          <div className="text-xs text-gray-500">@{user.username}</div>
                        )}
                      </div>
                    </Link>
                  ))}
                </>
              )}
              {groups.length > 0 && (
                <>
                  <div className="px-3 py-1 text-xs font-semibold text-gray-500 uppercase tracking-wide bg-gray-100">Groups</div>
                  {groups.slice(0, 4).map((group) => (
                    <Link
                      key={`group-${group.id}`}
                      href={`/groups/${group.id}`}
                      onClick={onResultClick}
                      className="flex items-center gap-3 px-3 py-2 hover:bg-gray-100 border-b border-gray-200 last:border-b-0 cursor-pointer"
                    >
                      <img
                        src={parseProfileImage(group.group_picture)}
                        alt={group.name}
                        className="w-8 h-8 rounded-full object-cover shrink-0"
                      />
                      <div>
                        <div className="text-sm font-medium text-black">{group.name}</div>
                        <div className="text-xs text-gray-500">{group.group_members} members</div>
                      </div>
                    </Link>
                  ))}
                </>
              )}
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default SearchBar;
