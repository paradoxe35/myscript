import { WithoutRepositoryBaseFields } from "@/types";
import { create } from "zustand";
import { DeleteLocalPage, GetLocalPages, SaveLocalPage } from "~wails/main/App";
import { repository } from "~wails/models";

type LocalPagesStore = {
  pages: Array<repository.Page>;
  getPages: () => Promise<void>;
  newPage: () => Promise<repository.Page>;
  newFolder: (name: string) => Promise<repository.Page>;
  savePageTitle: (
    title: string,
    page: repository.Page
  ) => Promise<repository.Page>;
  savePageBlocks: (
    page: repository.Page,
    blocks: any,
    htmlContent: string
  ) => Promise<repository.Page>;
  deletePage: (id: number) => Promise<void>;
};

type TPage = WithoutRepositoryBaseFields<repository.Page>;

export const DEFAULT_PAGE_TITLE = "New Page";

export const DEFAULT_FOLDER_TITLE = "New Folder";

export const useLocalPagesStore = create<LocalPagesStore>((set) => ({
  pages: [],

  getPages: async () => {
    const pages = await GetLocalPages();
    set({ pages });
  },

  newPage: async () => {
    const body: TPage = {
      title: DEFAULT_PAGE_TITLE,
      blocks: [],
      html_content: "",
      is_folder: false,
      order: 0,
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    set({ pages: await GetLocalPages() });

    return newPage;
  },

  newFolder: async (name: string) => {
    const body: TPage = {
      title: name || DEFAULT_FOLDER_TITLE,
      blocks: [],
      html_content: "",
      is_folder: true,
      order: 0,
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    console.log(await GetLocalPages());

    set({ pages: await GetLocalPages() });

    return newPage;
  },

  savePageTitle: async (title, page) => {
    const body: TPage = {
      ...page,
      title,
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    set({ pages: await GetLocalPages() });

    return newPage;
  },

  savePageBlocks: async (page, blocks, htmlContent) => {
    const body: TPage = {
      ...page,
      blocks,
      html_content: htmlContent,
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    const pages = await GetLocalPages();
    set({ pages });

    return newPage;
  },

  deletePage: async (id) => {
    await DeleteLocalPage(id);

    const pages = await GetLocalPages();
    set({ pages });
  },
}));
