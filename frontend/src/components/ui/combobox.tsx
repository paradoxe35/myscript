import * as React from "react";
import { Check, ChevronsUpDown, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
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
import { cn } from "@/lib/utils";

export type ComboboxOption = {
  value: string;
  label?: string;
  hint?: string;
};

type ComboboxProps = {
  value: string;
  options: ComboboxOption[];
  onChange: (value: string) => void;
  placeholder?: string;
  searchPlaceholder?: string;
  /** Lets the user commit whatever they typed, for values not in the list. */
  allowCustom?: boolean;
  emptyLabel?: string;
  disabled?: boolean;
  className?: string;
};

export function Combobox({
  value,
  options,
  onChange,
  placeholder = "Select…",
  searchPlaceholder = "Search or type…",
  allowCustom = false,
  emptyLabel = "Nothing found.",
  disabled,
  className,
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");

  const matches = React.useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return options;

    return options.filter(
      (option) =>
        option.value.toLowerCase().includes(needle) ||
        option.label?.toLowerCase().includes(needle),
    );
  }, [options, query]);

  const typed = query.trim();
  const isNew =
    allowCustom &&
    typed !== "" &&
    !options.some(
      (option) => option.value.toLowerCase() === typed.toLowerCase(),
    );

  const commit = (next: string) => {
    onChange(next);
    setQuery("");
    setOpen(false);
  };

  const selected = options.find((option) => option.value === value);

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) setQuery("");
      }}
    >
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            "w-full justify-between font-normal",
            !value && "text-muted-foreground",
            className,
          )}
        >
          <span className="truncate">
            {value ? (selected?.label ?? value) : placeholder}
          </span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>

      <PopoverContent
        className="w-[--radix-popover-trigger-width] p-0"
        align="start"
      >
        {/* Filtering is done above so the "use what I typed" row always shows. */}
        <Command shouldFilter={false}>
          <CommandInput
            value={query}
            onValueChange={setQuery}
            placeholder={searchPlaceholder}
          />

          <CommandList>
            {matches.length === 0 && !isNew && (
              <CommandEmpty>{emptyLabel}</CommandEmpty>
            )}

            {isNew && (
              <CommandGroup>
                <CommandItem value={typed} onSelect={() => commit(typed)}>
                  <Plus className="mr-2 h-4 w-4 shrink-0" />
                  <span className="truncate">Use “{typed}”</span>
                </CommandItem>
              </CommandGroup>
            )}

            {matches.length > 0 && (
              <CommandGroup>
                {matches.map((option) => (
                  <CommandItem
                    key={option.value}
                    value={option.value}
                    onSelect={() => commit(option.value)}
                  >
                    <Check
                      className={cn(
                        "mr-2 h-4 w-4 shrink-0",
                        option.value === value ? "opacity-100" : "opacity-0",
                      )}
                    />
                    <span className="flex min-w-0 flex-col">
                      <span className="truncate">
                        {option.label ?? option.value}
                      </span>
                      {option.hint && (
                        <span className="truncate text-xs text-muted-foreground">
                          {option.hint}
                        </span>
                      )}
                    </span>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
