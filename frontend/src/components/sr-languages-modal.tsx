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
import { useEffect, useState } from "react";

type Props = {
  trigger: React.ReactNode | null;
  onLanguageSelected: (languageCode: string) => void;
};

export function SRLanguagesModal(props: Props) {
  const transcriberStore = useTranscriberStore();
  const languages = transcriberStore.languages;

  const [selectedLanguage, setSelectedLanguage] = useState<string | null>(null);

  const handleLanguageSelected = () => {
    requestAnimationFrame(() => {
      selectedLanguage && props.onLanguageSelected(selectedLanguage);
    });
  };

  useEffect(() => {
    transcriberStore.getLanguages();
  }, []);

  return (
    <Dialog>
      <DialogTrigger asChild>{props.trigger}</DialogTrigger>

      <DialogContent className="sm:max-w-md z-50">
        <DialogHeader>
          <DialogTitle>Select your script language</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        <Command>
          <CommandInput placeholder="Search language..." className="h-9" />
          <CommandList>
            <CommandEmpty>No language found.</CommandEmpty>
            <CommandGroup>
              {languages.map((language) => (
                <CommandItem
                  key={language.Code}
                  value={language.Code}
                  onSelect={(currentValue) => {
                    setSelectedLanguage(currentValue);
                  }}
                >
                  {language.Name}
                  <Check
                    className={cn(
                      "ml-auto",
                      selectedLanguage === language.Code
                        ? "opacity-100"
                        : "opacity-0"
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>

        <DialogFooter className="sm:justify-end">
          <DialogClose asChild>
            <Button
              type="button"
              disabled={!selectedLanguage}
              onClick={handleLanguageSelected}
            >
              Start Reading
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
