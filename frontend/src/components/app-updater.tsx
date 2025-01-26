import { useEffect, useState } from "react";
import { toast } from "sonner";
import { CheckForUpdates, PerformUpdate, IsDevMode } from "~wails/main/App";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { RefreshCcw } from "lucide-react";

export function AppUpdater() {
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateProgress, setUpdateProgress] = useState(0);
  const [updateMessage, setUpdateMessage] = useState("");

  const handleUpdate = async () => {
    setIsUpdating(true);
    setUpdateMessage("Starting update process...");

    try {
      setUpdateMessage("Downloading update...");
      setUpdateProgress(30);

      // Perform the actual update
      await PerformUpdate();

      setUpdateMessage("Finalizing update...");
      setUpdateProgress(90);

      // Simulate final steps
      await new Promise((resolve) => setTimeout(resolve, 2000));

      toast.success("Update completed successfully! Restarting...");
      setTimeout(() => window.location.reload(), 1000);
    } catch (error) {
      toast.error("Update failed. Please try again.");
      console.error("Update error:", error);
    } finally {
      setIsUpdating(false);
      setUpdateProgress(0);
    }
  };

  useEffect(() => {
    const checkForUpdates = async () => {
      try {
        const response = await CheckForUpdates();

        if (response) {
          toast.custom(
            (t) => (
              <div className="flex items-center gap-4 p-4 bg-background border rounded-lg shadow-lg">
                <div>
                  <h3 className="font-semibold">New Update Available!</h3>
                  <p className="text-sm text-muted-foreground">{response}</p>
                </div>
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    onClick={() => toast.dismiss(t)}
                  >
                    Remind later
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => {
                      toast.dismiss(t);
                      handleUpdate();
                    }}
                  >
                    Update now
                  </Button>
                </div>
              </div>
            ),
            {
              id: "update-available",
              duration: Infinity,
              position: "top-right",
            }
          );
        }
      } catch (error) {
        console.error("Update check failed:", error);
      }
    };

    let intervalID: NodeJS.Timeout | undefined;

    (async () => {
      //   const isDevMode = await IsDevMode();
      //   if (isDevMode) return;

      checkForUpdates();
      intervalID = setInterval(checkForUpdates, 1000 * 60 * 3);
    })();

    return () => clearInterval(intervalID);
  }, []);

  return (
    <Dialog open={isUpdating} modal>
      <DialogContent className="sm:max-w-md" showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>Updating Application</DialogTitle>
          <DialogDescription>
            Please wait while we install the latest update.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col items-center gap-4 py-4">
          <RefreshCcw className="h-8 w-8 animate-spin" />
          <p className="text-sm font-medium">{updateMessage}</p>
          <Progress value={updateProgress} className="w-full" />
          <p className="text-xs text-muted-foreground">
            Do not close the application during this process
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}
