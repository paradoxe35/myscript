import {
  ArrowDownWideNarrow,
  CheckCheck,
  LucideProps,
  RefreshCcwDot,
  StepForward,
  WrapText,
} from "lucide-react";
import { useEditor } from "novel";
import { getPrevText } from "novel/utils";
import {
  CommandGroup,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "../ui/command";

export type Option =
  | "continue"
  | "improve"
  | "shorter"
  | "longer"
  | "fix"
  | "zap";

type OptionItem = {
  value: Option;
  label: string;
  icon: React.ForwardRefExoticComponent<
    Omit<LucideProps, "ref"> & React.RefAttributes<SVGSVGElement>
  >;
};

const options: OptionItem[] = [
  {
    value: "improve",
    label: "Improve writing",
    icon: RefreshCcwDot,
  },

  {
    value: "fix",
    label: "Fix grammar",
    icon: CheckCheck,
  },
  {
    value: "shorter",
    label: "Make shorter",
    icon: ArrowDownWideNarrow,
  },
  {
    value: "longer",
    label: "Make longer",
    icon: WrapText,
  },
];

interface AISelectorCommandsProps {
  onSelect: (value: string, option: Option) => void;
}

const AISelectorCommands = ({ onSelect }: AISelectorCommandsProps) => {
  const { editor } = useEditor();

  return (
    <CommandList>
      <CommandGroup heading="Edit or review selection">
        {options.map((option) => (
          <CommandItem
            onSelect={(value) => {
              const slice = editor?.state.selection.content();
              const text = editor?.storage.markdown.serializer.serialize(
                slice?.content
              );
              onSelect(text, value as Option);
            }}
            className="flex gap-2 px-4"
            key={option.value}
            value={option.value}
          >
            <option.icon className="h-4 w-4 text-purple-500" />
            {option.label}
          </CommandItem>
        ))}
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Use AI to do more">
        <CommandItem
          onSelect={() => {
            if (!editor) return;
            const pos = editor.state.selection.from;

            const text = getPrevText(editor, pos);
            onSelect(text, "continue");
          }}
          value="continue"
          className="gap-2 px-4"
        >
          <StepForward className="h-4 w-4 text-purple-500" />
          Continue writing
        </CommandItem>
      </CommandGroup>
    </CommandList>
  );
};

export default AISelectorCommands;
