import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useEffect, useState } from "react";
import { useLocalPagesStore } from "@/store/local-pages";
import { repository } from "~wails/models";

type Props = {
  page: repository.Page;
  onOpenChange?(open: boolean): void;
  open: boolean;
};

export function RenameFolderModal(props: Props) {
  const [name, setName] = useState("");
  const savePageTitle = useLocalPagesStore((state) => state.savePageTitle);

  const handleCreate = async () => {
    const folderName = name.trim();
    if (!folderName) {
      return;
    }

    savePageTitle(name, props.page);
  };

  useEffect(() => {
    setName(props.page.title);
  }, [props.page.title]);

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Rename folder</DialogTitle>
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
              Save
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
