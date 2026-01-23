import { EB_Garamond } from "next/font/google";
import "../globals.css";
import NavBar from "../../components/ui/NavBar";
import SideBar from "../../components/ui/SideBar";
import SuggestedFriends from "../../components/ui/Suggested_Friends";

const ebGaramond = EB_Garamond({
  subsets: ["latin"],
  weight: ["400"],
  display: "swap",
});


export const metadata = {
  title: "Pulse",
  description: "A simple social networking app built with Next.js",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body className={`${ebGaramond.className} bg-(--color-customPurple) font-medium text-white`}>
        <div className="sticky top-0 z-50">
          <NavBar />
        </div>
        <div className="grid grid-cols-[300px_1fr_300px] gap-10">
          <div className="flex flex-col gap-5 mt-10 ml-5 sticky top-10 self-start">
            <SideBar />
            <SuggestedFriends />
          </div>
          
          <div className="flex justify-center mt-5">
            {children}
          </div>
          
          
        </div>
      </body>
    </html>
  );
}
