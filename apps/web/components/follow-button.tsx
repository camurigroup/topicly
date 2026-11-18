"use client";

import { Loader2, UserCheck, UserPlus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { type FollowResult } from "@/lib/api";

/** Toggles a follow; sends anonymous visitors to the login page. */
export function FollowButton({
  following,
  onToggle,
  onChange,
  size = "sm",
}: {
  following: boolean;
  onToggle: () => Promise<{ data: FollowResult }>;
  onChange: (res: FollowResult) => void;
  size?: "sm" | "default";
}) {
  const { user } = useAuth();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function click() {
    if (!user) {
      router.push("/login");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const { data } = await onToggle();
      onChange(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "falha");
    } finally {
      setBusy(false);
    }
  }

  return (
    <span className="inline-flex flex-col items-end gap-1">
      <Button variant={following ? "outline" : "default"} size={size} onClick={click} disabled={busy} aria-pressed={following}>
        {busy ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : following ? (
          <UserCheck className="h-4 w-4" />
        ) : (
          <UserPlus className="h-4 w-4" />
        )}
        {following ? "Seguindo" : "Seguir"}
      </Button>
      {error && <span className="text-xs text-destructive">{error}</span>}
    </span>
  );
}
