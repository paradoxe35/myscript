import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PropsWithChildren, useState } from "react";
import { useLocalPagesStore } from "@/store/local-pages";

export function NewFolderModal(props: PropsWithChildren<{}>) {
  const [name, setName] = useState("");
  const localPagesStore = useLocalPagesStore();

  const handleCreate = async () => {
    const folderName = name.trim();

    if (!folderName) {
      return;
    }

    setName("");
    localPagesStore.newFolder(folderName);
  };

  return (
    <Dialog>
      <DialogTrigger asChild>{props.children}</DialogTrigger>

      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>New folder</DialogTitle>
          <DialogDescription></DialogDescription>
        </DialogHeader>

        <div className="flex items-center space-x-2">
          <div className="grid flex-1 gap-2">
            <Label htmlFor="link" className="sr-only">
              Name
            </Label>

            <Input
              placeholder="Enter the folder name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
        </div>

        <DialogFooter className="sm:justify-end">
          <DialogClose asChild>
            <Button type="button" variant="default" onClick={handleCreate}>
              Create
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
