import { useConfigStore } from "@/store/config";
import { WithoutRepositoryBaseFields } from "@/types";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useReducer,
} from "react";
import { toast } from "sonner";
import { repository } from "~wails/models";

export type WhisperSource = "local" | "openai";

type SettingsState = WithoutRepositoryBaseFields<repository.Config> & {
  WhisperSource: WhisperSource;
};

export type SettingsContextValue = {
  state: SettingsState;
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

  const configStore = useConfigStore();

  useEffect(() => {
    configStore.fetchConfig();
  }, []);

  useEffect(() => {
    const config = configStore.config;

    if (config) {
      dispatch({
        ...(config as any),
        WhisperSource: config.WhisperSource || "local",
      });
    }
  }, [configStore.config]);

  const handleSave = useCallback(() => {
    configStore.writeConfig(state).then(() => {
      toast.success("Settings saved successfully!");
    });
  }, [state]);

  return (
    <SettingsContext.Provider value={{ state, dispatch, handleSave }}>
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
