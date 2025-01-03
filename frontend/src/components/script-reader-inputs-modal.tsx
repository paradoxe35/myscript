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
  const config = useConfigStore((state) => state.config);

  const transcriberStore = useTranscriberStore();
  const languages = transcriberStore.languages;
  const micInputDevices = transcriberStore.micInputDevices;

  const [selectedLanguageCode, setSelectedLanguageCode] = useState<
    string | null
  >(null);

  const [micInputDevice, setMicInputDevice] =
    useState<microphone.MicInputDevice | null>(null);

  const handleLanguageSelected = () => {
    requestAnimationFrame(() => {
      if (selectedLanguageCode && micInputDevice) {
        props.onLanguageSelected(selectedLanguageCode, micInputDevice.ID);
      }
    });
  };

  useEffect(() => {
    transcriberStore.getLanguages();
  }, [config?.TranscriberSource]);

  useEffect(() => {
    transcriberStore.getMicInputDevices().then((micInputDevices) => {
      const defaultDevice = micInputDevices.find(
        (device) => device.IsDefault === 1
      );

      if (defaultDevice) {
        setMicInputDevice(defaultDevice);
      }
    });
  }, []);

  const canSubmit = selectedLanguageCode && micInputDevice;

  return {
    selectedLanguageCode,
    setSelectedLanguageCode,
    handleLanguageSelected,
    languages,
    canSubmit,

    micInputDevice,
    micInputDevices,
    setMicInputDevice,
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

export default function SRInputsModal(props: Props) {
  const ctx = useSRInputs(props);

  return (
    <Dialog>
      <DialogTrigger asChild>{props.trigger}</DialogTrigger>

      <DialogContent className="sm:max-w-md z-50" showCloseButton={false}>
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
              onClick={ctx.handleLanguageSelected}
            >
              Start Reading
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
