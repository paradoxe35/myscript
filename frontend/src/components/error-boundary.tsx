import { PropsWithChildren } from "react";

import { ErrorBoundary } from "react-error-boundary";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "./ui/button";
import { WindowReloadApp } from "~wails-runtime";

export function AppErrorBoundary(props: PropsWithChildren) {
  return <ErrorBoundary fallback={<Error />}>{props.children}</ErrorBoundary>;
}

function Error() {
  return (
    <div className="flex flex-col items-center justify-center h-screen">
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>An error has occurred.</CardTitle>
          <CardDescription>
            This might be a temporary issue. Please click the button below to
            reload the application and try your action again.
          </CardDescription>
        </CardHeader>

        <CardFooter>
          <Button onClick={() => WindowReloadApp()}>Reload</Button>
        </CardFooter>
      </Card>
    </div>
  );
}
