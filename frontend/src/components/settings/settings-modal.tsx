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
import { TRANSCRIBER_SOURCES, useSettings, TranscriberSource } from "./context";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { LocalWhisperInputs } from "./settings-local-whisper";
import { GetAppVersion } from "~wails/main/App";

export function SettingsModal(props: PropsWithChildren) {
  const [version, setVersion] = useState("");
  const { state, configModified, handleSave } = useSettings();

  useEffect(() => {
    GetAppVersion().then((v) => setVersion(v));
  }, []);

  return (
    <Dialog>
      <DialogTrigger asChild>{props.children}</DialogTrigger>

      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
          <DialogDescription className="text-xs text-white/50">
            Version: {version}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <NotionInputs />

          <Separator />

          <div className="flex flex-col gap-4 relative">
            <Label>Speech Recognition (Whisper, Wit.ai)</Label>
            <SelectSpeechSource />
          </div>

          <Separator />

          {state.TranscriberSource === "local" && <LocalWhisperInputs />}
          {state.TranscriberSource === "openai" && <OpenAIApiKeyInput />}
          {state.TranscriberSource === "witai" && <WitAIHint />}
        </div>

        <DialogFooter>
          <Button type="submit" disabled={!configModified} onClick={handleSave}>
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function NotionInputs() {
  const { state, dispatch } = useSettings();

  return (
    <div className="flex flex-col gap-4">
      <Label>Notion API Key</Label>
      <ApiKeyInput
        className="col-span-3"
        tabIndex={-1}
        value={state.NotionApiKey}
        onChange={(e) => dispatch({ NotionApiKey: e.target.value })}
      />
    </div>
  );
}

function WitAIHint() {
  return (
    <>
      <p className="text-xs text-white/50">
        <b>Wit.ai</b> doesn't require any extra configuration. However, it may
        not be as accurate as OpenAI Whisper.
      </p>

      <p className="text-xs text-white/50">
        <em>Internet connection is required</em>
      </p>
    </>
  );
}

function OpenAIApiKeyInput() {
  const { state, dispatch } = useSettings();

  return (
    <>
      <div className="flex flex-col gap-3 relative">
        <Label className="text-xs text-white/50">OpenAI API Key</Label>

        <ApiKeyInput
          className="col-span-3"
          tabIndex={-1}
          value={state.OpenAIApiKey}
          onChange={(e) => dispatch({ OpenAIApiKey: e.target.value })}
        />
      </div>

      <p className="text-xs text-white/50">
        <em>Internet connection is required</em>
      </p>
    </>
  );
}

function SelectSpeechSource() {
  const { state, dispatch } = useSettings();

  const items = Object.values(TRANSCRIBER_SOURCES);

  return (
    <Select
      value={state.TranscriberSource}
      onValueChange={(value: TranscriberSource) => {
        dispatch({ TranscriberSource: value });
      }}
    >
      <SelectTrigger className="w-[190px]">
        <SelectValue placeholder="Select a source" />
      </SelectTrigger>

      <SelectContent>
        <SelectGroup>
          {items.map((item) => {
            return (
              <SelectItem key={item.key} value={item.key}>
                {item.name}
              </SelectItem>
            );
          })}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
