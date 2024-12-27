import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import { ThemeProvider } from "./components/theme-provider";
import App from "./app";
import { Toaster } from "./components/ui/sonner";
import { TranscriberInit } from "./components/transcriber-init";
import { LocalWhisperInit } from "./components/local-whisper-init";

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <App />
      <Toaster />

      <TranscriberInit />
      <LocalWhisperInit />
    </ThemeProvider>
  </React.StrictMode>
);
