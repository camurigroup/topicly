import Link from "next/link";
import { Fragment } from "react";

import { cn } from "@/lib/utils";

// @handles (3–32 chars) not preceded by a word char, "@" or "." (e-mails), and http(s) URLs.
const TOKEN_RE = /(^|[^a-zA-Z0-9_@.])@([a-zA-Z0-9_]{3,32})\b|(https?:\/\/[^\s<>"']+[^\s<>"'.,;:!?)\]])/g;

/**
 * Renders user content as plain text with @mentions linked to profiles and
 * URLs linked externally. Built from React nodes only (no innerHTML), so user
 * input can never inject markup.
 */
export function RichText({ text, className }: { text: string; className?: string }) {
  const nodes: React.ReactNode[] = [];
  let last = 0;
  let key = 0;
  for (const m of text.matchAll(TOKEN_RE)) {
    const start = m.index ?? 0;
    if (m[2]) {
      const prefix = m[1] ?? "";
      const at = start + prefix.length;
      nodes.push(text.slice(last, at));
      nodes.push(
        <Link key={key++} href={`/u/${m[2]}`} className="font-medium text-primary hover:underline">
          @{m[2]}
        </Link>,
      );
      last = at + 1 + m[2].length;
    } else if (m[3]) {
      nodes.push(text.slice(last, start));
      nodes.push(
        <a
          key={key++}
          href={m[3]}
          target="_blank"
          rel="noopener noreferrer nofollow ugc"
          className="break-all text-primary underline-offset-2 hover:underline"
        >
          {m[3]}
        </a>,
      );
      last = start + m[3].length;
    }
  }
  nodes.push(text.slice(last));

  return (
    <div className={cn("whitespace-pre-wrap break-words", className)}>
      {nodes.map((n, i) => (
        <Fragment key={i}>{n}</Fragment>
      ))}
    </div>
  );
}
