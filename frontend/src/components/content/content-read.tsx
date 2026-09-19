import { convertNotionToHtml, wrapLists } from "@/lib/notion-to-html";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useMemo } from "react";
import { useContentReader } from "./use-content-reader";

export function ContentRead() {
  const activePageStore = useActivePageStore();

  const activePage = activePageStore.page;
  const readMode = activePageStore.readMode;

  const html = useMemo(() => {
    const blocks = activePage?.blocks || [];

    switch (activePage?.__typename) {
      case "notion_page":
        return wrapLists(convertNotionToHtml(blocks));

      case "local_page":
        return activePage.page.html_content;
    }

    return "";
  }, [activePage]);

  const { containerRef, onClick } = useContentReader(html);

  return (
    <div
      key={String(readMode)}
      ref={containerRef}
      onClick={onClick}
      className={cn(
        "content-column px-8 sm:px-12 block",
        "prose prose-lg dark:prose-invert prose-zoom",
        readMode && "reader"
      )}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
