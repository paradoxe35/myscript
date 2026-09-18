import { EditorBubble, removeAIHighlight, useEditor } from "novel";
import { Fragment, type ReactNode, useEffect } from "react";
import Magic from "../ui/icons/magic";
import { AISelector } from "./ai-selector";
import { Button } from "@/components/ui/button";
import { useAIProvidersStore } from "@/store/ai-providers";

interface GenerativeMenuSwitchProps {
  children: ReactNode;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const GenerativeMenuSwitch = ({
  children,
  open,
  onOpenChange,
}: GenerativeMenuSwitchProps) => {
  const { editor } = useEditor();
  const aiEnabled = useAIProvidersStore((store) => store.enabled);
  const checkEnabled = useAIProvidersStore((store) => store.checkEnabled);

  useEffect(() => {
    checkEnabled();
  }, []);

  useEffect(() => {
    if (!open && editor) removeAIHighlight(editor);
  }, [open]);

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
        <AISelector open={open} onOpenChange={onOpenChange} />
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
