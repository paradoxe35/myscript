import { EditorBubble, useEditor } from "novel";
import { removeAIHighlight } from "novel/extensions";
import {} from "novel/plugins";
import { Fragment, type ReactNode, useEffect } from "react";
import Magic from "../ui/icons/magic";
import { AISelector } from "./ai-selector";
import { Button } from "@/components/noveljs/ui/button";

interface GenerativeMenuSwitchProps {
  openAIApiKey?: string;
  children: ReactNode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}
const GenerativeMenuSwitch = ({
  openAIApiKey,
  children,
  open,
  onOpenChange,
}: GenerativeMenuSwitchProps) => {
  const { editor } = useEditor();

  useEffect(() => {
    if (!open && editor) removeAIHighlight(editor);
  }, [open]);

  const aiEnabled = !!openAIApiKey;

  return (
    <EditorBubble
      tippyOptions={{
        placement: open ? "bottom-start" : "top",
        onHidden: () => {
          onOpenChange(false);
          editor?.chain().unsetHighlight().run();
        },
      }}
      className="flex w-fit max-w-[90vw] overflow-hidden rounded-md border border-muted bg-background shadow-xl"
    >
      {open && aiEnabled && (
        <AISelector
          open={open}
          openAIApiKey={openAIApiKey}
          onOpenChange={onOpenChange}
        />
      )}

      {!open && (
        <Fragment>
          {aiEnabled && (
            <Button
              className="gap-1 rounded-none text-blue-500"
              variant="ghost"
              onClick={() => onOpenChange(true)}
              size="sm"
            >
              <Magic className="h-5 w-5" />
              Ask AI
            </Button>
          )}

          {children}
        </Fragment>
      )}
    </EditorBubble>
  );
};

export default GenerativeMenuSwitch;
