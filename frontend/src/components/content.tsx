import debounce from "lodash/debounce";
import { EditorJS } from "@/components/editorjs";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useEffect, useMemo, useRef, useState } from "react";
import { DEFAULT_PAGE_TITLE, useLocalPagesStore } from "@/store/local-pages";

export function Content() {
  const activePageStore = useActivePageStore();
  const activePage = activePageStore.page;

  return (
    <div className="flex flex-1 flex-col gap-4 p-4">
      <ContentTitle />

      <EditorJS />
    </div>
  );
}

function ContentTitle() {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const savePageTitle = useLocalPagesStore((state) => state.savePageTitle);

  const activePageStore = useActivePageStore();
  const [title, setTitle] = useState("");

  const activePage = activePageStore.page;

  const pageId =
    activePage?.__typename === "local_page"
      ? activePage.page.ID
      : activePage?.page.id;

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
        textareaRef.current.focus();
      }
    }, 100);
  }, [pageId]);

  // Auto grow textarea height
  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
    }
  }, [title]);

  // Save page title when input change
  const setTitleCallback = useMemo(() => {
    if (activePage?.__typename !== "local_page" || activePage?.readMode) return;

    return debounce((title: string) => {
      title = title.trim() === "" ? DEFAULT_PAGE_TITLE : title;

      savePageTitle(title, activePage.page).then((newPage) => {
        // Update active page
        activePageStore.setActivePage({
          ...activePage,
          page: newPage,
        });
      });
    }, 500);
  }, [activePage]);

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    if (activePage?.readMode) return;

    const newTitle = e.target.value;
    setTitle(newTitle);
    setTitleCallback?.(newTitle);
  };

  return (
    <div className="flex justify-center">
      <textarea
        ref={textareaRef}
        placeholder={pageId ? "New Page" : "Select a page or create a new one"}
        readOnly={activePage?.readMode || !pageId}
        value={title}
        onChange={handleChange}
        rows={1}
        className={cn(
          "font-bold text-3xl bg-background text-foreground py-2 rounded-md placeholder:text-foreground/30",
          "max-w-[650px] mb-2 justify-self-center outline-none border-none w-full block",
          "resize-none overflow-hidden"
        )}
      />
    </div>
  );
}
