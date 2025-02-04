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
import {
  GetAppVersion,
  IsGoogleAuthEnabled,
  StartGoogleAuthorization,
  StartSynchronizer,
} from "~wails/main/App";
import { EventsOn, EventsOnce } from "~wails-runtime";
import { useDebouncedCallback } from "use-debounce";
import {
  isGoogleAPIInvalidGrantError,
  useGoogleAuthTokenStore,
} from "@/store/google-auth-token";
import { Loader2 } from "lucide-react";

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
    GetAppVersion().then((v) => setAppVersion(v));
  }, []);

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
  // Google auth token
  const googleAuthToken = useGoogleAuthTokenStore((state) => state.token);
  const getGoogleAuthToken = useGoogleAuthTokenStore((state) => state.getToken);
  const deleteGoogleAuthToken = useGoogleAuthTokenStore(
    (state) => state.deleteToken
  );

  const [authorizing, setAuthorizing] = useState(false);

  // Services status
  const [googleAuthEnabled, setGoogleAuthEnabled] = useState(false);

  const cloudEnabled = googleAuthEnabled;

  // Start synchronization here
  const startSynchronizer = useDebouncedCallback(() => {
    StartSynchronizer().catch((error) => {
      // If the error is due to an expired token,
      // delete it and show a toast
      if (isGoogleAPIInvalidGrantError(error)) {
        deleteGoogleAuthToken();
        toast.warning("Google auth token expired, please re-authorize");
      } else {
        console.log("Error starting synchronizer");
      }
    });
  }, 1000);

  useEffect(() => {
    if (cloudEnabled && googleAuthToken) {
      startSynchronizer();
    }
  }, [cloudEnabled, googleAuthToken]);

  useEffect(() => {
    getGoogleAuthToken();
  }, []);

  useEffect(() => {
    IsGoogleAuthEnabled().then(setGoogleAuthEnabled);
  }, []);

  useEffect(() => {
    return EventsOn("on-google-authorization-timeout", () => {
      toast.warning("Google authorization timeout");
      setAuthorizing(false);
    });
  }, []);

  useEffect(() => {
    return EventsOn("on-google-authorization-error", (error) => {
      toast.error("Error Google authorization: " + error);
    });
  }, []);

  const onGoogleAuthorizationSuccess = useCallback(async () => {
    const token = await getGoogleAuthToken();

    if (!token) {
      return;
    }

    const toastID = toast.message(
      <div className="flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span>Restoring your data from drive</span>
      </div>,
      {
        description: "This might take a few moments...",
        duration: Infinity,
      }
    );

    let onSuccessEvent: VoidFunction | undefined;
    let onFailureEvent: VoidFunction | undefined;

    // Dismiss toast when sync is completed
    onSuccessEvent = EventsOnce("on-sync-success", () => {
      toast.dismiss(toastID);

      onSuccessEvent?.();
      onFailureEvent?.();
    });

    // Dismiss toast when sync fails
    onFailureEvent = EventsOnce("on-sync-failure", (error) => {
      toast.dismiss(toastID);

      onSuccessEvent?.();
      onFailureEvent?.();
    });
  }, [getGoogleAuthToken]);

  const startGoogleAuthorization = useCallback(() => {
    setAuthorizing(true);

    return StartGoogleAuthorization()
      .then(onGoogleAuthorizationSuccess)
      .finally(() => setAuthorizing(false));
  }, [onGoogleAuthorizationSuccess]);

  return {
    googleAuthEnabled,
    cloudEnabled,
    googleAuthToken,
    startGoogleAuthorization,
    deleteGoogleAuthToken,
    authorizing,
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
