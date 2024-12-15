import { create } from "zustand";
import { GetLocalPages, SaveLocalPage } from "~wails/main/App";
import { repository } from "~wails/models";

type LocalPagesStore = {
  pages: Array<repository.Page>;
  getPages: () => Promise<void>;
  newPage: () => Promise<repository.Page>;
  savePageTitle: (
    title: string,
    page: repository.Page
  ) => Promise<repository.Page>;
};

export const DEFAULT_PAGE_TITLE = "New Page";

export const useLocalPagesStore = create<LocalPagesStore>((set) => ({
  pages: [],

  getPages: async () => {
    const pages = await GetLocalPages();
    set({ pages });
  },

  newPage: async () => {
    const newPage = await SaveLocalPage(
      repository.Page.createFrom({
        title: DEFAULT_PAGE_TITLE,
        blocks: [],
      })
    );

    const pages = await GetLocalPages();
    set({ pages });

    return newPage;
  },

  savePageTitle: async (title, page) => {
    const newPage = await SaveLocalPage(
      repository.Page.createFrom({
        ...page,
        title,
      })
    );

    const pages = await GetLocalPages();
    set({ pages });

    return newPage;
  },
}));
