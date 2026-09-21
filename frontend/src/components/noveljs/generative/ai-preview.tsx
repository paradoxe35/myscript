import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import Markdown from "react-markdown";

/**
 * Renders the Markdown answer. The height cap goes on the viewport; on the
 * root it would only clip a long answer.
 */
export function AIPreview({ markdown }: { markdown: string }) {
  return (
    <ScrollArea className="border-b" viewportClassName="max-h-[320px]">
      <div
        className={cn(
          "prose prose-sm max-w-none px-4 py-3 dark:prose-invert",
          "prose-headings:mt-3 prose-headings:mb-1 prose-p:my-1.5",
          "prose-ul:my-1.5 prose-ol:my-1.5 prose-li:my-0.5 prose-pre:my-2",
        )}
      >
        <Markdown>{markdown}</Markdown>
      </div>
    </ScrollArea>
  );
}
