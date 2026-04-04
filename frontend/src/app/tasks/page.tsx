"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import {
  isLoggedIn,
  listTasks,
  createTask,
  updateTask,
  deleteTask,
  type Task,
} from "@/lib/api";
import { Header } from "@/components/header";
import { TaskCard } from "@/components/task-card";
import { TaskFormDialog } from "@/components/task-form";
import { Button } from "@/components/ui/button";

export default function TasksPage() {
  const router = useRouter();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showCreate, setShowCreate] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  const fetchTasks = useCallback(async () => {
    try {
      const data = await listTasks();
      setTasks(data.items ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load tasks");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isLoggedIn()) {
      router.push("/login");
      return;
    }
    fetchTasks();
  }, [router, fetchTasks]);

  async function handleCreate(data: {
    title: string;
    description: string;
    due_date: number;
  }) {
    await createTask(data.title, data.description, data.due_date);
    await fetchTasks();
  }

  async function handleUpdate(data: {
    title: string;
    description: string;
    due_date: number;
  }) {
    if (!editingTask) return;
    await updateTask(editingTask.id, editingTask.version, data);
    setEditingTask(null);
    await fetchTasks();
  }

  async function handleDelete(task: Task) {
    if (!confirm(`Delete "${task.title}"?`)) return;
    try {
      await deleteTask(task.id, task.version);
      await fetchTasks();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete");
    }
  }

  return (
    <div className="flex flex-col min-h-full">
      <Header />
      <main className="flex-1 max-w-2xl mx-auto w-full p-6">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-semibold">Tasks</h2>
          <Button onClick={() => setShowCreate(true)}>New Task</Button>
        </div>

        {error && <p className="text-sm text-destructive mb-4">{error}</p>}

        {loading ? (
          <p className="text-muted-foreground">Loading...</p>
        ) : tasks.length === 0 ? (
          <p className="text-muted-foreground">No tasks yet.</p>
        ) : (
          <div className="space-y-3">
            {tasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                onEdit={setEditingTask}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}

        <TaskFormDialog
          open={showCreate}
          onClose={() => setShowCreate(false)}
          onSubmit={handleCreate}
        />

        {editingTask && (
          <TaskFormDialog
            open={true}
            onClose={() => setEditingTask(null)}
            onSubmit={handleUpdate}
            task={editingTask}
          />
        )}
      </main>
    </div>
  );
}
