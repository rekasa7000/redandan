import { z } from "zod";

export const taskFormSchema = z.object({
  title: z.string().min(1, "Title is required"),
  description: z.string().optional(),
  context_id: z.string().optional(),
  priority: z.enum(["low", "medium", "high", "urgent"]).optional(),
  status: z.enum(["todo", "in_progress", "done", "archived"]).optional(),
  due_date: z.string().optional(),
  tags: z.array(z.string()).optional(),
});

export type TaskFormValues = z.infer<typeof taskFormSchema>;
