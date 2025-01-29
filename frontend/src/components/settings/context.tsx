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
import { repository, whisper } from "~wails/models";

import isEqual from "lodash/isEqual";
import { useLocalWhisperStore } from "@/store/local-whisper";
import { GetAppVersion, IsSynchronizerEnabled } from "~wails/main/App";

export type TranscriberSource = "local" | "openai" | "witai" | "groq";

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
  groq: {
    name: "Groq",
    key: "groq",
  },
};

type SettingsState = WithoutRepositoryBaseFields<repository.Config> & {
  TranscriberSource: TranscriberSource;
};

type SettingsContextValue = ReturnType<typeof useSettingsHook>;

const SettingsContext = createContext<SettingsContextValue>({} as any);

function reducer(state: SettingsState, action: Partial<SettingsState>) {
  return {
    ...state,
    ...action,
  };
}

function useSettingsHook() {
  const [appVersion, setAppVersion] = useState("");

  const [state, dispatch] = useReducer(reducer, { TranscriberSource: "local" });

  const cloudHook = useCloudSettings();

  const localWhisperStore = useLocalWhisperStore();
  const configStore = useConfigStore();

  useEffect(() => {
    configStore.fetchConfig();
  }, []);

  useEffect(() => {
    GetAppVersion().then((v) => setAppVersion(v));
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

  function validateOpenAIApiKey(state: SettingsState) {
    if (state.TranscriberSource === "openai") {
      if (!state.OpenAIApiKey) {
        toast.error("OpenAI API key is required");
        return false;
      }

      if (!state.OpenAIApiKey.startsWith("sk-")) {
        toast.error("OpenAI API key must start with 'sk-'");
        return false;
      }
    }

    return true;
  }

  function validateGroqApiKey(state: SettingsState) {
    if (state.TranscriberSource === "groq") {
      if (!state.GroqApiKey) {
        toast.error("Groq API key is required");
        return false;
      }

      if (!state.GroqApiKey.startsWith("gsk_")) {
        toast.error("Groq API key must start with 'gsk_'");
        return false;
      }
    }
    return true;
  }

  const handleSave = useCallback(() => {
    const valid = [
      validateOpenAIApiKey(state),
      validateGroqApiKey(state),
    ].every((v) => v);

    if (!valid) return;

    configStore.writeConfig(state).then(() => {
      toast.success("Settings saved successfully!");
    });
  }, [state]);

  const configModified = useMemo(() => {
    return !isEqual(state, configStore.config);
  }, [state, configStore.config]);

  return {
    state,
    dispatch,
    handleSave,
    appVersion,
    configModified,
    bestWhisperModel: localWhisperStore.bestModel,
    whisperModels: localWhisperStore.models,
    cloud: cloudHook,
  };
}

function useCloudSettings() {
  const [cloudEnabled, setCloudEnabled] = useState(false);

  useEffect(() => {
    IsSynchronizerEnabled().then((enabled) => {
      setCloudEnabled(enabled);
    });
  }, []);

  return {
    cloudEnabled,
  };
}

export function SettingsProvider({ children }: React.PropsWithChildren) {
  const hook = useSettingsHook();

  return (
    <SettingsContext.Provider value={hook}>{children}</SettingsContext.Provider>
  );
}

export function useSettings() {
  const context = useContext(SettingsContext);

  if (!context) {
    throw new Error("useSettings must be used within a SettingsProvider");
  }

  return context;
}
