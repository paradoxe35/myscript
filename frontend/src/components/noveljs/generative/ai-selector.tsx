import { Command, CommandInput } from "@/components/ui/command";

import { ArrowUp } from "lucide-react";
import { addAIHighlight, useEditor } from "novel";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import CrazySpinner from "../ui/icons/crazy-spinner";
import Magic from "../ui/icons/magic";
import AICompletionCommands from "./ai-completion-command";
import AISelectorCommands from "./ai-selector-commands";
import { AIPreview } from "./ai-preview";
import { useAICompletion } from "./use-ai-completion";
import { Button } from "@/components/ui/button";

interface AISelectorProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AISelector({ onOpenChange }: AISelectorProps) {
  const { editor } = useEditor();
  const [inputValue, setInputValue] = useState("");

  const { completion, generate, cancel, isLoading, error } = useAICompletion();

  useEffect(() => {
    if (error) {
      toast.error(error.message);
    }
  }, [error]);

  const hasCompletion = completion.length > 0;

  return (
    <Command className="w-[350px]">
      {hasCompletion && (
        <AIPreview markdown={completion} className="max-h-[400px]" />
      )}

      {isLoading && (
        <div className="flex h-12 w-full items-center px-4 text-sm font-medium text-blue-500">
          <Magic className="mr-2 h-4 w-4 shrink-0" />
          AI is thinking
          <div className="ml-2 mt-1">
            <CrazySpinner />
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="ml-auto text-muted-foreground"
            onClick={cancel}
          >
            Stop
          </Button>
        </div>
      )}

      {!isLoading && (
        <>
          <div className="relative">
            <CommandInput
              value={inputValue}
              onValueChange={setInputValue}
              autoFocus
              className="mr-6"
              placeholder={
                hasCompletion
                  ? "Tell AI what to do next"
                  : "Ask AI to edit or generate..."
              }
              onFocus={() => editor && addAIHighlight(editor)}
            />
            <Button
              size="icon"
              className="absolute right-2 top-1/2 h-6 w-6 -translate-y-1/2 rounded-full bg-blue-500 hover:bg-blue-900"
              onClick={() => {
                if (completion) {
                  return generate(completion, "zap", inputValue).then(() =>
                    setInputValue(""),
                  );
                }

                const slice = editor?.state.selection.content();
                const text = editor?.storage.markdown.serializer.serialize(
                  slice?.content,
                );

                generate(text, "zap", inputValue).then(() => setInputValue(""));
              }}
            >
              <ArrowUp className="h-4 w-4" />
            </Button>
          </div>

          {hasCompletion ? (
            <AICompletionCommands
              onDiscard={() => {
                editor?.chain().unsetHighlight().focus().run();
                onOpenChange(false);
              }}
              completion={completion}
            />
          ) : (
            <AISelectorCommands
              onSelect={(value, option) => {
                generate(value, option);
              }}
            />
          )}
        </>
      )}
    </Command>
  );
}
