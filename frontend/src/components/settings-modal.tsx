import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { PropsWithChildren, useEffect, useState } from "react";
import { Separator } from "./ui/separator";
import { ApiKeyInput } from "./ui/api-key-input";
import { useAtom } from "jotai";
import { configAtom } from "@/store/config";
import { repository } from "~wails/models";
import { toast } from "sonner";

export function SettingsModal(props: PropsWithChildren) {
  const [notionApiKey, setNotionApiKey] = useState("");
  const [openAiApiKey, setOpenAiApiKey] = useState("");

  const [config, writeConfig] = useAtom(configAtom);

  useEffect(() => {
    console.log(config);

    if (config) {
      setNotionApiKey(config.NotionApiKey || "");
      setOpenAiApiKey(config.OpenAIApiKey || "");
    }
  }, [config]);

  const handleSave = () => {
    writeConfig(
      repository.Config.createFrom({
        ...config,
        NotionApiKey: notionApiKey,
        OpenAIApiKey: openAiApiKey,
      })
    ).then(() => {
      toast.success("Settings saved successfully!");
    });
  };

  return (
    <Dialog>
      <DialogTrigger asChild>{props.children}</DialogTrigger>

      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="flex flex-col gap-4">
            <Label htmlFor="name">Notion API Key</Label>
            <ApiKeyInput
              className="col-span-3"
              value={notionApiKey}
              onChange={(e) => setNotionApiKey(e.target.value)}
            />
          </div>

          <Separator />

          <div className="flex flex-col gap-4">
            <Label htmlFor="name">OpenAI API Key</Label>
            <ApiKeyInput
              className="col-span-3"
              value={openAiApiKey}
              onChange={(e) => setOpenAiApiKey(e.target.value)}
            />

            <small className="text-muted-foreground">
              If you prefer not to use the local Whisper, you can enter your
              OpenAI API key here.
            </small>
          </div>
        </div>
        <DialogFooter>
          <Button type="submit" onClick={handleSave}>
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
