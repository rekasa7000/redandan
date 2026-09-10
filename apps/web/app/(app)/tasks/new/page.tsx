"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { TaskForm } from "@/components/tasks/task-form";
import { ApiError, createTask, listContexts, type Context, type TaskPayload } from "@/lib/api";
import { clearToken } from "@/lib/session";

export default function NewTaskPage() {
  const router = useRouter();
  const [contexts, setContexts] = useState<Context[]>([]);

  useEffect(() => {
    listContexts()
      .then(setContexts)
      .catch((err) => {
        if (err instanceof ApiError && err.status === 401) {
          clearToken();
          router.replace("/login");
        }
      });
  }, [router]);

  async function handleSubmit(values: TaskPayload) {
    const task = await createTask(values);
    router.push(task.id ? `/tasks/${task.id}` : "/tasks");
  }

  return (
    <div className="mx-auto max-w-lg p-6">
      <Card>
        <CardHeader>
          <CardTitle>New task</CardTitle>
        </CardHeader>
        <CardContent>
          <TaskForm contexts={contexts} onSubmit={handleSubmit} submitLabel="Create task" />
        </CardContent>
      </Card>
    </div>
  );
}
