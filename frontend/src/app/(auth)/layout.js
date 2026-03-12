import AuthGate from "../../components/auth/AuthGate";
import { ToastProvider } from "../../components/ui/Toast";

export default function AuthLayout({ children }) {
  return (
    <ToastProvider>
      <AuthGate requireAuth={false}>{children}</AuthGate>
    </ToastProvider>
  );
}
