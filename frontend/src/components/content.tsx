import debounce from "lodash/debounce";
import { EditorJS, type API } from "@/components/editorjs";
import { cn } from "@/lib/utils";
import { useActivePageStore } from "@/store/active-page";
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { DEFAULT_PAGE_TITLE, useLocalPagesStore } from "@/store/local-pages";
import { useContentZoomStore } from "@/store/content-zoom";
import { ContentRead } from "./content-read";

export function Content() {
  const $content = useRef<HTMLDivElement>(null);
  const activePageStore = useActivePageStore();
  const zoomStore = useContentZoomStore();

  useLayoutEffect(() => {
    if ($content.current) {
      // @ts-ignore
      $content.current.style.zoom = `${zoomStore.zoom / 100}`;
    }
  }, [zoomStore.zoom]);

  const canEdit = activePageStore.canEdit();

  return (
    <div className="flex flex-1 flex-col gap-4 p-4" ref={$content}>
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
      // @ts-ignore
      document.body.scrollTo({ top: 0, behavior: "instant" });
    });
  }, [activePageStore.getPageId()]);

  return null;
}

function ContentEditor() {
  const activePageStore = useActivePageStore();
  const savePageBlocks = useLocalPagesStore((state) => state.savePageBlocks);

  const activePage = activePageStore.page;

  const setBlocksCallback = useMemo(() => {
    if (activePage?.__typename !== "local_page" || activePage?.viewOnly) return;

    return debounce(async (editorAPI: API) => {
      const output = await editorAPI.saver.save().catch(() => ({ blocks: [] }));

      savePageBlocks(output.blocks, activePage.page).then((newPage) => {
        activePageStore.setActivePage({
          ...activePage,
          blocks: output.blocks,
        });
      });
    }, 500);
  }, [activePage]);

  return (
    <EditorJS
      key={activePageStore.version}
      defaultBlocks={activePage?.blocks || []}
      onChange={setBlocksCallback}
    />
  );
}

function ContentTitle() {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const savePageTitle = useLocalPagesStore((state) => state.savePageTitle);

  const activePageStore = useActivePageStore();
  const [title, setTitle] = useState("");

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
    if (activePage?.__typename !== "local_page" || activePage?.viewOnly) return;

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
          "font-bold text-3xl bg-background text-foreground py-2 rounded-md placeholder:text-foreground/30",
          "max-w-[650px] mb-2 justify-self-center outline-none border-none w-full block",
          "resize-none overflow-hidden"
        )}
      />
    </div>
  );
}
