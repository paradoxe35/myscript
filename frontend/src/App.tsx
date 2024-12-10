import { ThemeProvider } from "./components/theme-provider";

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      Ca va ?
    </ThemeProvider>
  );
}

export default App;
