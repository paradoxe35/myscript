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
import { Separator } from "../ui/separator";
import { ApiKeyInput } from "../ui/api-key-input";
import { toast } from "sonner";
import { useConfigStore } from "@/store/config";

export function SettingsModal(props: PropsWithChildren) {
  const [notionApiKey, setNotionApiKey] = useState("");
  const [openAiApiKey, setOpenAiApiKey] = useState("");

  const configStore = useConfigStore();

  useEffect(() => {
    configStore.fetchConfig();
  }, []);

  useEffect(() => {
    const config = configStore.config;

    if (config) {
      setNotionApiKey(config.NotionApiKey || "");
      setOpenAiApiKey(config.OpenAIApiKey || "");
    }
  }, [configStore.config]);

  const handleSave = () => {
    configStore
      .writeConfig({
        NotionApiKey: notionApiKey,
        OpenAIApiKey: openAiApiKey,
      })
      .then(() => {
        toast.success("Settings saved successfully!");
      });
  };

  return (
    <Dialog>
      <DialogTrigger asChild>{props.children}</DialogTrigger>

      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="flex flex-col gap-4">
            <Label htmlFor="name">Notion API Key</Label>
            <ApiKeyInput
              className="col-span-3"
              tabIndex={-1}
              value={notionApiKey}
              onChange={(e) => setNotionApiKey(e.target.value)}
            />
          </div>

          <Separator />

          <div className="flex flex-col gap-4">
            <Label htmlFor="name">OpenAI API Key</Label>
            <ApiKeyInput
              className="col-span-3"
              tabIndex={-1}
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
