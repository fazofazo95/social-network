import Image from "next/image";

const Echo_Button = () => {
      const handleEchoBtn = () => {
    const echoSection = document.getElementById("echo-section");
    if (echoSection.classList.contains("hidden")) {
      echoSection.classList.remove("hidden");
      echoSection.classList.add("flex");
    } else {
      echoSection.classList.add("hidden");
      echoSection.classList.remove("flex");
    }
  };
    return ( 
                  <button onClick={handleEchoBtn} className="flex cursor-pointer gap-1">
                    <Image
                      src="/echo_icon.svg"
                      alt="Echo Icon"
                      width={20}
                      height={20}
                    />
                    Echo
                  </button>
     );
}
 
export default Echo_Button;
