import { createRoot } from "react-dom/client";
import App from "./app";
import { Toaster as SonnerToaster } from "./components/ui/sonner";
import { Toaster } from "./components/ui/toaster";

import "./styles/style.css";
import "./styles/prosemirror.css";
import "highlight.js/styles/github-dark.css";
import { Providers } from "./app/providers";

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <Providers>
    <SonnerToaster />
    <Toaster />
    <App />
  </Providers>
);
