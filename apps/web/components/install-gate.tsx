"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

import { installApi } from "@/lib/install-api";

/**
 * Sends visitors to the installer while the forum has not been installed
 * yet, and keeps the installer unreachable once it has.
 */
export function InstallGate() {
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    let active = true;
    installApi
      .status()
      .then(({ data }) => {
        if (!active) return;
        if (!data.installed && pathname !== "/install") router.replace("/install");
        if (data.installed && pathname === "/install") router.replace("/");
      })
      .catch(() => {
        /* API offline: leave the page as is */
      });
    return () => {
      active = false;
    };
  }, [pathname, router]);

  return null;
}
