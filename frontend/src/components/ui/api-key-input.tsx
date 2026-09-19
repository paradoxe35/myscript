import * as React from "react";
import { EyeIcon, EyeOffIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

export interface ApiKeyInputProps extends Omit<
  React.InputHTMLAttributes<HTMLInputElement>,
  "type"
> {}

const ApiKeyInput = React.forwardRef<HTMLInputElement, ApiKeyInputProps>(
  ({ className, disabled, ...props }, ref) => {
    const [visible, setVisible] = React.useState(false);

    return (
      <div className="relative w-full min-w-0">
        <Input
          ref={ref}
          type={visible ? "text" : "password"}
          disabled={disabled}
          autoComplete="off"
          autoCorrect="off"
          autoCapitalize="off"
          spellCheck={false}
          className={cn(
            // Room for the toggle, so a long key never runs under it.
            "pr-10 font-mono tracking-tight",
            "placeholder:font-sans placeholder:tracking-normal",
            className,
          )}
          {...props}
        />

        <Button
          type="button"
          variant="ghost"
          size="icon"
          disabled={disabled}
          onClick={() => setVisible((shown) => !shown)}
          aria-label={visible ? "Hide API key" : "Show API key"}
          aria-pressed={visible}
          className="absolute inset-y-0 right-0 h-full w-10 text-muted-foreground hover:bg-transparent hover:text-foreground"
        >
          {visible ? (
            <EyeOffIcon className="h-4 w-4" />
          ) : (
            <EyeIcon className="h-4 w-4" />
          )}
        </Button>
      </div>
    );
  },
);
ApiKeyInput.displayName = "ApiKeyInput";

export { ApiKeyInput };
