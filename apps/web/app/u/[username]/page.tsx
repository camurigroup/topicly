"use client";

import { CalendarDays, Loader2, Star } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { Card, CardContent } from "@/components/ui/card";
import { api, type PublicProfile } from "@/lib/api";

type State =
  | { kind: "loading" }
  | { kind: "ready"; profile: PublicProfile }
  | { kind: "error"; message: string };

export default function ProfilePage() {
  const params = useParams<{ username: string }>();
  const username = params.username;
  const [state, setState] = useState<State>({ kind: "loading" });

  useEffect(() => {
    let active = true;
    setState({ kind: "loading" });
    api
      .publicProfile(username)
      .then(({ data }) => active && setState({ kind: "ready", profile: data }))
      .catch((err) =>
        active && setState({ kind: "error", message: err instanceof Error ? err.message : "erro" }),
      );
    return () => {
      active = false;
    };
  }, [username]);

  if (state.kind === "loading") {
    return (
      <div className="flex justify-center py-20 text-muted-foreground">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (state.kind === "error") {
    return (
      <div className="mx-auto max-w-xl text-center text-muted-foreground">
        <p>Perfil não encontrado.</p>
      </div>
    );
  }

  const { profile } = state;
  const joined = new Date(profile.createdAt).toLocaleDateString("pt-BR", {
    year: "numeric",
    month: "long",
  });

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Card className="overflow-hidden">
        <div className="h-28 w-full bg-secondary">
          {profile.bannerUrl && (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={profile.bannerUrl} alt="" className="h-full w-full object-cover" />
          )}
        </div>
        <CardContent className="pt-0">
          <div className="-mt-10 flex items-end gap-4">
            <div className="h-20 w-20 overflow-hidden rounded-full border-4 border-card bg-muted">
              {profile.avatarUrl ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={profile.avatarUrl} alt={profile.displayName} className="h-full w-full object-cover" />
              ) : (
                <div className="flex h-full w-full items-center justify-center text-2xl font-semibold text-muted-foreground">
                  {profile.displayName.charAt(0).toUpperCase()}
                </div>
              )}
            </div>
            <div className="pb-1">
              <h1 className="text-xl font-bold leading-tight">{profile.displayName}</h1>
              <p className="text-sm text-muted-foreground">@{profile.username}</p>
            </div>
          </div>

          {profile.bio && <p className="mt-4 text-sm">{profile.bio}</p>}

          <div className="mt-4 flex flex-wrap gap-4 text-sm text-muted-foreground">
            <span className="inline-flex items-center gap-1.5">
              <Star className="h-4 w-4" /> {profile.reputation} de reputação
            </span>
            <span className="inline-flex items-center gap-1.5">
              <CalendarDays className="h-4 w-4" /> Entrou em {joined}
            </span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
