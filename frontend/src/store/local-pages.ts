import { create } from "zustand";
import { GetLocalPages, SaveLocalPage } from "~wails/main/App";
import { repository } from "~wails/models";

type LocalPagesStore = {
  pages: Array<repository.Page>;
  getPages: () => Promise<void>;
  newPage: () => Promise<repository.Page>;
};

export const useLocalPagesStore = create<LocalPagesStore>((set) => ({
  pages: [],

  getPages: async () => {
    const pages = await GetLocalPages();
    set({ pages });
  },

  newPage: async () => {
    const newPage = await SaveLocalPage(
      repository.Page.createFrom({
        title: "New Page",
        blocks: [],
      })
    );

    const pages = await GetLocalPages();
    set({ pages });

    return newPage;
  },
}));
