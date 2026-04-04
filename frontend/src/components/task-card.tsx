"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Task } from "@/lib/api";

type Props = {
  task: Task;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
};

export function TaskCard({ task, onEdit, onDelete }: Props) {
  const dueDate = task.due_date
    ? new Date(task.due_date).toLocaleDateString("ja-JP")
    : "-";

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <CardTitle className="text-base">{task.title}</CardTitle>
          <span className="text-xs text-muted-foreground px-2 py-1 rounded bg-muted">
            {task.status}
          </span>
        </div>
      </CardHeader>
      <CardContent>
        {task.description && (
          <p className="text-sm text-muted-foreground mb-3">
            {task.description}
          </p>
        )}
        <div className="flex items-center justify-between">
          <span className="text-xs text-muted-foreground">Due: {dueDate}</span>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => onEdit(task)}>
              Edit
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => onDelete(task)}
            >
              Delete
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
