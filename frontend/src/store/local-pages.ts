import { WithoutRepositoryBaseFields } from "@/types";
import { create } from "zustand";
import {
  DeleteLocalPage,
  GetLocalPage,
  GetLocalPages,
  SaveLocalPage,
  SetPageExpanded,
  UpdateLocalPageOrder,
  UpdateLocalPageTitle,
} from "~wails/main/App";
import { repository } from "~wails/models";

type LocalPagesStore = {
  pages: Array<repository.Page>;
  getPages: () => Promise<void>;
  getPage: (ID: string) => Promise<repository.Page>;
  newPage: () => Promise<repository.Page>;
  newFolder: (name: string) => Promise<repository.Page>;
  togglePageExpanded: (page: repository.Page) => Promise<void>;
  expandPages: (IDs: string[]) => Promise<void>;
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

// Whether a folder is open belongs to this machine and has its own binding.
type TPage = Omit<WithoutRepositoryBaseFields<repository.Page>, "expanded">;

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
      order: get().pages.length + 1,
      Children: [],
    };

    const newPage = await SaveLocalPage(repository.Page.createFrom(body));

    get().getPages();

    return newPage;
  },

  saveNewPageOrder: async (page) => {
    return UpdateLocalPageOrder(page.ID, page.ParentID || null, page.order);
  },

  togglePageExpanded: async (page) => {
    await SetPageExpanded(page.ID, !page.expanded);

    get().getPages();
  },

  expandPages: async (IDs) => {
    await Promise.all(IDs.map((ID) => SetPageExpanded(ID, true)));

    get().getPages();
  },

  savePageTitle: async (title, page) => {
    const newPage = await UpdateLocalPageTitle(page.ID, title);

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
