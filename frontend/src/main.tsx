import React from "react";
import { createRoot } from "react-dom/client";
import { ThemeProvider } from "./components/theme-provider";
import App from "./app";
import { Toaster } from "./components/ui/sonner";

import { AsyncPromptModalProvider } from "./components/async-prompt-modal";
import { AppErrorBoundary } from "./components/error-boundary";

import "./styles/style.css";
import "./styles/prosemirror.css";
import "highlight.js/styles/github-dark.css";

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AppErrorBoundary>
        <AsyncPromptModalProvider>
          <App />
          <Toaster />
        </AsyncPromptModalProvider>
      </AppErrorBoundary>
    </ThemeProvider>
  </React.StrictMode>
);
