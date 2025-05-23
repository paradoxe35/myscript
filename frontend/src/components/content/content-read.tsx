import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from "@/components/ui/context-menu";

import { convertNotionToHtml, wrapLists } from "@/lib/notion-to-html";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useMemo } from "react";
import { useContentReadMarker } from "./use-content-read-marker";
import { useContentZoomStore } from "@/store/content-zoom";

export function ContentRead() {
  const zoomStore = useContentZoomStore();
  const activePageStore = useActivePageStore();

  const { containerRef, moveMarker } = useContentReadMarker();

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
    <ContextMenu>
      <ContextMenuTrigger
        disabled={!readMode}
        key={String(readMode)}
        className={cn(
          "px-8 sm:px-12 max-w-[846px]",
          "w-full block mx-auto",
          "prose prose-lg dark:prose-invert",

          // zoomStore.zoom === 80 && "prose-sm",
          // zoomStore.zoom === 90 && "prose-base",
          zoomStore.zoom === 100 && "prose-lg",
          zoomStore.zoom === 200 && "prose-xl",
          zoomStore.zoom === 300 && "prose-2xl"
        )}
        ref={containerRef}
        dangerouslySetInnerHTML={{ __html: html }}
      />

      <ContextMenuContent className="w-40">
        <ContextMenuItem onClick={moveMarker}>Move Marker</ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}
