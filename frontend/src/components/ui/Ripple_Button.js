import { useEffect, useState } from "react";


const Ripple_Button = () => {
    const [status, setStatus] = useState("not_rippled");
    const [currentFrameIndex, setCurrentFrameIndex] = useState(0);
    
    const imageFrames = [
        "/ripples/Frame_1.svg",
        "/ripples/Frame_2.svg",
        "/ripples/Frame_3.svg",
        "/ripples/Frame_4.svg",
        "/ripples/Frame_5.svg",
        "/ripples/Frame_6.svg",
        "/ripples/Frame_7.svg",
        "/ripples/Frame_8.svg",
        // "/ripples/Frame_9.svg",
        // "/ripples/Frame_10.svg",
        // "/ripples/Frame_11.svg",
        // "/ripples/Frame_12.svg",
        // "/ripples/Frame_13.svg",
        // "/ripples/Frame_14.svg",
        // "/ripples/Frame_15.svg",
        // "/ripples/Frame_16.svg",
        "/ripples/Frame_15.svg"
    ];

    useEffect(() => {
        if (status === "rippled") {
            let frameIndex = 0;
            const intervalId = setInterval(() => {
                frameIndex++;
                if (frameIndex >= imageFrames.length) {
                    clearInterval(intervalId);
                    setCurrentFrameIndex(frameIndex - 1);
                } else {
                    setCurrentFrameIndex(frameIndex);
                }
                //play with the speed here 
            }, 50);
     
            return () => clearInterval(intervalId);
        } else {
            setCurrentFrameIndex(0);
        }
    }, [status]);
    const handleRipple = () => {
          if (status === "rippled") {
            setStatus("not_rippled");
          } else {
            setStatus("rippled");

          }
    };
    
    return ( 
         <button id="ripple_btn" className="flex cursor-pointer gap-1" onClick={handleRipple}>
            <img src={imageFrames[currentFrameIndex]} alt="Ripple Icon" width={20} height={20} />
                Ripple</button> 
                );
}
 
export default Ripple_Button;