"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Context } from "@/lib/api";

const ALL = "all";

export function TaskFilters({
  contexts,
  status,
  contextId,
  onChange,
}: {
  contexts: Context[];
  status: string;
  contextId: string;
  onChange: (filters: { status: string; contextId: string }) => void;
}) {
  return (
    <div className="flex flex-wrap gap-2">
      <Select
        value={status || ALL}
        onValueChange={(value) => onChange({ status: value === ALL ? "" : value, contextId })}
      >
        <SelectTrigger>
          <SelectValue placeholder="Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All statuses</SelectItem>
          <SelectItem value="todo">To do</SelectItem>
          <SelectItem value="in_progress">In progress</SelectItem>
          <SelectItem value="done">Done</SelectItem>
          <SelectItem value="archived">Archived</SelectItem>
        </SelectContent>
      </Select>

      <Select
        value={contextId || ALL}
        onValueChange={(value) => onChange({ status, contextId: value === ALL ? "" : value })}
      >
        <SelectTrigger>
          <SelectValue placeholder="Context" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All contexts</SelectItem>
          {contexts.map((context) => (
            <SelectItem key={context.id} value={context.id ?? ""}>
              {context.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
