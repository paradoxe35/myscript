import { convertNotionToHtml, wrapLists } from "@/lib/notion-to-html";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useMemo } from "react";
import { useContentReadMarker } from "./use-content-read-marker";
import { useContentZoomStore } from "@/store/content-zoom";

export function ContentRead() {
  const zoomStore = useContentZoomStore();

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
        return activePage.page.html_content;
    }

    return "";
  }, [activePage]);

  return (
    <div
      key={String(readMode)}
      className={cn(
        "w-full block mx-auto",
        "prose prose-lg max-w-[750px] dark:prose-invert",

        // zoomStore.zoom === 80 && "prose-sm",
        // zoomStore.zoom === 90 && "prose-base",
        zoomStore.zoom === 100 && "prose-lg",
        zoomStore.zoom === 200 && "prose-xl",
        zoomStore.zoom === 300 && "prose-2xl"
      )}
      ref={containerRef}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  );
}
