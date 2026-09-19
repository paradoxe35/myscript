import {
  EditorCommand,
  EditorCommandEmpty,
  EditorCommandItem,
  EditorCommandList,
  EditorContent,
  type EditorInstance,
  EditorRoot,
  ImageResizer,
  type JSONContent,
  handleCommandNavigation,
  handleImageDrop,
  handleImagePaste,
} from "novel";
import { useEffect, useState } from "react";
import { defaultExtensions } from "./extensions";
import { ColorSelector } from "./selectors/color-selector";
import { LinkSelector } from "./selectors/link-selector";
import { NodeSelector } from "./selectors/node-selector";
import { MathSelector } from "./selectors/math-selector";
import { Separator } from "@/components/ui/separator";

import GenerativeMenuSwitch from "./generative/generative-menu-switch";
import { AIPrompt } from "./generative/ai-prompt";
import { uploadFn } from "./image-upload";
import { TextButtons } from "./selectors/text-buttons";
import { slashCommand, suggestionItems } from "./slash-command";
import { useAIProvidersStore } from "@/store/ai-providers";

import { cn } from "@/lib/utils";

const extensions = [...defaultExtensions, slashCommand];

type EditorProps = {
  className?: string;
  initialContent?: JSONContent;
  onUpdate?: (editor: EditorInstance) => void;
};

const NovelEditor = (props: EditorProps) => {
  const aiEnabled = useAIProvidersStore((store) => store.enabled);
  const checkAIEnabled = useAIProvidersStore((store) => store.checkEnabled);

  const [openNode, setOpenNode] = useState(false);
  const [openColor, setOpenColor] = useState(false);
  const [openLink, setOpenLink] = useState(false);
  const [openAI, setOpenAI] = useState(false);

  useEffect(() => {
    checkAIEnabled();
  }, []);

  const commands = suggestionItems.filter(
    (item) => aiEnabled || item.title !== "Ask AI",
  );

  return (
    <div className={cn("relative w-full max-w-screen-md", props.className)}>
      <EditorRoot>
        <EditorContent
          className="relative min-h-[500px] w-full max-w-screen-lg bg-background"
          initialContent={props.initialContent}
          extensions={extensions}
          editorProps={{
            handleDOMEvents: {
              keydown: (_view, event) => handleCommandNavigation(event),
            },
            handlePaste: (view, event) =>
              handleImagePaste(view, event, uploadFn),
            handleDrop: (view, event, _slice, moved) =>
              handleImageDrop(view, event, moved, uploadFn),
            attributes: {
              class:
                "prose prose-lg dark:prose-invert prose-zoom focus:outline-none max-w-full !pt-0",
            },
          }}
          onUpdate={({ editor }) => props.onUpdate?.(editor)}
          slotAfter={<ImageResizer />}
        >
          <EditorCommand className="h-auto max-h-[330px] overflow-y-auto overscroll-contain rounded-md border border-muted bg-background px-1 py-2 shadow-md transition-all">
            <EditorCommandEmpty className="px-2 text-muted-foreground">
              No results
            </EditorCommandEmpty>
            <EditorCommandList>
              {commands.map((item) => (
                <EditorCommandItem
                  value={item.title}
                  onCommand={(val) => item.command?.(val)}
                  className="flex w-full items-center space-x-2 rounded-md px-2 py-1 text-left text-sm hover:bg-accent aria-selected:bg-accent"
                  key={item.title}
                >
                  <div className="flex h-10 w-10 items-center justify-center rounded-md border border-muted bg-background">
                    {item.icon}
                  </div>
                  <div>
                    <p className="font-medium">{item.title}</p>
                    <p className="text-xs text-muted-foreground">
                      {item.description}
                    </p>
                  </div>
                </EditorCommandItem>
              ))}
            </EditorCommandList>
          </EditorCommand>

          <GenerativeMenuSwitch open={openAI} onOpenChange={setOpenAI}>
            <Separator orientation="vertical" className="h-auto" />
            <NodeSelector open={openNode} onOpenChange={setOpenNode} />
            <Separator orientation="vertical" className="h-auto" />
            <LinkSelector open={openLink} onOpenChange={setOpenLink} />
            <Separator orientation="vertical" className="h-auto" />
            <MathSelector />
            <Separator orientation="vertical" className="h-auto" />
            <TextButtons />
            <Separator orientation="vertical" className="h-auto" />
            <ColorSelector open={openColor} onOpenChange={setOpenColor} />
          </GenerativeMenuSwitch>
          <AIPrompt />
        </EditorContent>
      </EditorRoot>
    </div>
  );
};

export default NovelEditor;
