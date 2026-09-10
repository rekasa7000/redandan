// Middleware already redirects unauthenticated users to /login before this
// layout runs. Here we only need to provide the shell for authenticated pages.
import { Sidebar } from "@/components/shared/sidebar";
import { BottomNav } from "@/components/shared/bottom-nav";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-svh">
      <Sidebar className="hidden md:flex" />
      <main className="min-w-0 flex-1 pb-16 md:pb-0">{children}</main>
      <BottomNav className="md:hidden" />
    </div>
  );
}
