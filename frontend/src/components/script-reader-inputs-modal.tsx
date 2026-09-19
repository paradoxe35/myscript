import { Check } from "lucide-react";

import { cn } from "@/lib/utils";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  AUTO_DETECT_LANGUAGE,
  useTranscriberStore,
} from "@/store/transcriber";
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useConfigStore } from "@/store/config";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { languages, main, stt } from "~wails/models";
import { useActivePageStore } from "@/store/active-page";
import { Checkbox } from "./ui/checkbox";
import { Label } from "./ui/label";
import { useContentReadStore } from "@/store/content-read";
import { useSpeechModelsStore } from "@/store/speech-models";

type Props = {
  trigger: React.ReactNode | null;
  onStartReading: (languageCode: string, micInputDevice: string) => void;
};

const SRInputsContext = createContext<ReturnType<typeof useSRInputs>>(
  {} as ReturnType<typeof useSRInputs>
);

function useSRInputs(props: Props) {
  const transcriberStore = useTranscriberStore();
  const config = useConfigStore((state) => state.config);
  const localSource = config?.TranscriberSource === "local";

  const [dialogOpen, setDialogOpen] = useState(false);

  const models = useModels(dialogOpen, localSource);
  const { languages, selectedLanguageCode, setSelectedLanguageCode } =
    useLanguages(localSource, models.selectedModel);

  const [micInputDevice, setMicInputDevice] = useState<stt.Device | null>(
    null
  );
  const micInputDevices = transcriberStore.micInputDevices;

  const onStartReading = () => {
    requestAnimationFrame(() => {
      if (selectedLanguageCode && micInputDevice) {
        props.onStartReading(selectedLanguageCode, micInputDevice.Name);
      }
    });
  };

  useEffect(() => {
    if (micInputDevice) {
      transcriberStore.setDefaultMicInput(micInputDevice);
    }
  }, [micInputDevice]);

  useEffect(() => {
    if (!dialogOpen) return;

    transcriberStore.getMicInputDevices().then(async (micInputDevices) => {
      let defaultDevice = await transcriberStore.getDefaultMicInput();

      if (!defaultDevice) {
        defaultDevice =
          micInputDevices.find((device) => device.IsDefault) ||
          micInputDevices[0];
      }
      if (defaultDevice) {
        setMicInputDevice(defaultDevice);
      }
    });
  }, [dialogOpen]);

  const canSubmit =
    !!selectedLanguageCode &&
    !!micInputDevice &&
    (!localSource || !!models.selectedModel);

  return {
    localSource,
    languages,
    selectedLanguageCode,
    setSelectedLanguageCode,
    onStartReading,
    canSubmit,

    ...models,

    micInputDevice,
    micInputDevices,
    setMicInputDevice,

    dialogOpen,
    setDialogOpen,
  };
}

// The model chosen here becomes the active one, so the settings panel and the
// next take agree.
function useModels(dialogOpen: boolean, localSource: boolean) {
  const speechModelsStore = useSpeechModelsStore();
  const configStore = useConfigStore();

  const downloadedModels = speechModelsStore.downloaded();
  const configuredID = configStore.config?.SpeechModelID;

  const selectedModel = useMemo(() => {
    return (
      downloadedModels.find((model) => model.ID === configuredID) ||
      downloadedModels[0]
    );
  }, [downloadedModels, configuredID]);

  useEffect(() => {
    if (dialogOpen && localSource) {
      speechModelsStore.fetchModels();
    }
  }, [dialogOpen, localSource]);

  useEffect(() => {
    if (localSource && selectedModel && selectedModel.ID !== configuredID) {
      configStore.writeConfig({ SpeechModelID: selectedModel.ID });
    }
  }, [localSource, selectedModel?.ID, configuredID]);

  const setSelectedModel = (model: main.SpeechModel) => {
    configStore.writeConfig({ SpeechModelID: model.ID });
  };

  return { downloadedModels, selectedModel, setSelectedModel };
}

function useLanguages(
  localSource: boolean,
  selectedModel: main.SpeechModel | undefined
) {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();
  const config = useConfigStore((state) => state.config);

  const [selectedLanguageCode, setSelectedLanguageCode] = useState<
    string | null
  >();
  const [pageLanguage, setPageLanguage] = useState<string | null>(null);

  const activePageID = activePageStore.getPageId();

  // A model that declares no languages (a dropped-in file) is left to detect.
  const available = useMemo<languages.Language[]>(() => {
    if (!localSource) {
      // A hosted service detects when no language is sent. A Wit app is built
      // for exactly one language, so it never does.
      return config?.TranscriberSource === "remote"
        ? [AUTO_DETECT_LANGUAGE, ...transcriberStore.languages]
        : transcriberStore.languages;
    }
    if (!selectedModel) return [];
    const spoken = selectedModel.Languages || [];
    return selectedModel.LanguageDetect || spoken.length === 0
      ? [AUTO_DETECT_LANGUAGE, ...spoken]
      : spoken;
  }, [
    localSource,
    config?.TranscriberSource,
    transcriberStore.languages,
    selectedModel,
  ]);

  // The remembered language sorts first, and stays there while the user browses.
  const languages = useMemo(() => {
    return available.slice().sort((a, b) => {
      if (a.Code === pageLanguage) return -1;
      if (b.Code === pageLanguage) return 1;
      return 0;
    });
  }, [available, pageLanguage]);

  useEffect(() => {
    if (!localSource) {
      transcriberStore.getLanguages();
    }
  }, [localSource, config?.TranscriberSource]);

  useEffect(() => {
    if (activePageID) {
      transcriberStore.getPageLanguage(activePageID).then((language) => {
        setPageLanguage(language || "en");
      });
    }
  }, [activePageID]);

  useEffect(() => {
    if (available.length === 0 || pageLanguage === null) return;

    const wanted = pageLanguage;
    const fallback = available.some((l) => l.Code === "en")
      ? "en"
      : available[0].Code;
    const code = available.some((l) => l.Code === wanted) ? wanted : fallback;
    setSelectedLanguageCode(code);
  }, [available, pageLanguage]);

  useEffect(() => {
    const cacheSelectedLanguageCode = async () => {
      if (activePageID && selectedLanguageCode) {
        const cachedLanguage = await transcriberStore.getPageLanguage(
          activePageID
        );
        if (cachedLanguage !== selectedLanguageCode) {
          transcriberStore.setPageLanguage(activePageID, selectedLanguageCode);
        }
      }
    };
    cacheSelectedLanguageCode();
  }, [activePageID, selectedLanguageCode]);

  return {
    languages,
    setSelectedLanguageCode,
    selectedLanguageCode,
  };
}

function modelSummary(model: main.SpeechModel) {
  const languages = model.Languages || [];
  const spoken =
    languages.length === 1 ? languages[0].Name : `${languages.length} languages`;
  return model.LanguageDetect ? `${spoken}, auto-detect` : spoken;
}

function ModelCommands() {
  const { downloadedModels, selectedModel, setSelectedModel } =
    useContext(SRInputsContext);

  return (
    <Command>
      <CommandInput placeholder="Search model..." className="h-9" />
      <CommandList className="min-h-[300px]">
        <CommandEmpty>
          No downloaded model. Download one in Settings, Speech Recognition.
        </CommandEmpty>
        <CommandGroup>
          {downloadedModels.map((model) => (
            <CommandItem
              key={model.ID}
              value={model.Name}
              onSelect={() => setSelectedModel(model)}
            >
              <div className="flex flex-col min-w-0">
                <span className="truncate">
                  {model.Name}
                  {model.Suggested && (
                    <span className="ml-2 text-[10px] uppercase tracking-wide text-primary">
                      Suggested
                    </span>
                  )}
                </span>
                <span className="text-xs opacity-60 truncate">
                  {modelSummary(model)} · {model.FitLabel}
                </span>
              </div>
              <Check
                className={cn(
                  "ml-auto",
                  selectedModel?.ID === model.ID ? "opacity-100" : "opacity-0"
                )}
              />
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
}

function LanguageCommands() {
  const { languages, setSelectedLanguageCode, selectedLanguageCode } =
    useContext(SRInputsContext);

  return (
    <Command>
      <CommandInput placeholder="Search language..." className="h-9" />
      <CommandList className="min-h-[300px]">
        <CommandEmpty>No language found.</CommandEmpty>
        <CommandGroup>
          {languages.map((language) => (
            <CommandItem
              key={language.Code}
              value={language.Name}
              className={`command-language-${language.Code}`}
              onSelect={() => {
                setSelectedLanguageCode(language.Code);
              }}
            >
              {language.Name}
              <Check
                className={cn(
                  "ml-auto",
                  selectedLanguageCode === language.Code
                    ? "opacity-100"
                    : "opacity-0"
                )}
              />
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
}

function MicInputDevices() {
  const { micInputDevices, setMicInputDevice, micInputDevice } =
    useContext(SRInputsContext);

  return (
    <Command>
      <CommandInput placeholder="Search device..." className="h-9" />
      <CommandList className="min-h-[300px]">
        <CommandEmpty>No microphone found.</CommandEmpty>
        <CommandGroup>
          {micInputDevices.map((device) => (
            <CommandItem
              key={device.Name}
              value={device.Name}
              onSelect={() => setMicInputDevice(device)}
            >
              {device.Name}
              {device.IsDefault && (
                <span className="ml-2 text-xs opacity-60">default</span>
              )}
              <Check
                className={cn(
                  "ml-auto",
                  micInputDevice?.Name === device.Name
                    ? "opacity-100"
                    : "opacity-0"
                )}
              />
            </CommandItem>
          ))}
        </CommandGroup>
      </CommandList>
    </Command>
  );
}

function ResumeRead() {
  const activePageStore = useActivePageStore();
  const contentReadStore = useContentReadStore();

  useEffect(() => {
    contentReadStore.resetResume();
  }, [activePageStore.getPageId()]);

  useEffect(() => {
    if (!activePageStore.readMode) {
      contentReadStore.resetResume();
    }
  }, [activePageStore.readMode]);

  return (
    <div className="flex items-center space-x-2">
      <Checkbox
        checked={contentReadStore.resume}
        onCheckedChange={(checked) => {
          contentReadStore.setResume(Boolean(checked));
        }}
        id="resume-read-position"
      />

      <Label htmlFor="resume-read-position" className="leading-none">
        Resume reading position
      </Label>
    </div>
  );
}

export default function SRInputsModal(props: Props) {
  const ctx = useSRInputs(props);

  return (
    <Dialog open={ctx.dialogOpen} onOpenChange={ctx.setDialogOpen}>
      <DialogTrigger asChild>{props.trigger}</DialogTrigger>

      <DialogContent className="sm:max-w-md" showCloseButton={false}>
        <SRInputsContext.Provider value={ctx}>
          <Tabs
            defaultValue={ctx.localSource ? "model" : "languages"}
            className="w-full"
          >
            <DialogHeader>
              <DialogTitle>
                <TabsList
                  className={cn(
                    "grid w-full",
                    ctx.localSource ? "grid-cols-3" : "grid-cols-2"
                  )}
                >
                  {ctx.localSource && (
                    <TabsTrigger value="model">Model</TabsTrigger>
                  )}
                  <TabsTrigger value="languages">Languages</TabsTrigger>
                  <TabsTrigger value="microphone">Microphone</TabsTrigger>
                </TabsList>
              </DialogTitle>

              <DialogDescription />
            </DialogHeader>

            {ctx.localSource && (
              <TabsContent value="model">
                <ModelCommands />
              </TabsContent>
            )}

            <TabsContent value="languages">
              <LanguageCommands />
            </TabsContent>

            <TabsContent value="microphone">
              <MicInputDevices />
            </TabsContent>
          </Tabs>
        </SRInputsContext.Provider>

        <ResumeRead />

        <DialogFooter className="justify-between sm:justify-between">
          <DialogClose asChild>
            <Button type="button" variant="outline">
              Cancel
            </Button>
          </DialogClose>

          <DialogClose asChild>
            <Button
              type="button"
              disabled={!ctx.canSubmit}
              onClick={ctx.onStartReading}
            >
              Start Reading
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
