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
import clone from "lodash/clone";

const cls = (text: string) => strNormalize(text).toLowerCase();

function locatePage(
  pageId: string | number,
  pages: repository.Page[]
): repository.Page | null {
  for (const page of pages) {
    if (page.ID === Number(pageId)) {
      return page;
    }

    const child = locatePage(pageId, page.Children || []);

    if (child) {
      return child;
    }
  }

  return null;
}

function useSidebarItems() {
  const { setOpenMobile } = useSidebar();

  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const notionPagesStore = useNotionPagesStore();

  const [search, setSearch] = useState("");
  const [localPages, setLocalPages] = useState<repository.Page[]>([]);

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

  useEffect(() => {
    const sortedPages = localPagesStore.pages
      .slice()
      .sort((a, b) => a.order - b.order);

    const pages = sortedPages.reduce((acc, page) => {
      acc[page.ID] = page;
      return acc;
    }, {} as Record<number, repository.Page>);

    const group = (parentId: number | null) => {
      const $pages = sortedPages.filter((page) => {
        if (pages[page.ID] && page.ParentID === parentId) {
          delete pages[page.ID];
          return true;
        }
        return false;
      });

      return $pages.map((page) => {
        const $page = clone(page);
        $page.Children = group(page.ID).sort((a, b) => a.order - b.order);
        return $page;
      });
    };

    setLocalPages(group(null));
  }, [localPagesStore.pages]);

  const getPageChildren = useCallback(
    (pageId: string, groupedPages: repository.Page[]) => {
      if (pageId === "root") {
        return groupedPages;
      }

      return locatePage(pageId, groupedPages);
    },
    []
  );

  const reorderLocalPages = useCallback<OnDragEndResponder<string>>(
    (result) => {
      const { source, destination } = result;
      // Drop outside the list or no movement
      if (
        !destination ||
        (source.index === destination.index &&
          source.droppableId === destination.droppableId)
      ) {
        return;
      }

      let groupedPages = localPages.slice();

      const draggableId = result.draggableId;
      const sourceCategoryId = source.droppableId;
      const destinationCategoryId = destination.droppableId;

      const sourcePage = locatePage(draggableId, groupedPages);
      if (!sourcePage) return;

      sourcePage.ParentID =
        destinationCategoryId === "root"
          ? (null as unknown as undefined)
          : Number(destinationCategoryId);

      const destinationItem = getPageChildren(
        destinationCategoryId,
        groupedPages
      );
      const sourceItem = getPageChildren(sourceCategoryId, groupedPages);

      if (!sourceItem || !destinationItem) return;

      // Source operation
      if (Array.isArray(sourceItem)) {
        groupedPages = sourceItem;
        groupedPages.splice(source.index, 1);
      } else {
        sourceItem.Children.splice(source.index, 1);
      }

      // Destination operation
      if (Array.isArray(destinationItem)) {
        groupedPages = destinationItem;
        groupedPages.splice(destination.index, 0, sourcePage);
      } else {
        destinationItem.Children.splice(destination.index, 0, sourcePage);
      }

      // Save new pages orders
      let order = 0;
      const applyOrder = (pages: repository.Page[]) => {
        for (const page of pages) {
          order++;
          page.order = order;
          localPagesStore.saveNewPageOrder(page);
          if (page.Children) {
            applyOrder(page.Children);
          }
        }
      };

      applyOrder(groupedPages);
      setLocalPages(groupedPages);
    },
    [localPagesStore, localPages, getPageChildren]
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

function useSidebarLocalPages() {}

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
