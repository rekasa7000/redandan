"use client";

import Link from "next/link";
import { Check } from "lucide-react";

import { cn, todayDateString } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { Context, Task } from "@/lib/api";

const priorityVariant: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  low: "outline",
  medium: "secondary",
  high: "default",
  urgent: "destructive",
};

const statusLabel: Record<string, string> = {
  todo: "To do",
  in_progress: "In progress",
  done: "Done",
  archived: "Archived",
};

function formatDueDate(dueDate: string) {
  // Compare calendar dates as strings — avoids timezone drift from Date arithmetic.
  const dateStr = dueDate.slice(0, 10);
  const isOverdue = dateStr < todayDateString();
  const label = new Date(`${dateStr}T00:00:00`).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
  });
  return { label, isOverdue };
}

export function TaskCard({
  task,
  context,
  onToggleDone,
}: {
  task: Task;
  context?: Context;
  onToggleDone?: (task: Task) => void;
}) {
  const done = task.status === "done";
  const due = task.due_date ? formatDueDate(task.due_date) : null;

  return (
    <div className="flex items-center gap-3 rounded-2xl bg-card px-4 py-3 shadow-sm ring-1 ring-foreground/5 dark:ring-foreground/10">
      <Button
        variant={done ? "default" : "outline"}
        size="icon-sm"
        className="shrink-0 rounded-full"
        aria-label={done ? "Mark as not done" : "Mark as done"}
        onClick={() => onToggleDone?.(task)}
      >
        {done && <Check className="size-3.5" />}
      </Button>

      <Link href={`/tasks/${task.id}`} className="min-w-0 flex-1">
        <p className={cn("truncate text-sm font-medium", done && "text-muted-foreground line-through")}>
          {task.title}
        </p>
        <div className="mt-1 flex flex-wrap items-center gap-1.5">
          {context && (
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <span
                className="size-1.5 shrink-0 rounded-full"
                style={{ backgroundColor: context.color }}
              />
              {context.name}
            </span>
          )}
          {task.priority && (
            <Badge variant={priorityVariant[task.priority] ?? "outline"}>{task.priority}</Badge>
          )}
          {task.status && task.status !== "todo" && (
            <Badge variant="outline">{statusLabel[task.status] ?? task.status}</Badge>
          )}
          {due && (
            <span className={cn("text-xs text-muted-foreground", due.isOverdue && !done && "text-destructive")}>
              {due.label}
            </span>
          )}
        </div>
      </Link>
    </div>
  );
}
