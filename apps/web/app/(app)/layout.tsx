// Middleware already redirects unauthenticated users to /login before this
// layout runs. Here we only need to provide the shell for authenticated pages.
export default function AppLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
