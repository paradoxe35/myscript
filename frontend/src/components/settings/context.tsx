import { useConfigStore } from "@/store/config";
import {
  TranscriberSource,
  useDeviceSettingsStore,
} from "@/store/device-settings";
import { useGoogleAuthTokenStore } from "@/store/google-auth-token";
import { isGoogleAPIInvalidGrantError } from "@/store/google-auth-token";
import { WithoutRepositoryBaseFields } from "@/types";
import { Loader2 } from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { toast } from "sonner";
import { useDebouncedCallback } from "use-debounce";
import { EventsOn, EventsOnce } from "~wails-runtime";
import {
  GetAppVersion,
  IsGoogleAuthEnabled,
  StartGoogleAuthorization,
  StartSynchronizer,
} from "~wails/main/App";
import { repository } from "~wails/models";

export const TRANSCRIBER_SOURCES: Array<{
  key: TranscriberSource;
  name: string;
  description: string;
}> = [
  {
    key: "local",
    name: "On this computer",
    description: "Private and offline, once a model is downloaded.",
  },
  {
    key: "remote",
    name: "Hosted service",
    description: "OpenAI, Groq, Gemini or your own endpoint. Needs an API key.",
  },
  {
    key: "witai",
    name: "Wit.ai",
    description: "Free and needs no key, but less accurate.",
  },
];

type ConfigPatch = Partial<WithoutRepositoryBaseFields<repository.Config>>;

type SettingsContextValue = ReturnType<typeof useSettingsState>;

function reportSaveError(error: unknown) {
  toast.error(`Could not save: ${error}`);
}

const SettingsContext = createContext<SettingsContextValue | null>(null);

function useSettingsState() {
  const [appVersion, setAppVersion] = useState("");

  const config = useConfigStore((state) => state.config);
  const fetchConfig = useConfigStore((state) => state.fetchConfig);
  const writeConfig = useConfigStore((state) => state.writeConfig);

  const deviceSettings = useDeviceSettingsStore((state) => state.settings);
  const fetchDeviceSettings = useDeviceSettingsStore((state) => state.fetch);
  const writeDeviceSettings = useDeviceSettingsStore((state) => state.write);

  const cloud = useCloudSettings();

  useEffect(() => {
    GetAppVersion().then(setAppVersion);
    fetchConfig();
    fetchDeviceSettings();
  }, []);

  const updateConfig = useCallback(
    (patch: ConfigPatch) => writeConfig(patch).catch(reportSaveError),
    [writeConfig],
  );

  const updateDeviceSettings = useCallback(
    (patch: Partial<repository.DeviceSettings>) =>
      writeDeviceSettings(patch).catch(reportSaveError),
    [writeDeviceSettings],
  );

  return {
    appVersion,
    config,
    updateConfig,
    deviceSettings,
    updateDeviceSettings,
    cloud,
  };
}

function useCloudSettings() {
  const googleAuthToken = useGoogleAuthTokenStore((state) => state.token);
  const getGoogleAuthToken = useGoogleAuthTokenStore((state) => state.getToken);
  const deleteGoogleAuthToken = useGoogleAuthTokenStore(
    (state) => state.deleteToken,
  );

  const [authorizing, setAuthorizing] = useState(false);
  const [googleAuthEnabled, setGoogleAuthEnabled] = useState(false);

  const cloudEnabled = googleAuthEnabled;

  const startSynchronizer = useDebouncedCallback(() => {
    StartSynchronizer().catch((error) => {
      if (isGoogleAPIInvalidGrantError(error)) {
        deleteGoogleAuthToken();
        toast.warning("Google auth token expired, please re-authorize");
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
    IsGoogleAuthEnabled().then(setGoogleAuthEnabled);
  }, []);

  useEffect(() => {
    return EventsOn("on-google-authorization-timeout", () => {
      toast.warning("Google authorization timed out");
      setAuthorizing(false);
    });
  }, []);

  useEffect(() => {
    return EventsOn("on-google-authorization-error", (error) => {
      toast.error("Google authorization failed: " + error);
    });
  }, []);

  // The backend clears a grant Google keeps refusing, so the UI stops showing
  // a connected account that cannot sync.
  useEffect(() => {
    return EventsOn("on-google-authorization-lost", () => {
      getGoogleAuthToken();
      toast.warning("Google Drive disconnected", {
        description:
          "The authorization expired or was revoked. Connect again to resume syncing.",
      });
    });
  }, [getGoogleAuthToken]);

  const onAuthorized = useCallback(async () => {
    const token = await getGoogleAuthToken();
    if (!token) return;

    const toastID = toast.message(
      <div className="flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin" />
        <span>Restoring your data from Drive</span>
      </div>,
      { description: "This might take a few moments...", duration: Infinity },
    );

    let clearSuccess: VoidFunction | undefined;
    let clearFailure: VoidFunction | undefined;

    const dismiss = () => {
      toast.dismiss(toastID);
      clearSuccess?.();
      clearFailure?.();
    };

    clearSuccess = EventsOnce("on-sync-success", dismiss);
    clearFailure = EventsOnce("on-sync-failure", dismiss);
  }, [getGoogleAuthToken]);

  const startGoogleAuthorization = useCallback(() => {
    setAuthorizing(true);

    return StartGoogleAuthorization()
      .then(onAuthorized)
      .finally(() => setAuthorizing(false));
  }, [onAuthorized]);

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
  return (
    <SettingsContext.Provider value={useSettingsState()}>
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
