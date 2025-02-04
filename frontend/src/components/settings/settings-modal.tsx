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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../ui/tabs";
import { SettingsCloud } from "./settings-cloud";
import { cn } from "@/lib/utils";
import { useTheme } from "../theme-provider";
import { MoonIcon, SunIcon } from "lucide-react";

export function SettingsModal(props: PropsWithChildren) {
  const { state, cloud, appVersion, configModified, handleSave } =
    useSettings();

  return (
    <Dialog>
      <DialogTrigger asChild>{props.children}</DialogTrigger>

      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle>Settings</DialogTitle>
          <DialogDescription className="text-xs dark:text-white/50 text-dark/50">
            Version: {appVersion}
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="api-keys" className="w-full">
          <TabsList
            className={cn(
              "grid w-full grid-cols-3",
              !cloud.cloudEnabled && "grid-cols-2"
            )}
          >
            <TabsTrigger value="api-keys">API Keys</TabsTrigger>

            <TabsTrigger value="speech-recognition">
              Speech Recognition
            </TabsTrigger>

            {cloud.cloudEnabled && (
              <TabsTrigger value="cloud">Cloud</TabsTrigger>
            )}
          </TabsList>

          {/* Notion API Key */}
          <TabsContent value="api-keys" className="min-h-48">
            <div className="grid gap-4 py-4">
              <NotionInputs />
              <Separator />

              {/* OpenAI API Key */}
              <OpenAIApiKey />
              <Separator />
            </div>
          </TabsContent>

          {/* Speech Recognition */}
          <TabsContent value="speech-recognition" className="min-h-48">
            <div className="grid gap-4 py-4">
              <div className="flex flex-col gap-4 relative">
                <Label>Speech Recognition (Whisper, Wit.ai)</Label>
                <SelectSpeechSource />
              </div>

              {state.TranscriberSource === "local" && <LocalWhisperInputs />}

              {state.TranscriberSource === "openai" && (
                <OpenAIApiKeyTranscriber />
              )}

              {state.TranscriberSource === "witai" && <WitAIHint />}

              {state.TranscriberSource === "groq" && (
                <>
                  <Separator />

                  <GroqApiKey />
                </>
              )}
            </div>
          </TabsContent>

          {/* Cloud */}
          {cloud.cloudEnabled && (
            <TabsContent value="cloud" className="min-h-48">
              <SettingsCloud />
            </TabsContent>
          )}
        </Tabs>

        <DialogFooter className="sm:justify-between items-center">
          <ThemeSwitch />

          <Button type="submit" disabled={!configModified} onClick={handleSave}>
            Save changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function ThemeSwitch() {
  const { theme, setTheme } = useTheme();

  return (
    <Button
      variant="secondary"
      size="icon"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
    >
      {theme === "dark" ? <MoonIcon /> : <SunIcon />}
    </Button>
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
      <p className="text-xs dark:text-white/50 text-dark/50">
        <b>Wit.ai</b> doesn't require any extra configuration. However, it may
        not be as accurate as OpenAI.
      </p>

      <p className="text-xs dark:text-white/50 text-dark/50">
        <em>Internet connection is required</em>
      </p>
    </>
  );
}

function OpenAIApiKey() {
  const { state, dispatch } = useSettings();

  return (
    <div className="flex flex-col gap-3 relative">
      <Label className="text-xs">OpenAI API Key</Label>

      <ApiKeyInput
        className="col-span-3"
        tabIndex={-1}
        value={state.OpenAIApiKey}
        onChange={(e) => dispatch({ OpenAIApiKey: e.target.value })}
      />

      <p className="text-xs dark:text-white/50 text-dark/50">
        <em>For speech recognition and text generation.</em>
      </p>
    </div>
  );
}

const GROQ_TRANSCRIBE_MODEL = "whisper-large-v3-turbo";

function GroqApiKey() {
  const { state, dispatch } = useSettings();

  return (
    <div className="flex flex-col gap-3 relative">
      <Label className="text-xs">Groq API Key</Label>

      <ApiKeyInput
        className="col-span-3"
        tabIndex={-1}
        value={state.GroqApiKey}
        onChange={(e) => dispatch({ GroqApiKey: e.target.value })}
      />

      <p className="text-xs dark:text-white/50 text-dark/50">
        <em>It uses the {GROQ_TRANSCRIBE_MODEL} model.</em>
      </p>
    </div>
  );
}

function OpenAIApiKeyTranscriber() {
  return (
    <>
      <p className="text-xs dark:text-white/50 text-dark/50">
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
