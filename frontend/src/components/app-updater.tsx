import { useEffect } from "react";

export function AppUpdater() {
  useEffect(() => {
    const intervalID = setInterval(() => {
      console.log("Checking for updates...");
    }, 1000 * 60 * 5);

    return () => clearInterval(intervalID);
  }, []);

  return <></>;
}
