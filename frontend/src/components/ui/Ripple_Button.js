import { useEffect, useState } from "react";
import Image from "next/image";


const Ripple_Button = () => {
    const [status, setStatus] = useState("not_rippled");
    
    

    // useEffect(() => {
    //     if (status === "rippled") {
    //         let frameIndex = 0;
    //         const intervalId = setInterval(() => {
    //             frameIndex++;
    //             if (frameIndex >= imageFrames.length) {
    //                 clearInterval(intervalId);
    //                 setCurrentFrameIndex(frameIndex - 1);
    //             } else {
    //                 setCurrentFrameIndex(frameIndex);
    //             }
    //         }, 50);
     
    //         return () => clearInterval(intervalId);
    //     } else {
    //         setCurrentFrameIndex(0);
    //     }
    // }, [status]);
    const handleRipple = () => {
          if (status === "rippled") {
            setStatus("not_rippled");
          } else {
            setStatus("rippled");

          }
    };
    
    return ( 
         <button id="ripple_btn" className="flex justify-center cursor-pointer gap-1" onClick={handleRipple}>
           <Image
               src={status === "not_rippled" ? "ripples/Unripple_icon.svg" : "ripples/Ripple_icon.svg"}
               alt="Ripple Icon"
               width={20}
               height={20}
           />
                Ripple</button> 
                );
}
 
export default Ripple_Button;