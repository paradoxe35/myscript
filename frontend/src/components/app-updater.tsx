import { useEffect } from "react";
import { toast } from "sonner";
import { CheckForUpdates, IsDevMode } from "~wails/main/App";

export function AppUpdater() {
  useEffect(() => {
    toast.info(
      "Checking for updates...",

      {
        id: "update-available",
        duration: 6000,
        position: "top-right",
        closeButton: false,
      }
    );

    let intervalID: NodeJS.Timeout | undefined = undefined;

    const checkForUpdates = async () => {
      console.log("Checking for updates...");
      const response = await CheckForUpdates();
      console.log(response);
    };

    (async () => {
      const isDevMode = await IsDevMode();
      if (!isDevMode) {
        return;
      }

      checkForUpdates();
      // Check for updates every 3 minutes
      intervalID = setInterval(checkForUpdates, 1000 * 60 * 3);
    })();

    return () => clearInterval(intervalID);
  }, []);

  return <></>;
}
