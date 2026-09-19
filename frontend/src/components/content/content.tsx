import { useAutoGrow } from "@/hooks/use-auto-grow";
import { appScrollElement } from "@/lib/dom";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useEffect, useRef, useState } from "react";
import { DEFAULT_PAGE_TITLE, useLocalPagesStore } from "@/store/local-pages";
import { ContentRead } from "./content-read";
import { ContentEditor } from "./content-editor";
import { useContentZoomStore } from "@/store/content-zoom";
import { useDebouncedCallback } from "use-debounce";

export function Content() {
  const activePageStore = useActivePageStore();
  const zoom = useContentZoomStore((state) => state.zoom);
  const canEdit = activePageStore.canEdit();

  return (
    <div
      className="flex flex-1 flex-col gap-4 p-4"
      style={{ "--content-zoom": zoom / 100 } as React.CSSProperties}
    >
      <ContentTitle />
      <ResetScroll />

      {canEdit && <ContentEditor />}
      {!canEdit && <ContentRead />}
    </div>
  );
}

function ResetScroll() {
  const activePageStore = useActivePageStore();

  useEffect(() => {
    requestAnimationFrame(() => {
      appScrollElement()?.scrollTo({ top: 0, behavior: "instant" });
    });
  }, [activePageStore.getPageId()]);

  return null;
}

function ContentTitle() {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const savePageTitle = useLocalPagesStore((state) => state.savePageTitle);

  const activePageStore = useActivePageStore();
  const [title, setTitle] = useState("");

  // The zoom changes the font size without changing any box, so the observer
  // inside the hook cannot see it.
  const zoom = useContentZoomStore((state) => state.zoom);

  useAutoGrow(textareaRef, title, zoom);

  const activePage = activePageStore.page;
  const pageId = activePageStore.getPageId();

  // When active page changed, reset the title
  useEffect(() => {
    if (pageId) {
      if (
        activePage?.__typename === "local_page" &&
        activePage.page.title === DEFAULT_PAGE_TITLE
      ) {
        setTitle("");
        return;
      }

      setTitle(activePage?.page.title || "");
    } else {
      setTitle("");
    }
  }, [pageId]);

  // Auto focus when page title is empty
  useEffect(() => {
    if (!textareaRef.current) return;

    setTimeout(() => {
      const noValue = textareaRef.current?.value.trim() === "";
      if (noValue) {
        textareaRef.current?.focus();
      }
    }, 100);
  }, [pageId]);

  // Save page title when input change
  const setTitleCallback = useDebouncedCallback((title: string) => {
    const activePage = activePageStore.page;

    if (activePage?.__typename !== "local_page" || activePage?.viewOnly) return;
    title = title.trim() === "" ? DEFAULT_PAGE_TITLE : title;

    savePageTitle(title, activePage.page).then((newPage) => {
      // Update active page
      activePageStore.setActivePage({
        ...activePage,
        page: newPage,
      });
    });
  }, 500);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    if (activePage?.viewOnly) return;

    const newTitle = e.target.value;
    setTitle(newTitle);
    setTitleCallback?.(newTitle);
  };

  return (
    <div className="flex justify-center">
      <textarea
        ref={textareaRef}
        placeholder={pageId ? "New Page" : "Select a page or create a new one"}
        readOnly={activePage?.viewOnly || activePageStore.readMode || !pageId}
        value={title}
        onChange={handleChange}
        rows={1}
        className={cn(
          "content-column px-8 sm:px-12",
          "font-bold title-zoom bg-background text-foreground py-2 rounded-md placeholder:text-foreground/30",
          "mb-2 outline-none border-none w-full block",
          "resize-none overflow-hidden",
        )}
      />
    </div>
  );
}
