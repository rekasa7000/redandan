"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutDashboard, ListTodo, Settings } from "lucide-react";

import { cn } from "@/lib/utils";
import { listContexts, type Context } from "@/lib/api";

const navItems = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/tasks", label: "Tasks", icon: ListTodo },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar({ className }: { className?: string }) {
  const pathname = usePathname();
  const [contexts, setContexts] = useState<Context[]>([]);

  useEffect(() => {
    listContexts()
      .then(setContexts)
      .catch(() => setContexts([]));
  }, []);

  return (
    <aside
      className={cn(
        "w-60 shrink-0 flex-col gap-6 border-r border-border bg-card px-4 py-6",
        className,
      )}
    >
      <span className="px-2 text-lg font-semibold tracking-tight">Reliva</span>

      <nav className="flex flex-col gap-1">
        {navItems.map(({ href, label, icon: Icon }) => {
          const active = href === "/" ? pathname === "/" : pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground",
                active && "bg-muted text-foreground",
              )}
            >
              <Icon className="size-4" />
              {label}
            </Link>
          );
        })}
      </nav>

      {contexts.length > 0 && (
        <div className="flex flex-col gap-1">
          <span className="px-2 text-xs font-medium text-muted-foreground">Contexts</span>
          {contexts.map((context) => (
            <Link
              key={context.id}
              href={`/tasks?context_id=${context.id}`}
              className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <span
                className="size-2 shrink-0 rounded-full"
                style={{ backgroundColor: context.color }}
              />
              {context.name}
            </Link>
          ))}
        </div>
      )}
    </aside>
  );
}
