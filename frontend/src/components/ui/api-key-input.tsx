"use client";

import * as React from "react";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

export interface ApiKeyInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "type"> {}

const ApiKeyInput = React.forwardRef<HTMLInputElement, ApiKeyInputProps>(
  ({ className, ...props }, ref) => {
    const [isVisible, setIsVisible] = React.useState(false);

    const toggleVisibility = () => {
      setIsVisible(!isVisible);
    };

    return (
      <div className="relative">
        <input
          type={isVisible ? "text" : "password"}
          className={cn(
            "flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
            className
          )}
          ref={ref}
          {...props}
        />
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="absolute right-0 top-0 h-full px-3 py-1 hover:bg-transparent"
          onClick={toggleVisibility}
          tabIndex={-1}
        >
          {isVisible ? (
            <EyeOffIcon className="h-4 w-4 text-muted-foreground" />
          ) : (
            <EyeIcon className="h-4 w-4 text-muted-foreground" />
          )}
          <span className="sr-only">
            {isVisible ? "Hide API key" : "Show API key"}
          </span>
        </Button>
      </div>
    );
  }
);
ApiKeyInput.displayName = "ApiKeyInput";

export { ApiKeyInput };
