/** First-run installer client. */
import { raw, type AuthResult } from "@/lib/api";

export interface InstallStatus {
  installed: boolean;
  version: string;
  installedAt: string | null;
  tokenRequired: boolean;
}

export interface InstallInput {
  token?: string;
  forumName: string;
  description?: string;
  username: string;
  email: string;
  password: string;
  displayName?: string;
}

export const installApi = {
  status: () => raw<InstallStatus>("/api/v1/install/status", {}),
  complete: (input: InstallInput) => raw<AuthResult>("/api/v1/install", { method: "POST", body: input }),
};
