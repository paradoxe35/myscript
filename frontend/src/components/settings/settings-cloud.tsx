import { useSettings } from "./context";
import { GoogleAuth } from "./google-auth";

export function SettingsCloud() {
  const { cloud } = useSettings();

  return (
    <div className="mt-5 flex flex-col gap-3">
      {cloud.googleAuthEnabled && <GoogleAuth />}

      {/* <Separator /> */}
    </div>
  );
}
