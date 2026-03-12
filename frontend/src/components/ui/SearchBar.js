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
        width={14}
        height={14}
        className="absolute left-4 top-1/2 -translate-y-1/2 w-3.5 pointer-events-none"
      />
      <input
        type="text"
        placeholder={placeholder}
        onChange={onChange}
        value={value}
        className="w-44 rounded-md bg-[#0d0d1a] border border-purple-500/30 ml-2 pl-7 pr-2 py-0.5 text-sm text-purple-100 placeholder-purple-400/50 outline-none focus:border-purple-500/50 focus:w-56 transition-all"
      />
      
      {value.trim().length > 0 && (
        <div className="absolute top-full left-0 mt-1 w-64 bg-[#1a1a2e] rounded-md border border-purple-500/30 shadow-custom z-50 max-h-72 overflow-y-auto">
          {isLoading ? (
            <div className="px-3 py-2 text-purple-400 text-sm">Loading...</div>
          ) : !hasResults ? (
            <div className="px-3 py-2 text-purple-400 text-sm">No results found</div>
          ) : (
            <>
              {users.length > 0 && (
                <>
                  <div className="px-3 py-1 text-xs font-semibold text-purple-400 uppercase tracking-wide bg-purple-900/20 border-b border-purple-500/20">Users</div>
                  {users.slice(0, 4).map((user) => (
                    <Link
                      key={`user-${user.id}`}
                      href={`/profile/${user.id}`}
                      onClick={onResultClick}
                      className="flex items-center gap-3 px-3 py-2 hover:bg-purple-900/20 border-b border-purple-500/10 last:border-b-0 cursor-pointer transition"
                    >
                      <img
                        src={parseProfileImage(user.profile_picture)}
                        alt={`${user.first_name} ${user.last_name}`}
                        className="w-8 h-8 rounded-full object-cover shrink-0"
                      />
                      <div>
                        <div className="text-sm font-medium text-purple-100">
                          {`${user.first_name || ""} ${user.last_name || ""}`.trim() || "User"}
                        </div>
                        {user.username && (
                          <div className="text-xs text-purple-400">@{user.username}</div>
                        )}
                      </div>
                    </Link>
                  ))}
                </>
              )}
              {groups.length > 0 && (
                <>
                  <div className="px-3 py-1 text-xs font-semibold text-purple-400 uppercase tracking-wide bg-purple-900/20 border-b border-purple-500/20">Groups</div>
                  {groups.slice(0, 4).map((group) => (
                    <Link
                      key={`group-${group.id}`}
                      href={`/groups/${group.id}`}
                      onClick={onResultClick}
                      className="flex items-center gap-3 px-3 py-2 hover:bg-purple-900/20 border-b border-purple-500/10 last:border-b-0 cursor-pointer transition"
                    >
                      <img
                        src={parseProfileImage(group.group_picture)}
                        alt={group.name}
                        className="w-8 h-8 rounded-full object-cover shrink-0"
                      />
                      <div>
                        <div className="text-sm font-medium text-purple-100">{group.name}</div>
                        <div className="text-xs text-purple-400">{group.group_members} members</div>
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
