import { useActivePageStore } from "@/store/active-page";
import { useLocalPagesStore } from "@/store/local-pages";
import { useNotionPagesStore } from "@/store/notion-pages";
import { NotionSimplePage } from "@/types";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { repository } from "~wails/models";
import { useSidebar } from "../ui/sidebar";
import { strNormalize } from "@/lib/string";
import { OnDragEndResponder } from "@hello-pangea/dnd";

function useSidebarItems() {
  const { setOpenMobile } = useSidebar();

  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const notionPagesStore = useNotionPagesStore();

  const [search, setSearch] = useState("");

  const activePage = activePageStore.page;

  useEffect(() => {
    localPagesStore.getPages();
  }, []);

  const createNewPage = useCallback(async () => {
    const newPage = await localPagesStore.newPage();

    activePageStore.setActivePage({
      __typename: "local_page",
      page: newPage,
      viewOnly: false,
    });

    setOpenMobile(false);
  }, [localPagesStore, activePageStore]);

  const togglePageExpanded = useCallback(
    async (page: repository.Page) => {
      await localPagesStore.togglePageExpanded(page);
    },
    [localPagesStore]
  );

  const onLocalPageClick = useCallback(
    (page: repository.Page) => {
      activePageStore.setActivePage({
        __typename: "local_page",
        viewOnly: false,
        page,
      });

      activePageStore.fetchPageBlocks();
      setOpenMobile(false);
    },
    [activePageStore]
  );

  const onNotionPageClick = useCallback(
    (page: NotionSimplePage) => {
      activePageStore.setActivePage({
        __typename: "notion_page",
        viewOnly: true,
        page,
      });

      activePageStore.fetchPageBlocks();
      setOpenMobile(false);
    },
    [activePageStore]
  );

  const refreshNotionPages = useCallback(() => {
    notionPagesStore.getPages();
    if (activePage?.__typename === "notion_page") {
      activePageStore.fetchPageBlocks();
    }
  }, [notionPagesStore, activePageStore, activePage]);

  // Refresh notion pages when the user is online
  useEffect(() => {
    addEventListener("online", refreshNotionPages);

    return () => {
      removeEventListener("online", refreshNotionPages);
    };
  }, []);

  const cls = useCallback((text: string) => {
    return strNormalize(text).toLowerCase();
  }, []);

  const notionPages = useMemo(() => {
    const pages = notionPagesStore.getSimplifiedPages();
    const searchValue = search.trim();

    if (!searchValue) {
      return pages;
    }

    return pages.filter((page) => {
      const words = searchValue.split(" ");
      return words.every((word) => {
        return cls(page.title).includes(cls(word));
      });
    });
  }, [notionPagesStore.pages, search]);

  const localPages = useMemo(() => {
    const pages = localPagesStore.pages
      .slice()
      .sort((a, b) => a.order - b.order);

    const searchValue = search.trim();

    if (!searchValue) {
      return pages;
    }

    return pages.filter((page) => {
      const words = searchValue.split(" ");
      return words.every((word) => {
        return cls(page.title).includes(cls(word));
      });
    });
  }, [localPagesStore.pages, search]);

  const reorderLocalPages = useCallback<OnDragEndResponder<string>>(
    (result) => {
      const { source, destination } = result;
      // Drop outside the list or no movement
      if (!destination || source.index === destination.index) {
        return;
      }

      localPagesStore.reorderPages(localPages, source.index, destination.index);
    },
    [localPagesStore, localPages]
  );

  return useMemo(() => {
    return {
      search,
      setSearch,

      localPages,
      notionPages,
      activePage,
      createNewPage,
      togglePageExpanded,
      onLocalPageClick,
      refreshNotionPages,
      onNotionPageClick,

      reorderLocalPages,
    };
  }, [
    search,
    setSearch,
    localPages,
    notionPages,
    activePage,
    createNewPage,
    togglePageExpanded,
    onLocalPageClick,
    refreshNotionPages,
    onNotionPageClick,
    localPagesStore,
    reorderLocalPages,
  ]);
}

type SidebarContext = ReturnType<typeof useSidebarItems>;

const SidebarContext = createContext<SidebarContext | null>(null);

export function useSidebarItemsContext() {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebarContext must be used within a SidebarProvider.");
  }
  return context;
}

export function SidebarItemsProvider(props: React.PropsWithChildren<{}>) {
  return (
    <SidebarContext.Provider value={useSidebarItems()}>
      {props.children}
    </SidebarContext.Provider>
  );
}
