import { useConfigStore } from "@/store/config";
import { WithoutRepositoryBaseFields } from "@/types";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useState,
} from "react";
import { toast } from "sonner";
import { GetBestWhisperModel } from "~wails/main/App";
import { repository } from "~wails/models";

import isEqual from "lodash/isEqual";

export type WhisperSource = "local" | "openai";

type SettingsState = WithoutRepositoryBaseFields<repository.Config> & {
  WhisperSource: WhisperSource;
};

export type SettingsContextValue = {
  state: SettingsState;
  bestWhisperModel: string | undefined;
  configModified: boolean;
  dispatch: React.Dispatch<Partial<SettingsState>>;
  handleSave: () => void;
};

const SettingsContext = createContext<SettingsContextValue>({} as any);

function reducer(state: SettingsState, action: Partial<SettingsState>) {
  return {
    ...state,
    ...action,
  };
}

export function SettingsProvider({ children }: React.PropsWithChildren) {
  const [state, dispatch] = useReducer(reducer, { WhisperSource: "local" });
  const [bestWhisperModel, setBestWhisperModel] = useState<string>();

  const configStore = useConfigStore();

  useEffect(() => {
    configStore.fetchConfig();
  }, []);

  useEffect(() => {
    const config = configStore.config as SettingsState;

    if (config) {
      dispatch(config);
    }
  }, [configStore.config]);

  useEffect(() => {
    if (state.WhisperSource === "local") {
      GetBestWhisperModel().then((bestWhisperModel) => {
        setBestWhisperModel(bestWhisperModel);
      });
    }
  }, [state.WhisperSource]);

  const handleSave = useCallback(() => {
    configStore.writeConfig(state).then(() => {
      toast.success("Settings saved successfully!");
    });
  }, [state]);

  const configModified = useMemo(() => {
    return !isEqual(state, configStore.config);
  }, [state, configStore.config]);

  return (
    <SettingsContext.Provider
      value={{ state, dispatch, handleSave, bestWhisperModel, configModified }}
    >
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const context = useContext(SettingsContext);

  if (!context) {
    throw new Error("useSettings must be used within a SettingsProvider");
  }

  return context;
}
