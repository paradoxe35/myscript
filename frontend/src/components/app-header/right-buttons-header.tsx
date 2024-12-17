import { useActivePageStore } from "@/store/active-page";
import { ScriptReaderControllers } from "./script-reader-controllers";
import { Separator } from "../ui/separator";
import { ZoomController } from "./zoom-controller";

export function RightButtonsHeader() {
  const activePage = useActivePageStore((state) => state.page);

  return (
    <div className="ml-auto flex items-center gap-2">
      {activePage && (
        <>
          <ScriptReaderControllers />
          <Separator orientation="vertical" className="mr-2 h-4" />
        </>
      )}

      <ZoomController />
    </div>
  );
}
