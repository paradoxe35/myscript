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
import { useTranscriberStore } from "@/store/transcriber";
import { createContext, useContext, useEffect, useState } from "react";
import { useConfigStore } from "@/store/config";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import { microphone } from "~wails/models";
import { useActivePageStore } from "@/store/active-page";
import { Checkbox } from "./ui/checkbox";
import { useContentReadStore } from "@/store/content-read";

type Props = {
  trigger: React.ReactNode | null;
  onLanguageSelected: (
    languageCode: string,
    micInputDeviceID: number[]
  ) => void;
};

const SRInputsContext = createContext<ReturnType<typeof useSRInputs>>(
  {} as ReturnType<typeof useSRInputs>
);

function useSRInputs(props: Props) {
  const transcriberStore = useTranscriberStore();
  const { languages, selectedLanguageCode, setSelectedLanguageCode } =
    useLanguages();

  const [micInputDevice, setMicInputDevice] =
    useState<microphone.MicInputDevice | null>(null);

  const micInputDevices = transcriberStore.micInputDevices;

  const onStartReading = () => {
    requestAnimationFrame(() => {
      if (selectedLanguageCode && micInputDevice) {
        props.onLanguageSelected(selectedLanguageCode, micInputDevice.ID);
      }
    });
  };

  useEffect(() => {
    if (micInputDevice) {
      transcriberStore.setDefaultMicInput(micInputDevice);
    }
  }, [micInputDevice]);

  useEffect(() => {
    transcriberStore.getMicInputDevices().then(async (micInputDevices) => {
      let defaultDevice = await transcriberStore.getDefaultMicInput();

      if (defaultDevice) {
        setMicInputDevice(defaultDevice);
        return;
      }

      defaultDevice = micInputDevices.find((device) => device.IsDefault === 1);
      if (defaultDevice) {
        setMicInputDevice(defaultDevice);
      }
    });
  }, []);

  const canSubmit = selectedLanguageCode && micInputDevice;

  return {
    languages,
    selectedLanguageCode,
    setSelectedLanguageCode,
    onStartReading,
    canSubmit,

    micInputDevice,
    micInputDevices,
    setMicInputDevice,
  };
}

function useLanguages() {
  const transcriberStore = useTranscriberStore();
  const activePageStore = useActivePageStore();
  const config = useConfigStore((state) => state.config);

  const [selectedLanguageCode, setSelectedLanguageCode] = useState<
    string | null
  >();

  const [languages, setLanguages] = useState(transcriberStore.languages);

  const activePageID = activePageStore.getPageId();

  useEffect(() => {
    setLanguages(transcriberStore.languages);
  }, [transcriberStore.languages]);

  // Set page language
  useEffect(() => {
    if (activePageID && selectedLanguageCode) {
      transcriberStore.setPageLanguage(activePageID, selectedLanguageCode);
    }
  }, [activePageID, selectedLanguageCode]);

  useEffect(() => {
    transcriberStore.getLanguages();
  }, [config?.TranscriberSource]);

  // Get page language
  useEffect(() => {
    if (activePageID) {
      transcriberStore.getPageLanguage(activePageID).then((language) => {
        const lan = language || "en";
        setSelectedLanguageCode(lan);

        // Sort languages, so the selected language is at the top
        setTimeout(() => {
          setLanguages((prev) => {
            return prev.slice().sort((a, b) => {
              if (a.Code === lan) return -1;
              if (b.Code === lan) return 1;
              return 0;
            });
          });
        }, 1000);
      });
    }
  }, [activePageID]);

  return {
    languages,
    setSelectedLanguageCode,
    selectedLanguageCode,
  };
}

function LanguageCommands() {
  const { languages, setSelectedLanguageCode, selectedLanguageCode } =
    useContext(SRInputsContext);

  return (
    <Command>
      <CommandInput placeholder="Search language..." className="h-9" />
      <CommandList>
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
        <CommandEmpty>No device found.</CommandEmpty>
        <CommandGroup>
          {micInputDevices.map((device) => (
            <CommandItem
              key={device.Name}
              value={device.Name}
              onSelect={() => setMicInputDevice(device)}
            >
              {device.Name}
              <Check
                className={cn(
                  "ml-auto",
                  micInputDevice === device ? "opacity-100" : "opacity-0"
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
    contentReadStore.setResume(false);
  }, [activePageStore.getPageId()]);

  useEffect(() => {
    // Reset resume when read mode is off
    if (!activePageStore.readMode) {
      contentReadStore.setResume(false);
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

      <label
        htmlFor="resume-read-position"
        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Resume reading position
      </label>
    </div>
  );
}

export default function SRInputsModal(props: Props) {
  const ctx = useSRInputs(props);

  return (
    <Dialog>
      <DialogTrigger asChild>{props.trigger}</DialogTrigger>

      <DialogContent className="sm:max-w-md" showCloseButton={false}>
        {/* Modal content */}
        <SRInputsContext.Provider value={ctx}>
          <Tabs defaultValue="languages" className="w-full">
            {/* Header */}
            <DialogHeader>
              <DialogTitle>
                <TabsList className="grid w-full grid-cols-2">
                  <TabsTrigger value="languages">Languages</TabsTrigger>
                  <TabsTrigger value="microphone">Microphone</TabsTrigger>
                </TabsList>
              </DialogTitle>

              <DialogDescription />
            </DialogHeader>

            {/* Body */}
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
