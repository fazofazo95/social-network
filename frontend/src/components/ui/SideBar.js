import Image from "next/image";
import Link from "next/link";

const SideBar = () => {
    return (       <aside>
        <nav>
          <ul className="flex flex-col gap-5">
            <li className=" flex items-center gap-4 text-blue-600">
              <Image src="/home_icon.svg" alt="Home Icon" width={20} height={20} />
              <Link href="/">Home</Link>
            </li>
            <li className=" flex items-center gap-4 text-purple-900">
              <Image src="/profil2_icon.svg" alt="Profile Icon" width={20} height={20} />
              <Link href="/profile">Profile</Link>
            </li>
            <li className=" flex items-center gap-4 text-purple-600">
              <Image src="/groups_icon.svg" alt="Groups Icon" width={20} height={20} />
              <Link href="/groups">Groups</Link>
            </li>
            <li className=" flex items-center gap-4 custom-pink-text">
              <Image src="/messages_icon.svg" alt="Messages Icon" width={20} height={20} />
              <Link href="/messages">Messages</Link>
            </li>
            <li className=" flex items-center gap-4 text-pink-500">
              <Image src="/settings_icon.svg" alt="Settings Icon" width={20} height={20} />
              <Link href="/settings">Settings</Link>
            </li>
          </ul>
        </nav>
      </aside> );
}
 
export default SideBar;