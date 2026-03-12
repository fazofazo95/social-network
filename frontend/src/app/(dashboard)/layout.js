import NavBar from "../../components/ui/NavBar";
import SideBar from "../../components/ui/SideBar";
import SuggestedFriends from "../../components/ui/Suggested_Friends";
import AuthGate from "../../components/auth/AuthGate";
import { ToastProvider } from "../../components/ui/Toast";

export default function DashboardLayout({ children }) {
  return (
    <ToastProvider>
      <AuthGate requireAuth>
          <div className="sticky top-0 z-50">
            <NavBar />
          </div>
          <div className="grid grid-cols-[300px_1fr_300px] gap-10">
            <div className="flex flex-col gap-5 mt-10 ml-5 sticky top-10 self-start">
              <SideBar />
              <SuggestedFriends />
            </div>

            <div className="flex justify-center mt-10">
              {children}
            </div>
          </div>
      </AuthGate>
    </ToastProvider>
  );
}
