import { useConfigStore } from "@/store/config";
import { WithoutRepositoryBaseFields } from "@/types";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
} from "react";
import { toast } from "sonner";
import { repository, whisper } from "~wails/models";

import isEqual from "lodash/isEqual";
import { useLocalWhisperStore } from "@/store/local-whisper";

export type TranscriberSource = "local" | "openai" | "witai";

export type TranscriberSources = {
  [key in TranscriberSource]: {
    name: string;
    key: key;
  };
};

export const TRANSCRIBER_SOURCES: TranscriberSources = {
  local: {
    name: "Local",
    key: "local",
  },
  openai: {
    name: "OpenAI",
    key: "openai",
  },
  witai: {
    name: "Wit.ai",
    key: "witai",
  },
};

type SettingsState = WithoutRepositoryBaseFields<repository.Config> & {
  TranscriberSource: TranscriberSource;
};

export type SettingsContextValue = {
  state: SettingsState;
  bestWhisperModel: string | undefined;
  whisperModels: whisper.WhisperModel[];
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
  const [state, dispatch] = useReducer(reducer, { TranscriberSource: "local" });

  const localWhisperStore = useLocalWhisperStore();
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
    if (state.TranscriberSource === "local") {
      localWhisperStore.getBestModel();
    }
  }, [state.TranscriberSource]);

  useEffect(() => {
    localWhisperStore.getModels();
  }, []);

  const handleSave = useCallback(() => {
    if (state.TranscriberSource === "openai") {
      if (!state.OpenAIApiKey) {
        toast.error("OpenAI API key is required");
        return;
      }

      if (!state.OpenAIApiKey.startsWith("sk-")) {
        toast.error("OpenAI API key must start with 'sk-'");
        return;
      }
    }

    configStore.writeConfig(state).then(() => {
      toast.success("Settings saved successfully!");
    });
  }, [state]);

  const configModified = useMemo(() => {
    return !isEqual(state, configStore.config);
  }, [state, configStore.config]);

  return (
    <SettingsContext.Provider
      value={{
        state,
        dispatch,
        handleSave,
        configModified,
        bestWhisperModel: localWhisperStore.bestModel,
        whisperModels: localWhisperStore.models,
      }}
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
