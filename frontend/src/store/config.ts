import { WithoutRepositoryBaseFields } from "@/types";
import { create } from "zustand";
import { GetConfig, SaveConfig } from "~wails/main/App";
import { repository } from "~wails/models";

type ConfigStore = {
  config: repository.Config | null;
  fetchConfig: () => Promise<void>;
  writeConfig: (config: TConfig) => Promise<void>;
};

type TConfig = WithoutRepositoryBaseFields<repository.Config>;

export const useConfigStore = create<ConfigStore>((set, get) => ({
  config: null,

  fetchConfig: async () => {
    const config = await GetConfig();
    set({ config });
  },

  writeConfig: async (config) => {
    const newConfig = await SaveConfig(
      repository.Config.createFrom({
        ...get().config,
        ...config,
      })
    );

    set({ config: newConfig });
  },
}));
