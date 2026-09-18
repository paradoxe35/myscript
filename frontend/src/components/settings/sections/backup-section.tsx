import { GoogleAuth } from "../google-auth";
import { SettingsGroup, SettingsPanel } from "../fields";

export function BackupSection() {
  return (
    <SettingsPanel>
      <SettingsGroup
        title="Google Drive"
        description="Keep your scripts backed up and in sync across your devices."
      >
        <GoogleAuth />
      </SettingsGroup>
    </SettingsPanel>
  );
}
