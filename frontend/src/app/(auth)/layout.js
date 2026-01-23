import { EB_Garamond } from "next/font/google";
import "../globals.css";



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
        {children}
      </body>
    </html>
  );
}
