"use client";

import { useRouter } from "next/navigation";
import { clearTokens } from "@/lib/api";
import { Button } from "@/components/ui/button";

export function Header() {
  const router = useRouter();

  function handleLogout() {
    clearTokens();
    router.push("/login");
  }

  return (
    <header className="border-b px-6 py-3 flex items-center justify-between">
      <h1 className="text-lg font-semibold">Todo App</h1>
      <Button variant="outline" size="sm" onClick={handleLogout}>
        Logout
      </Button>
    </header>
  );
}
