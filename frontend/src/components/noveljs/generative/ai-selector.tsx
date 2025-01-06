import { Command, CommandInput } from "@/components/ui/command";

import { ArrowUp } from "lucide-react";
import { useEditor } from "novel";
import { addAIHighlight } from "novel/extensions";
import { useEffect, useState } from "react";
import Markdown from "react-markdown";
import { toast } from "sonner";
import CrazySpinner from "../ui/icons/crazy-spinner";
import Magic from "../ui/icons/magic";
import AICompletionCommands from "./ai-completion-command";
import AISelectorCommands from "./ai-selector-commands";
import { useOpenAICompletion } from "./use-openai-completion";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";

interface AISelectorProps {
  openAIApiKey: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AISelector({ onOpenChange, openAIApiKey }: AISelectorProps) {
  const { editor } = useEditor();
  const [inputValue, setInputValue] = useState("");

  const { completion, generateCompletion, isLoading, error } =
    useOpenAICompletion(openAIApiKey);

  useEffect(() => {
    if (error) {
      toast.error(error.message);
    }
  }, [error]);

  const hasCompletion = completion.length > 0;

  return (
    <Command className="w-[350px]">
      {hasCompletion && (
        <div className="flex max-h-[400px]">
          <ScrollArea>
            <div className="prose p-2 px-4 prose-sm dark:prose-invert">
              <Markdown>{completion}</Markdown>
            </div>
          </ScrollArea>
        </div>
      )}

      {isLoading && (
        <div className="flex h-12 w-full items-center px-4 text-sm font-medium text-muted-foreground text-blue-500">
          <Magic className="mr-2 h-4 w-4 shrink-0  " />
          AI is thinking
          <div className="ml-2 mt-1">
            <CrazySpinner />
          </div>
        </div>
      )}

      {!isLoading && (
        <>
          <div className="relative">
            <CommandInput
              value={inputValue}
              onValueChange={setInputValue}
              autoFocus
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
                  return generateCompletion(completion, "zap", inputValue).then(
                    () => setInputValue("")
                  );
                }

                const slice = editor?.state.selection.content();
                const text = editor?.storage.markdown.serializer.serialize(
                  slice?.content
                );

                generateCompletion(text, "zap", inputValue).then(() =>
                  setInputValue("")
                );
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
                generateCompletion(value, option);
              }}
            />
          )}
        </>
      )}
    </Command>
  );
}
