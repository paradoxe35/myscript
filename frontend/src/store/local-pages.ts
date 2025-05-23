import { WithoutRepositoryBaseFields } from "@/types";
import { create } from "zustand";
import {
  DeleteLocalPage,
  GetLocalPage,
  GetLocalPages,
  SaveLocalPage,
  UpdateLocalPageOrder,
} from "~wails/main/App";
import { repository } from "~wails/models";

type LocalPagesStore = {
  pages: Array<repository.Page>;
  getPages: () => Promise<void>;
  getPage: (ID: string) => Promise<repository.Page>;
  newPage: () => Promise<repository.Page>;
  newFolder: (name: string) => Promise<repository.Page>;
  togglePageExpanded: (page: repository.Page) => Promise<void>;
  savePage: (page: repository.Page) => Promise<repository.Page>;
  saveNewPageOrder: (page: repository.Page) => Promise<void>;
  savePageTitle: (
    title: string,
    page: repository.Page
  ) => Promise<repository.Page>;
  savePageBlocks: (
    page: repository.Page,
    blocks: any,
    htmlContent: string
  ) => Promise<repository.Page>;
  deletePage: (ID: string) => Promise<void>;
};

type TPage = WithoutRepositoryBaseFields<repository.Page>;

export const DEFAULT_PAGE_TITLE = "New Page";

export const DEFAULT_FOLDER_TITLE = "New Folder";

export const useLocalPagesStore = create<LocalPagesStore>((set, get) => ({
  pages: [],

  getPages: async () => {
    const pages = await GetLocalPages();
    set({ pages });
  },

  getPage: async (ID) => {
    const page = await GetLocalPage(ID);
    return page;
  },

  newPage: async () => {
    const body: TPage = {
      title: DEFAULT_PAGE_TITLE,
      blocks: [],
      html_content: "",
      is_folder: false,
      expanded: false,
      order: get().pages.length + 1,
      Children: [],
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    get().getPages();

    return newPage;
  },

  newFolder: async (name: string) => {
    const body: TPage = {
      title: name || DEFAULT_FOLDER_TITLE,
      blocks: [],
      html_content: "",
      is_folder: true,
      expanded: false,
      order: get().pages.length + 1,
      Children: [],
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    get().getPages();

    return newPage;
  },

  savePage: async (page) => {
    return SaveLocalPage(repository.Page.createFrom(page));
  },

  saveNewPageOrder: async (page) => {
    return UpdateLocalPageOrder(page.ID, page.ParentID || null, page.order);
  },

  togglePageExpanded: async (page) => {
    const body: TPage = {
      ...page,
      expanded: !page.expanded,
    };

    await SaveLocalPage(repository.Page.createFrom(body));

    get().getPages();
  },

  savePageTitle: async (title, page) => {
    const body: TPage = {
      ...page,
      title,
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    get().getPages();

    return newPage;
  },

  savePageBlocks: async (page, blocks, htmlContent) => {
    const body: TPage = {
      ...page,
      blocks,
      html_content: htmlContent,
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    get().getPages();

    return newPage;
  },

  deletePage: async (ID) => {
    await DeleteLocalPage(ID);

    get().getPages();
  },
}));
