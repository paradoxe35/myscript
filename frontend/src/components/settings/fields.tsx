import { cn } from "@/lib/utils";
import { PropsWithChildren } from "react";

export function SettingsPanel({ children }: PropsWithChildren) {
  return <div className="flex flex-col gap-6">{children}</div>;
}

type GroupProps = PropsWithChildren<{
  title: string;
  description?: string;
  action?: React.ReactNode;
}>;

export function SettingsGroup({
  title,
  description,
  action,
  children,
}: GroupProps) {
  return (
    <section className="flex flex-col gap-3">
      <header className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-0.5">
          <h3 className="text-sm font-medium leading-none">{title}</h3>
          {description && (
            <p className="text-xs text-muted-foreground">{description}</p>
          )}
        </div>
        {action}
      </header>

      {children}
    </section>
  );
}

type FieldProps = PropsWithChildren<{
  label: string;
  hint?: string;
  htmlFor?: string;
  className?: string;
}>;

export function Field({
  label,
  hint,
  htmlFor,
  className,
  children,
}: FieldProps) {
  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <label
        htmlFor={htmlFor}
        className="text-xs font-medium text-muted-foreground"
      >
        {label}
      </label>
      {children}
      {hint && <p className="text-xs text-muted-foreground/80">{hint}</p>}
    </div>
  );
}

export function SettingsCard({
  className,
  children,
}: PropsWithChildren<{ className?: string }>) {
  return (
    <div className={cn("rounded-lg border bg-card/50 p-4", className)}>
      {children}
    </div>
  );
}

export function Hint({ children }: PropsWithChildren) {
  return <p className="text-xs text-muted-foreground">{children}</p>;
}
