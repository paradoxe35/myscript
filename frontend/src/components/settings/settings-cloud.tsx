import { Loader2 } from "lucide-react";
import { Button } from "../ui/button";
import { useSettings } from "./context";

export function SettingsCloud() {
  const { cloud } = useSettings();

  return (
    <div className="mt-5">
      <p className="text-xs text-white/50">
        Store your data in your Google Drive account and sync it seamlessly
        across all your devices.
      </p>

      {cloud.authorizing && (
        <Button className="mt-5" disabled>
          <Loader2 className="animate-spin" />
          Connecting
        </Button>
      )}

      {!cloud.authorizing && (
        <Button className="mt-5" onClick={cloud.startAuthorization}>
          <GoogleDriveIcon />
          Connect to Google Drive
        </Button>
      )}
    </div>
  );
}

function AuthorizationTimeout() {}

function GoogleDriveIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 32 32"
      width="64px"
      height="64px"
      {...props}
    >
      <path d="M 11.4375 5 L 11.15625 5.46875 L 3.15625 18.46875 L 2.84375 18.96875 L 3.125 19.5 L 7.125 26.5 L 7.40625 27 L 24.59375 27 L 24.875 26.5 L 28.875 19.5 L 29.15625 18.96875 L 28.84375 18.46875 L 20.84375 5.46875 L 20.5625 5 Z M 13.78125 7 L 19.4375 7 L 26.21875 18 L 20.5625 18 Z M 12 7.90625 L 14.96875 12.75 L 8.03125 24.03125 L 5.15625 19 Z M 16.15625 14.65625 L 18.21875 18 L 14.09375 18 Z M 12.875 20 L 26.28125 20 L 23.40625 25 L 9.78125 25 Z" />
    </svg>
  );
}
