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

import {
  mutateTree,
  moveItemOnTree,
  type RenderItemParams,
  type TreeItem,
  type TreeData,
  type ItemId,
  type TreeSourcePosition,
  type TreeDestinationPosition,
} from "@atlaskit/tree";
import { useSyncRef } from "@/hooks/use-sync-ref";

const cls = (text: string) => strNormalize(text).toLowerCase();

const rootId = "0";

const defaultPagesTree: TreeData = {
  rootId,
  items: {},
};

function useSidebarItems() {
  const { setOpenMobile } = useSidebar();

  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const notionPagesStore = useNotionPagesStore();

  const [search, setSearch] = useState("");
  const [pagesTree, setPagesTree] = useState<TreeData>(defaultPagesTree);

  const $pagesTree = useSyncRef(pagesTree);

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

  const sortedPages = useMemo(() => {
    return localPagesStore.pages.slice().sort((a, b) => a.order - b.order);
  }, [localPagesStore.pages]);

  useEffect(() => {
    // Create items for all pages
    const items = sortedPages.reduce((acc, page) => {
      const children = sortedPages
        .filter((child) => child.ParentID === page.ID)
        .map((child) => child.ID);

      return {
        ...acc,
        [page.ID]: {
          id: page.ID,
          children: children,
          hasChildren: children.length > 0,
          isExpanded: page.expanded,
          data: page,
        },
      };
    }, {} as TreeData["items"]);

    // Add root item pointing to top-level pages
    const rootChildren = sortedPages
      .filter((page) => !page.ParentID)
      .map((page) => page.ID);

    console.log(rootChildren);

    items[rootId] = {
      id: rootId,
      children: rootChildren,
      hasChildren: rootChildren.length > 0,
      isExpanded: true,
      data: null, // Or add root data if needed
    };

    setPagesTree({ rootId, items });
  }, [sortedPages]);

  const reorderLocalPages = useCallback(
    (
      sourcePosition: TreeSourcePosition,
      destinationPosition?: TreeDestinationPosition
    ) => {
      if (
        !destinationPosition ||
        (sourcePosition.parentId === destinationPosition?.parentId &&
          sourcePosition.index === destinationPosition.index)
      ) {
        return;
      }

      let items = $pagesTree.current.items;

      const sourceItemId = items[sourcePosition.parentId].children.at(
        sourcePosition.index
      );

      const isValidDestination =
        destinationPosition.parentId === rootId ||
        items[destinationPosition.parentId].data?.is_folder;

      if (!sourceItemId || !isValidDestination) {
        return;
      }

      const item: repository.Page = items[sourceItemId].data;

      item.ParentID =
        destinationPosition.parentId === rootId
          ? (null as any)
          : Number(destinationPosition.parentId);

      const newTree = moveItemOnTree(
        $pagesTree.current,
        sourcePosition,
        destinationPosition
      );

      items = newTree.items;

      // Save new pages orders
      let order = 0;
      const applyOrder = (pages: repository.Page[]) => {
        for (let page of pages) {
          if (page.ID === item.ID) {
            page = item;
          }

          page.order = ++order;
          localPagesStore.saveNewPageOrder(page);

          const pageChildren = items[page.ID].children.map((id) => {
            return items[id].data as repository.Page;
          });

          if (pageChildren.length > 0) {
            applyOrder(pageChildren);
          }
        }
      };

      const rootChildren = items[rootId].children.map((id) => {
        return items[id].data as repository.Page;
      });

      applyOrder(rootChildren);

      setPagesTree(newTree);
    },
    [sortedPages]
  );

  return useMemo(() => {
    return {
      search,
      setSearch,

      pagesTree,
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
    pagesTree,
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
