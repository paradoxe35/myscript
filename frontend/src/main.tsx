import React from "react";
import { createRoot } from "react-dom/client";
import { ThemeProvider } from "./components/theme-provider";
import App from "./app";
import { Toaster } from "./components/ui/sonner";

import "./styles/style.css";
import "./styles/prosemirror.css";
import { AsyncPromptModalProvider } from "./components/async-prompt-modal";

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <AsyncPromptModalProvider>
        <App />
        <Toaster />
      </AsyncPromptModalProvider>
    </ThemeProvider>
  </React.StrictMode>
);
