import { ApiKeyInput } from "@/components/ui/api-key-input";
import { Button } from "@/components/ui/button";
import { SecretKey, useSecretsStore } from "@/store/secrets";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Field } from "../fields";

type SecretFieldProps = {
  secret: SecretKey;
  label: string;
  hint?: string;
  placeholder?: string;
  onSaved?: () => void;
};

export function SecretField({
  secret,
  label,
  hint,
  placeholder,
  onSaved,
}: SecretFieldProps) {
  const load = useSecretsStore((store) => store.load);
  const store = useSecretsStore((state) => state.save);

  const [value, setValue] = useState("");
  const [saved, setSaved] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    load(secret).then((stored) => {
      setValue(stored);
      setSaved(stored);
    });
  }, [secret]);

  const save = async () => {
    setSaving(true);
    try {
      await store(secret, value.trim());
      setSaved(value.trim());
      setValue(value.trim());
      onSaved?.();
      toast.success(`${label} saved`);
    } catch (error) {
      toast.error(String(error));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Field label={label} hint={hint}>
      <div className="flex gap-2">
        <ApiKeyInput
          value={value}
          placeholder={placeholder}
          onChange={(event) => setValue(event.target.value)}
        />
        <Button
          variant="outline"
          className="shrink-0"
          disabled={saving || value.trim() === saved}
          onClick={save}
        >
          {saving && <Loader2 className="h-4 w-4 animate-spin" />}
          Save
        </Button>
      </div>
    </Field>
  );
}
