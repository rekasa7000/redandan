"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { TaskForm } from "@/components/tasks/task-form";
import {
  ApiError,
  deleteTask,
  getTask,
  listContexts,
  updateTask,
  type Context,
  type Task,
  type TaskPayload,
} from "@/lib/api";
import { clearToken } from "@/lib/session";

export default function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();

  const [task, setTask] = useState<Task | null>(null);
  const [contexts, setContexts] = useState<Context[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    Promise.all([getTask(id), listContexts()])
      .then(([task, contexts]) => {
        setTask(task);
        setContexts(contexts);
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 401) {
          clearToken();
          router.replace("/login");
          return;
        }
        setError(err instanceof Error ? err.message : "Could not load task");
      })
      .finally(() => setLoading(false));
  }, [id, router]);

  async function handleSubmit(values: TaskPayload) {
    const updated = await updateTask(id, values);
    setTask(updated);
  }

  async function handleDelete() {
    if (!confirm("Delete this task? This cannot be undone.")) return;
    setDeleting(true);
    try {
      await deleteTask(id);
      router.push("/tasks");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not delete task");
      setDeleting(false);
    }
  }

  if (loading) {
    return <p className="p-6 text-sm text-muted-foreground">Loading…</p>;
  }

  if (error && !task) {
    return <p className="p-6 text-sm text-destructive">{error}</p>;
  }

  if (!task) return null;

  return (
    <div className="mx-auto max-w-lg space-y-4 p-6">
      <Card>
        <CardHeader>
          <CardTitle>Edit task</CardTitle>
        </CardHeader>
        <CardContent>
          <TaskForm
            contexts={contexts}
            task={task}
            onSubmit={handleSubmit}
            submitLabel="Save changes"
          />
        </CardContent>
      </Card>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Button
        variant="destructive"
        className="w-full"
        onClick={handleDelete}
        disabled={deleting}
      >
        {deleting ? "Deleting…" : "Delete task"}
      </Button>
    </div>
  );
}
