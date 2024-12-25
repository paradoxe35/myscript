import { Check, ChevronsUpDown } from "lucide-react";

import { cn } from "@/lib/utils";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
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

  const [open, setOpen] = useState(false);
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

      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Select your script language</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              role="combobox"
              aria-expanded={open}
              className="w-full justify-between"
            >
              {selectedLanguage
                ? languages.find((lang) => lang.Code === selectedLanguage)?.Name
                : "Select Language..."}
              <ChevronsUpDown className="opacity-50" />
            </Button>
          </PopoverTrigger>

          <PopoverContent className="w-full p-0">
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
                        setOpen(false);
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
          </PopoverContent>
        </Popover>

        <DialogFooter className="sm:justify-start">
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
