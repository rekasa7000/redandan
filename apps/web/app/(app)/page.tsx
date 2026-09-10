"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { TaskCard } from "@/components/tasks/task-card";
import { ApiError, listContexts, listTasks, updateTask, type Context, type Task } from "@/lib/api";
import { clearToken } from "@/lib/session";
import { addDaysToDateString, todayDateString } from "@/lib/utils";

export default function DashboardPage() {
  const router = useRouter();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [contexts, setContexts] = useState<Context[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [taskList, contextList] = await Promise.all([listTasks(), listContexts()]);
      setTasks(taskList);
      setContexts(contextList);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.replace("/login");
        return;
      }
      setError(err instanceof Error ? err.message : "Could not load dashboard");
    } finally {
      setLoading(false);
    }
  }, [router]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- fetch-on-mount, no RSC/SWR in this app (see docs/architecture.md)
    load();
  }, [load]);

  async function toggleDone(task: Task) {
    if (!task.id) return;
    const nextStatus = task.status === "done" ? "todo" : "done";
    setTasks((prev) => prev.map((t) => (t.id === task.id ? { ...t, status: nextStatus } : t)));
    try {
      await updateTask(task.id, { status: nextStatus });
    } catch {
      setTasks((prev) => prev.map((t) => (t.id === task.id ? task : t)));
    }
  }

  // Calendar-date string comparisons — avoids timezone drift from Date arithmetic.
  const today = todayDateString();
  const in7Days = addDaysToDateString(today, 7);
  const dateOf = (t: Task) => t.due_date?.slice(0, 10);

  const active = tasks.filter((t) => t.status !== "done" && t.status !== "archived");
  const overdue = active.filter((t) => {
    const d = dateOf(t);
    return d && d < today;
  });
  const dueToday = active.filter((t) => dateOf(t) === today);
  const upcoming = active.filter((t) => {
    const d = dateOf(t);
    return d && d > today && d < in7Days;
  });

  const contextsById = new Map(contexts.map((c) => [c.id, c]));

  function renderSection(title: string, items: Task[]) {
    if (items.length === 0) return null;
    return (
      <section className="space-y-2">
        <h2 className="text-sm font-medium text-muted-foreground">{title}</h2>
        <div className="flex flex-col gap-2">
          {items.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              context={contextsById.get(task.context_id ?? "")}
              onToggleDone={toggleDone}
            />
          ))}
        </div>
      </section>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <div className="flex items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
        <Button asChild size="sm">
          <Link href="/tasks/new">
            <Plus /> New task
          </Link>
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : active.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Nothing on your plate.{" "}
          <Link href="/tasks/new" className="underline underline-offset-4">
            Add a task
          </Link>{" "}
          to get started.
        </p>
      ) : (
        <>
          {renderSection("Overdue", overdue)}
          {renderSection("Due today", dueToday)}
          {renderSection("Upcoming", upcoming)}
          {overdue.length === 0 && dueToday.length === 0 && upcoming.length === 0 && (
            <p className="text-sm text-muted-foreground">
              No deadlines in the next 7 days.{" "}
              <Link href="/tasks" className="underline underline-offset-4">
                View all tasks
              </Link>
              .
            </p>
          )}
        </>
      )}
    </div>
  );
}
