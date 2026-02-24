
'use client';
import Image from "next/image";

const Echo_Button = ({ targetId, onToggle }) => {
  const handleEchoBtn = () => {
    if (!targetId) {
      return;
    }

    const echoSection = document.getElementById(targetId);
    if (!echoSection) {
      return;
    }

    const willOpen = echoSection.classList.contains("hidden");

    if (echoSection.classList.contains("hidden")) {
      echoSection.classList.remove("hidden");
      echoSection.classList.add("flex");
    } else {
      echoSection.classList.add("hidden");
      echoSection.classList.remove("flex");
    }

    if (typeof onToggle === "function") {
      onToggle(willOpen);
    }
  };
  return (
    <button type="button" onClick={handleEchoBtn} className="flex cursor-pointer gap-1">
      <Image
        src="/echo_icon.svg"
        alt="Echo Icon"
        width={20}
        height={20}
      />
      Echo
    </button>
  );
};
 
export default Echo_Button;
