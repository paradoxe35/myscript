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
import { Separator } from "./ui/separator";
import { ApiKeyInput } from "./ui/api-key-input";

export function SettingsModal(props: PropsWithChildren) {
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
            <ApiKeyInput className="col-span-3" />
          </div>

          <Separator />

          <div className="flex flex-col gap-4">
            <Label htmlFor="name">OpenAI API Key</Label>
            <ApiKeyInput className="col-span-3" />

            <small className="text-muted-foreground">
              If you prefer not to use the local Whisper, you can enter your
              OpenAI API key here.
            </small>
          </div>
        </div>
        <DialogFooter>
          <Button type="submit">Save changes</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
