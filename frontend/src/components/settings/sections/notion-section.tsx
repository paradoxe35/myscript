import { SettingsGroup, SettingsPanel } from "../fields";
import { SecretField } from "./secret-field";

export function NotionSection() {
  return (
    <SettingsPanel>
      <SettingsGroup
        title="Notion"
        description="Read your Notion pages and use them as scripts."
      >
        <SecretField
          secret="notion"
          label="Integration token"
          placeholder="ntn_..."
          hint="Create an internal integration in Notion, then share the pages you want with it."
        />
      </SettingsGroup>
    </SettingsPanel>
  );
}
