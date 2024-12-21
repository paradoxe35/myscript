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
import { PropsWithChildren } from "react";
import { Separator } from "../ui/separator";
import { ApiKeyInput } from "../ui/api-key-input";
import { useSettings, WhisperSource } from "./context";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

export function SettingsModal(props: PropsWithChildren) {
  const { state, dispatch, handleSave } = useSettings();

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
            <Label>Notion API Key</Label>
            <ApiKeyInput
              className="col-span-3"
              tabIndex={-1}
              value={state.NotionApiKey}
              onChange={(e) => dispatch({ NotionApiKey: e.target.value })}
            />
          </div>

          <Separator />

          <div className="flex flex-col gap-4 relative">
            <Label>Speech Recognition (Whisper)</Label>
            <SelectSpeechSource />
          </div>

          <Separator />

          {state.WhisperSource === "local" && <LocalWhisperInputs />}
          {state.WhisperSource === "openai" && <OpenAIApiKeyInput />}
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

function LocalWhisperInputs() {
  return <></>;
}

function OpenAIApiKeyInput() {
  const { state, dispatch } = useSettings();

  return (
    <div className="flex flex-col gap-3 relative">
      <Label className="text-xs text-white/50">OpenAI API Key</Label>

      <ApiKeyInput
        className="col-span-3"
        tabIndex={-1}
        value={state.OpenAIApiKey}
        onChange={(e) => dispatch({ OpenAIApiKey: e.target.value })}
      />
    </div>
  );
}

function SelectSpeechSource() {
  const { state, dispatch } = useSettings();

  return (
    <Select
      value={state.WhisperSource}
      onValueChange={(value: WhisperSource) => {
        dispatch({ WhisperSource: value });
      }}
    >
      <SelectTrigger className="w-[190px]">
        <SelectValue placeholder="Select a source" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          <SelectItem value="local">Local Whisper</SelectItem>
          <SelectItem value="openai">OpenAI Whisper</SelectItem>
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
