import { useEffect, useState } from "react";
import { Label } from "../ui/label";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "../ui/select";
import { Separator } from "../ui/separator";
import { useSettings } from "./context";
import { useLocalWhisperStore } from "@/store/local-whisper";
import { Checkbox } from "../ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";
import { Eye } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../ui/table";

export function LocalWhisperInputs() {
  const { state, dispatch, bestWhisperModel, whisperModels } = useSettings();

  return (
    <>
      <div className="flex flex-col gap-3 relative">
        <Label className="text-xs text-white/70">{"Local Whisper Model"}</Label>

        <Select
          value={state.LocalWhisperModel || bestWhisperModel}
          onValueChange={(value: string) => {
            dispatch({ LocalWhisperModel: value });
          }}
        >
          <SelectTrigger className="w-[190px]">
            <SelectValue placeholder="Select a model" />
          </SelectTrigger>

          <SelectContent>
            <SelectGroup>
              <SelectLabel>Models</SelectLabel>
              {whisperModels.map((item) => {
                return (
                  <SelectItem key={item.Name} value={item.Name}>
                    {item.Name}
                  </SelectItem>
                );
              })}
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <p className="text-xs text-white/50">
        Best model based on available resources: <b>{bestWhisperModel}</b>
        <br />
        <em>Doesn't require Internet connection</em>
      </p>

      <Separator />

      <ModelsRamRequirements />

      <Separator />

      {/* <p className="text-xs text-white/50">
          <b>GPU acceleration:</b> This feature is currently not supported.
        </p> */}

      <LocalWhisperModelsDownload />
    </>
  );
}

function LocalWhisperModelsDownload() {
  const { state, bestWhisperModel, whisperModels } = useSettings();

  const modelName = state.LocalWhisperModel || bestWhisperModel;
  const model = whisperModels.find((item) => item.Name === modelName);

  return (
    <div className="flex flex-col gap-4">
      <p className="text-xs text-white/50">
        Choose the model you want to use for speech recognition.
      </p>

      <ul className="flex flex-col gap-2">
        {model?.HasAlsoAnEnglishOnlyModel && (
          <SingleModel modelName={model.Name} englishOnly />
        )}

        {model && <SingleModel modelName={model.Name} />}
      </ul>
    </div>
  );
}

function SingleModel(props: { modelName: string; englishOnly?: boolean }) {
  const [status, setStatus] = useState<
    "idle" | "downloading" | "downloaded" | "error"
  >("idle");

  const localWhisperStore = useLocalWhisperStore();

  useEffect(() => {}, []);

  return (
    <>
      <div className="flex items-center space-x-2">
        <Checkbox id={`single-model-show-${props.englishOnly}`} />
        <label
          htmlFor={`single-model-show-${props.englishOnly}`}
          className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          {props.englishOnly ? "English Only" : ""}
        </label>
      </div>
    </>
  );
}

function ModelsRamRequirements() {
  const { whisperModels } = useSettings();

  return (
    <>
      <p className="text-xs text-white/50 flex items-center gap-1">
        Here are the available models and their RAM requirements{" "}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <Eye className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>

          <DropdownMenuContent className="w-56">
            <Table>
              <TableHeader>
                <TableRow className="text-xs text-white/50">
                  <TableHead>Model</TableHead>
                  <TableHead>RAM (GB)</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {whisperModels.map((item) => {
                  return (
                    <TableRow key={item.Name}>
                      <TableCell className="font-medium">{item.Name}</TableCell>
                      <TableCell>{item.RAMRequired} GB</TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </DropdownMenuContent>
        </DropdownMenu>
      </p>
    </>
  );
}
