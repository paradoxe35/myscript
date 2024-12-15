import { atom } from "jotai";
import { GetConfig, SaveConfig } from "~wails/main/App";
import { repository } from "~wails/models";

export const configAtom = atom(
  async () => GetConfig(),
  async (_get, _set, payload: repository.Config) => {
    return SaveConfig(payload);
  }
);
