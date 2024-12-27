import { useEffect, useMemo, useState } from "react";
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
import { local_whisper } from "~wails/models";

export function LocalWhisperInputs() {
  const { state, dispatch, bestWhisperModel, whisperModels } = useSettings();
  const modelName = state.LocalWhisperModel || bestWhisperModel;

  return (
    <>
      <div className="flex flex-col gap-3 relative">
        <Label className="text-xs text-white/70">{"Local Whisper Model"}</Label>

        <Select
          value={modelName}
          onValueChange={(value: string) => {
            dispatch({ LocalWhisperModel: value });
          }}
        >
          <SelectTrigger className="w-[190px]">
            <SelectValue placeholder="Select a model" />
          </SelectTrigger>

          <SelectContent>
            <SelectGroup>
              {/* <SelectLabel>Models</SelectLabel> */}
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

      {/* 
        <ModelsRamRequirements />
        <Separator />

        <p className="text-xs text-white/50">
          <b>GPU acceleration:</b> This feature is currently not supported.
        </p> 
    */}

      <LocalWhisperModelsDownload key={modelName} />
    </>
  );
}

function LocalWhisperModelsDownload() {
  const { state, bestWhisperModel, whisperModels } = useSettings();
  const localWhisperStore = useLocalWhisperStore();

  const [downloading, setDownloading] = useState(false);

  const [selectedModels, setSelectedModels] = useState<
    local_whisper.LocalWhisperModel[]
  >([]);

  const modelName = state.LocalWhisperModel || bestWhisperModel;
  const model = whisperModels.find((item) => item.Name === modelName);

  useEffect(() => {
    localWhisperStore.areSomeModelsDownloading().then((status) => {
      setDownloading(status);
    });
  }, []);

  useEffect(() => {
    setSelectedModels([]);
  }, [modelName]);

  const startDownload = async () => {
    console.log("Downloading... models:", selectedModels);

    localWhisperStore.downloadModels(selectedModels).then((res) => {
      console.log("Downloaded models:", res);
    });
  };

  return (
    <div className="flex flex-col gap-4">
      <p className="text-xs text-white/50">
        Choose the models you want to use for speech recognition. If some of
        your scripts are in English, consider downloading the English-only
        model.
      </p>

      <ul className="flex flex-col gap-3">
        {model?.HasAlsoAnEnglishOnlyModel && (
          <SingleModel
            selectedModels={selectedModels}
            setSelectedModels={setSelectedModels}
            modelName={model.Name}
            englishOnly={true}
          />
        )}

        {model && (
          <SingleModel
            selectedModels={selectedModels}
            setSelectedModels={setSelectedModels}
            modelName={model.Name}
            englishOnly={false}
          />
        )}
      </ul>

      <div>
        <Button
          size="sm"
          variant="secondary"
          onClick={startDownload}
          disabled={selectedModels.length === 0}
        >
          Download
        </Button>
      </div>
    </div>
  );
}

type SingleModelProps = {
  selectedModels: local_whisper.LocalWhisperModel[];
  setSelectedModels: React.Dispatch<
    React.SetStateAction<local_whisper.LocalWhisperModel[]>
  >;

  modelName: string;
  englishOnly: boolean;
};

function SingleModel(props: SingleModelProps) {
  const [status, setStatus] = useState<"idle" | "downloading" | "downloaded">(
    "idle"
  );

  const localWhisperStore = useLocalWhisperStore();

  const localWhisperModelObject = useMemo(
    () => localWhisperStore.newModelObject(props.modelName, props.englishOnly),
    []
  );

  useEffect(() => {
    (async () => {
      const downloading = await localWhisperStore.isModelDownloading(
        localWhisperModelObject
      );

      if (downloading) {
        setStatus("downloading");
        return;
      }

      const exists = await localWhisperStore.existsModel(
        localWhisperModelObject
      );

      if (exists) {
        setStatus("downloaded");
      } else {
        setStatus("idle");
      }
    })();
  }, []);

  const selected = props.selectedModels.some(
    (item) =>
      item.Name === props.modelName && item.EnglishOnly === props.englishOnly
  );

  const toggleSelected = () => {
    if (status !== "idle") {
      return;
    }

    if (selected) {
      props.setSelectedModels(
        props.selectedModels.filter(
          (item) =>
            !(
              item.Name === props.modelName &&
              item.EnglishOnly === props.englishOnly
            )
        )
      );
    } else {
      props.setSelectedModels((prev) => [...prev, localWhisperModelObject]);
    }
  };

  return (
    <>
      <div className="flex items-center space-x-2">
        <Checkbox
          checked={
            status === "downloaded" || status === "downloading" || selected
          }
          id={`single-model-show-${props.englishOnly}`}
          disabled={status !== "idle"}
          onCheckedChange={toggleSelected}
        />

        <label
          htmlFor={`single-model-show-${props.englishOnly}`}
          className="text-sm opacity-90 leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          {props.englishOnly ? "English Only" : "Multilingual"}
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
