import { AsyncPromptModalProvider } from "@/components/async-prompt-modal";
import { AppErrorBoundary } from "@/components/error-boundary";
import { ThemeProvider } from "@/components/theme-provider";
import React, { PropsWithChildren } from "react";

export function Providers(props: PropsWithChildren) {
  return (
    <React.StrictMode>
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <AppErrorBoundary>
          <AsyncPromptModalProvider>{props.children}</AsyncPromptModalProvider>
        </AppErrorBoundary>
      </ThemeProvider>
    </React.StrictMode>
  );
}
