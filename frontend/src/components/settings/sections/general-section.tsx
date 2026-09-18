import { Button } from "@/components/ui/button";
import { MoonIcon, SunIcon } from "lucide-react";
import { useTheme } from "../../theme-provider";
import { useSettings } from "../context";
import { Field, SettingsGroup, SettingsPanel } from "../fields";

export function GeneralSection() {
  const { appVersion } = useSettings();
  const { theme, setTheme } = useTheme();

  return (
    <SettingsPanel>
      <SettingsGroup title="Appearance">
        <Field label="Theme">
          <div className="flex gap-2">
            <ThemeButton
              active={theme === "light"}
              onClick={() => setTheme("light")}
              icon={<SunIcon className="h-4 w-4" />}
              label="Light"
            />
            <ThemeButton
              active={theme === "dark"}
              onClick={() => setTheme("dark")}
              icon={<MoonIcon className="h-4 w-4" />}
              label="Dark"
            />
            <ThemeButton
              active={theme === "system"}
              onClick={() => setTheme("system")}
              label="System"
            />
          </div>
        </Field>
      </SettingsGroup>

      <SettingsGroup title="About">
        <p className="text-xs text-muted-foreground">
          MyScript {appVersion || "—"}
        </p>
      </SettingsGroup>
    </SettingsPanel>
  );
}

type ThemeButtonProps = {
  active: boolean;
  onClick: () => void;
  label: string;
  icon?: React.ReactNode;
};

function ThemeButton({ active, onClick, label, icon }: ThemeButtonProps) {
  return (
    <Button
      variant={active ? "default" : "outline"}
      size="sm"
      onClick={onClick}
      className="gap-2"
    >
      {icon}
      {label}
    </Button>
  );
}
