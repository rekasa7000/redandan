"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { TaskCard } from "@/components/tasks/task-card";
import { TaskFilters } from "@/components/tasks/task-filters";
import { ApiError, listContexts, listTasks, updateTask, type Context, type Task } from "@/lib/api";
import { clearToken } from "@/lib/session";

export default function TasksPage() {
  return (
    <Suspense fallback={<p className="p-6 text-sm text-muted-foreground">Loading…</p>}>
      <TasksPageInner />
    </Suspense>
  );
}

function TasksPageInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const status = searchParams.get("status") ?? "";
  const contextId = searchParams.get("context_id") ?? "";

  const [tasks, setTasks] = useState<Task[]>([]);
  const [contexts, setContexts] = useState<Context[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const [taskList, contextList] = await Promise.all([
        listTasks({ status: status || undefined, context_id: contextId || undefined }),
        listContexts(),
      ]);
      setTasks(taskList);
      setContexts(contextList);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearToken();
        router.replace("/login");
        return;
      }
      setError(err instanceof Error ? err.message : "Could not load tasks");
    } finally {
      setLoading(false);
    }
  }, [status, contextId, router]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- fetch-on-mount, no RSC/SWR in this app (see docs/architecture.md)
    load();
  }, [load]);

  function updateFilters({ status, contextId }: { status: string; contextId: string }) {
    const params = new URLSearchParams();
    if (status) params.set("status", status);
    if (contextId) params.set("context_id", contextId);
    router.push(`/tasks${params.toString() ? `?${params}` : ""}`);
  }

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

  const contextsById = new Map(contexts.map((c) => [c.id, c]));

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <div className="flex items-center justify-between gap-4">
        <h1 className="text-2xl font-semibold tracking-tight">Tasks</h1>
        <Button asChild size="sm">
          <Link href="/tasks/new">
            <Plus /> New task
          </Link>
        </Button>
      </div>

      <TaskFilters
        contexts={contexts}
        status={status}
        contextId={contextId}
        onChange={updateFilters}
      />

      {error && <p className="text-sm text-destructive">{error}</p>}

      {loading ? (
        <p className="text-sm text-muted-foreground">Loading…</p>
      ) : tasks.length === 0 ? (
        <p className="text-sm text-muted-foreground">No tasks match these filters.</p>
      ) : (
        <div className="flex flex-col gap-2">
          {tasks.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              context={contextsById.get(task.context_id ?? "")}
              onToggleDone={toggleDone}
            />
          ))}
        </div>
      )}
    </div>
  );
}
