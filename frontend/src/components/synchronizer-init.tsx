import {
  isGoogleAPIInvalidGrantError,
  useGoogleAuthTokenStore,
} from "@/store/google-auth-token";
import { refreshAfterSync } from "@/store/sync-refresh";
import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { EventsOn } from "~wails-runtime";
import { StopSynchronizer } from "~wails/main/App";

export function SynchronizerInit() {
  const googleAuthTokenStore = useGoogleAuthTokenStore();

  const syncFailures = useRef(0);

  useEffect(() => {
    return EventsOn("on-sync-failure", async (error) => {
      if (isGoogleAPIInvalidGrantError(error)) {
        syncFailures.current += 1;

        if (syncFailures.current === 4) {
          console.log("Too many sync attempts, refreshing Google auth token");

          await StopSynchronizer().catch(console.error);

          googleAuthTokenStore.refreshToken().catch(() => {
            googleAuthTokenStore.deleteToken();
            toast.warning("Google auth token expired, please re-authorize");
            console.log("Failed to refresh Google auth token, deleting it");
          });
        }
      }

      console.log(
        "Is Invalid Grant Error:",
        isGoogleAPIInvalidGrantError(error)
      );
      console.error("Synchronization failed:", error);
    });
  }, []);

  useEffect(() => EventsOn("on-sync-success", refreshAfterSync), []);

  return <></>;
}
