import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverAnchor,
  PopoverContent,
} from "@/components/ui/popover";
import { useEditorAIStore } from "@/store/editor-ai";
import { ArrowUp, Check, RotateCcw, Trash2 } from "lucide-react";
import { getPrevText, useEditor } from "novel";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import CrazySpinner from "../ui/icons/crazy-spinner";
import Magic from "../ui/icons/magic";
import { AIPreview } from "./ai-preview";
import { insertMarkdown } from "./markdown";
import { useAICompletion } from "./use-ai-completion";

type Caret = { top: number; left: number };

/** The slash-command entry point: a prompt at the caret, with nothing selected. */
export function AIPrompt() {
  const { editor } = useEditor();

  const open = useEditorAIStore((store) => store.prompting);
  const close = useEditorAIStore((store) => store.closePrompt);

  const { generate, cancel, completion, isLoading, error } = useAICompletion();
  const [instruction, setInstruction] = useState("");
  const [caret, setCaret] = useState<Caret | null>(null);

  useEffect(() => {
    if (!open || !editor) return;

    const { bottom, left } = editor.view.coordsAtPos(
      editor.state.selection.from,
    );
    setCaret({ top: bottom + 8, left });
    setInstruction("");
  }, [open, editor]);

  useEffect(() => {
    if (error) toast.error(error.message);
  }, [error]);

  const dismiss = () => {
    cancel();
    setInstruction("");
    close();
  };

  const ask = () => {
    if (!editor || instruction.trim() === "") return;

    const before = getPrevText(editor, editor.state.selection.from);
    generate(before, "write", instruction.trim());
  };

  const insert = () => {
    if (editor) insertMarkdown(editor, completion);
    dismiss();
  };

  const hasCompletion = completion.length > 0;

  return (
    <Popover open={open} onOpenChange={(next) => !next && dismiss()}>
      <PopoverAnchor asChild>
        <div
          className="pointer-events-none fixed h-0 w-0"
          style={{ top: caret?.top ?? 0, left: caret?.left ?? 0 }}
        />
      </PopoverAnchor>

      <PopoverContent align="start" className="w-[420px] p-0" sideOffset={0}>
        <div>
          {hasCompletion && <AIPreview markdown={completion} />}

          {isLoading && (
            <div className="flex h-12 items-center px-3 text-sm font-medium text-blue-500">
              <Magic className="mr-2 h-4 w-4 shrink-0" />
              AI is writing
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
            <div className="relative">
              <Magic className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-blue-500" />
              <Input
                autoFocus
                value={instruction}
                onChange={(event) => setInstruction(event.target.value)}
                placeholder={
                  hasCompletion
                    ? "Tell AI what to change"
                    : "Ask AI to write anything…"
                }
                onKeyDown={(event) => {
                  if (event.key === "Enter" && !event.shiftKey) {
                    event.preventDefault();
                    ask();
                  }
                }}
                className="border-0 pl-9 pr-11 shadow-none focus-visible:ring-0"
              />
              <Button
                size="icon"
                className="absolute right-2 top-1/2 h-6 w-6 -translate-y-1/2 rounded-full"
                disabled={instruction.trim() === ""}
                onClick={ask}
              >
                <ArrowUp className="h-4 w-4" />
              </Button>
            </div>
          )}

          {hasCompletion && !isLoading && (
            <div className="flex items-center gap-1 border-t p-1">
              <Button
                variant="ghost"
                size="sm"
                className="gap-2"
                onClick={insert}
              >
                <Check className="h-4 w-4 text-muted-foreground" />
                Insert
              </Button>
              <Button variant="ghost" size="sm" className="gap-2" onClick={ask}>
                <RotateCcw className="h-4 w-4 text-muted-foreground" />
                Try again
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="ml-auto gap-2"
                onClick={dismiss}
              >
                <Trash2 className="h-4 w-4 text-muted-foreground" />
                Discard
              </Button>
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
