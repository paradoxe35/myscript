import * as React from "react";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ApiKeyInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "type"> {}

const ApiKeyInput = React.forwardRef<HTMLInputElement, ApiKeyInputProps>(
  ({ className, disabled, ...props }, ref) => {
    const [visible, setVisible] = React.useState(false);

    return (
      <div className="relative w-full min-w-0">
        <input
          ref={ref}
          type={visible ? "text" : "password"}
          disabled={disabled}
          autoComplete="off"
          autoCorrect="off"
          autoCapitalize="off"
          spellCheck={false}
          className={cn(
            "h-9 w-full rounded-md border border-input bg-background py-1 pl-3 text-sm shadow-sm transition-colors",
            // Room for the toggle, so a long key never runs under it.
            "pr-10",
            "font-mono tracking-tight placeholder:font-sans placeholder:tracking-normal placeholder:text-muted-foreground",
            "focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
            "disabled:cursor-not-allowed disabled:opacity-50",
            className
          )}
          {...props}
        />

        <button
          type="button"
          disabled={disabled}
          onClick={() => setVisible((shown) => !shown)}
          aria-label={visible ? "Hide API key" : "Show API key"}
          aria-pressed={visible}
          className={cn(
            "absolute inset-y-0 right-0 flex w-10 items-center justify-center rounded-r-md",
            "text-muted-foreground transition-colors hover:text-foreground",
            "focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
            "disabled:cursor-not-allowed disabled:opacity-50"
          )}
        >
          {visible ? (
            <EyeOffIcon className="h-4 w-4" />
          ) : (
            <EyeIcon className="h-4 w-4" />
          )}
        </button>
      </div>
    );
  }
);
ApiKeyInput.displayName = "ApiKeyInput";

export { ApiKeyInput };
