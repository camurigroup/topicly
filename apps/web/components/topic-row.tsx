import { Heart, MessageSquare, Pin, Eye, Lock } from "lucide-react";
import Link from "next/link";

import { type Topic } from "@/lib/api";
import { cn, formatCount, timeAgo } from "@/lib/utils";

/** A single topic in a list, linking to its detail page. */
export function TopicRow({ topic }: { topic: Topic }) {
  return (
    <li className="rounded-lg border border-border bg-card p-4 text-card-foreground">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <Link href={`/t/${topic.id}`} className="font-medium hover:underline">
            <span className="inline-flex items-center gap-1.5">
              {topic.isPinned && <Pin className="h-3.5 w-3.5 text-primary" />}
              {topic.isLocked && <Lock className="h-3.5 w-3.5 text-muted-foreground" />}
              {topic.title}
            </span>
          </Link>
          <p className="mt-1 text-xs text-muted-foreground">
            por{" "}
            <Link href={`/u/${topic.author.username}`} className="hover:underline">
              {topic.author.displayName}
            </Link>{" "}
            · {timeAgo(topic.lastActivityAt)}
          </p>
        </div>
        <div className="flex shrink-0 items-center gap-3 text-xs text-muted-foreground">
          <Stat icon={<MessageSquare className="h-3.5 w-3.5" />} value={topic.replyCount} />
          <Stat icon={<Heart className={cn("h-3.5 w-3.5", topic.liked && "fill-current text-primary")} />} value={topic.likeCount} />
          <Stat icon={<Eye className="h-3.5 w-3.5" />} value={topic.viewCount} />
        </div>
      </div>
    </li>
  );
}

function Stat({ icon, value }: { icon: React.ReactNode; value: number }) {
  return (
    <span className="inline-flex items-center gap-1">
      {icon}
      {formatCount(value)}
    </span>
  );
}
