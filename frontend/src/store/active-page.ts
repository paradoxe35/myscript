import { NotionSimplePage } from "@/types";
import { create } from "zustand";
import { repository } from "~wails/models";
import { persist, createJSONStorage } from "zustand/middleware";

type NotionActivePage = {
  __typename: "notion_page";
  page: NotionSimplePage;
  readMode: true;
  blocks?: Array<any>;
};

type LocalActivePage = {
  __typename: "local_page";
  page: repository.Page;
  readMode: boolean;
  blocks?: Array<any>;
};

type ActivePage = NotionActivePage | LocalActivePage;

type ActivePageStore = {
  page: ActivePage | null;
  setActivePage: (page: ActivePage) => void;
};

export const useActivePageStore = create(
  persist<ActivePageStore>(
    (set) => ({
      page: null,

      setActivePage: (page: ActivePage) => {
        set({ page });
      },
    }),

    {
      name: "active-page",
      storage: createJSONStorage(() => localStorage),
    }
  )
);
