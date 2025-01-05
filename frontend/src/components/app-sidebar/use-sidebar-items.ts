import { useActivePageStore } from "@/store/active-page";
import { useLocalPagesStore } from "@/store/local-pages";
import { useNotionPagesStore } from "@/store/notion-pages";
import { NotionSimplePage } from "@/types";
import { useEffect, useMemo, useState } from "react";
import { repository } from "~wails/models";
import { useSidebar } from "../ui/sidebar";

export function useSidebarItems() {
  const { setOpenMobile } = useSidebar();

  const activePageStore = useActivePageStore();
  const localPagesStore = useLocalPagesStore();
  const notionPagesStore = useNotionPagesStore();

  const [search, setSearch] = useState("");

  const activePage = activePageStore.page;

  useEffect(() => {
    localPagesStore.getPages();
  }, []);

  const createNewPage = async () => {
    const newPage = await localPagesStore.newPage();

    activePageStore.setActivePage({
      __typename: "local_page",
      page: newPage,
      viewOnly: false,
    });

    setOpenMobile(false);
  };

  const onLocalPageClick = (page: repository.Page) => {
    activePageStore.setActivePage({
      __typename: "local_page",
      viewOnly: false,
      page,
    });

    activePageStore.fetchPageBlocks();
    setOpenMobile(false);
  };

  const onNotionPageClick = (page: NotionSimplePage) => {
    activePageStore.setActivePage({
      __typename: "notion_page",
      viewOnly: true,
      page,
    });

    activePageStore.fetchPageBlocks();
    setOpenMobile(false);
  };

  const refreshNotionPages = () => {
    notionPagesStore.getPages();
    if (activePage?.__typename === "notion_page") {
      activePageStore.fetchPageBlocks();
    }
  };

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
        return page.title.toLowerCase().includes(word.toLowerCase());
      });
    });
  }, [notionPagesStore.pages, search]);

  const localPages = useMemo(() => {
    const pages = localPagesStore.pages;
    const searchValue = search.trim();

    if (!searchValue) {
      return pages;
    }

    return pages.filter((page) => {
      const words = searchValue.split(" ");
      return words.every((word) => {
        return page.title.toLowerCase().includes(word.toLowerCase());
      });
    });
  }, [localPagesStore.pages, search]);

  return {
    search,
    setSearch,

    localPages,
    notionPages,
    activePage,
    createNewPage,
    onLocalPageClick,
    refreshNotionPages,
    onNotionPageClick,
  };
}
