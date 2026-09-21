import { create } from "zustand";
import { GetDeviceSettings, SaveDeviceSettings } from "~wails/main/App";
import { repository } from "~wails/models";

export type TranscriberSource = "local" | "remote" | "witai";

type DeviceSettingsStore = {
  settings: repository.DeviceSettings | null;
  fetch: () => Promise<void>;
  write: (patch: Partial<repository.DeviceSettings>) => Promise<void>;
};

export const useDeviceSettingsStore = create<DeviceSettingsStore>(
  (set, get) => ({
    settings: null,

    fetch: async () => {
      set({ settings: await GetDeviceSettings() });
    },

    // Saving replaces the whole row, so a write before the first fetch must
    // not wipe the other fields.
    write: async (patch) => {
      const current = get().settings ?? (await GetDeviceSettings());
      const settings = await SaveDeviceSettings(
        repository.DeviceSettings.createFrom({ ...current, ...patch })
      );
      set({ settings });
    },
  })
);

export function transcriberSource(
  settings: repository.DeviceSettings | null
): TranscriberSource {
  return (settings?.TranscriberSource || "local") as TranscriberSource;
}
