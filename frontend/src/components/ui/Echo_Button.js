'use client';

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
      <span className="text-lg">💧</span>
      Echo
    </button>
  );
};
 
export default Echo_Button;
