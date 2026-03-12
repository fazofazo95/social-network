import Link from "next/link";

const SideBar = () => {
    return (       <aside className="bg-[#1a1a2e] border border-purple-500/30 rounded-xl p-4 shadow-custom">
        <nav>
          <ul className="flex flex-col gap-5">
            <li className="flex items-center gap-4 text-blue-400 hover:bg-purple-900/20 rounded-lg px-2 py-1.5 transition">
              <span className="text-lg">🏠</span>
              <Link href="/">Home</Link>
            </li>
            <li className="flex items-center gap-4 text-purple-300 hover:bg-purple-900/20 rounded-lg px-2 py-1.5 transition">
              <span className="text-lg">👤</span>
              <Link href="/profile">Profile</Link>
            </li>
            <li className="flex items-center gap-4 text-purple-400 hover:bg-purple-900/20 rounded-lg px-2 py-1.5 transition">
              <span className="text-lg">👥</span>
              <Link href="/groups">Groups</Link>
            </li>
            <li className="flex items-center gap-4 custom-pink-text hover:bg-purple-900/20 rounded-lg px-2 py-1.5 transition">
              <span className="text-lg">💬</span>
              <Link href="/messages">Messages</Link>
            </li>
            <li className="flex items-center gap-4 text-pink-400 hover:bg-purple-900/20 rounded-lg px-2 py-1.5 transition">
              <span className="text-lg">⚙️</span>
              <Link href="/settings">Settings</Link>
            </li>
          </ul>
        </nav>
      </aside> );
}
 
export default SideBar;