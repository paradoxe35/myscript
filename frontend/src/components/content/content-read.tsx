import edjsHTML from "editorjs-html";
import { convertNotionToHtml, wrapLists } from "@/lib/notion-to-html";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useMemo } from "react";
import { useContentReadMarker } from "./use-content-read-marker-v2";

const edjsParser = edjsHTML();

export function ContentRead() {
  const containerRef = useContentReadMarker();
  const activePageStore = useActivePageStore();

  const activePage = activePageStore.page;
  const readMode = activePageStore.readMode;

  const html = useMemo(() => {
    const blocks = activePage?.blocks || [];

    switch (activePage?.__typename) {
      case "notion_page":
        return wrapLists(convertNotionToHtml(blocks));

      case "local_page":
        const result = edjsParser.parseStrict({ blocks });
        return Array.isArray(result) ? result.join("\n") : result.message;
    }

    return "";
  }, [activePage]);

  return (
    <div
      key={String(readMode)}
      className={cn(
        "w-full block mx-auto",
        "prose max-w-[650px] dark:prose-invert"
      )}
      ref={containerRef}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
