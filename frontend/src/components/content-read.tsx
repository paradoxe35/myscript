import { convertNotionToHtml, wrapLists } from "@/lib/notion-to-html";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useMemo } from "react";

export function ContentRead() {
  const activePage = useActivePageStore((state) => state.page);

  const html = useMemo(() => {
    const blocks = activePage?.blocks || [];

    switch (activePage?.__typename) {
      case "notion_page":
        return wrapLists(convertNotionToHtml(blocks));

      case "local_page":
        return "";
    }

    return "";
  }, [activePage]);

  return (
    <div className="flex justify-center">
      <div
        className={cn("prose max-w-[650px] dark:prose-invert")}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </div>
  );
}
