import Image from "next/image";
import Follow_Bottom from "../Follow_Button";

const SuggestedFriends = () => {
  return (
    <section className="border rounded-xl w-64 border-purple-500 px-1  shadow-custom">
      <h1 className="custom-pink-text pt-2 pb-1">Suggested Friends</h1>
      <ul className="px-3 py-2 flex flex-col gap-3">
        <li className="flex gap-1">
          <Image
            src="/profil2_icon.svg"
            alt="Profile Icon"
            width={20}
            height={20}
          />
          <span>Friend 1</span>
          <Follow_Bottom />
        </li>

        <li className="flex  gap-1">
          <Image
            src="/profil2_icon.svg"
            alt="Profile Icon"
            width={20}
            height={20}
          />
          <span>Friend 2</span>
          <Follow_Bottom />
        </li>

        <li className="flex  gap-1">
          <Image
            src="/profil2_icon.svg"
            alt="Profile Icon"
            width={20}
            height={20}
          />
          <span>Friend 3</span>
          <Follow_Bottom />
        </li>
        <li className="flex gap-1">
          <Image
            src="/profil2_icon.svg"
            alt="Profile Icon"
            width={20}
            height={20}
          />
          <span>Friend 4</span>
          <Follow_Bottom />
        </li>
      </ul>
    </section>
  );
};

export default SuggestedFriends;
