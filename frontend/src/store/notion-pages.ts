import { create } from "zustand";
import { useConfigStore } from "./config";

type NotionPagesStore = {};

export const useNotionPagesStore = create<NotionPagesStore>((set, get) => ({}));
