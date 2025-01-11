import {
  createContext,
  PropsWithChildren,
  useCallback,
  useContext,
  useRef,
  useState,
} from "react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { cn } from "@/lib/utils";

type AsyncPromptModalParams = {
  title?: string;
  description?: string;
  placeholder?: string;
  confirmationPrompt?: boolean;
};

type PromptFn = AsyncPromptModalContext["prompt"];

type AsyncPromptModalContext = {
  prompt<T extends AsyncPromptModalParams = AsyncPromptModalParams>(
    params?: T
  ): Promise<
    T extends { confirmationPrompt: true } ? boolean | null : string | null
  >;
};

const AsyncPromptModalContext = createContext<AsyncPromptModalContext>({
  prompt: () => Promise.resolve(null),
});

export const useAsyncPromptModal = () => {
  return useContext(AsyncPromptModalContext);
};

useAsyncPromptModal.prompt = (async (params: AsyncPromptModalParams) =>
  null) as PromptFn;

export const AsyncPromptModalProvider = ({ children }: PropsWithChildren) => {
  const promptRef = useRef<PromptFn | null>(null);

  const prompt = useCallback(async (params: AsyncPromptModalParams) => {
    if (!promptRef.current) {
      return null;
    }

    return promptRef.current(params);
  }, []) as PromptFn;

  useAsyncPromptModal.prompt = prompt;

  return (
    <AsyncPromptModalContext.Provider value={{ prompt }}>
      {children}
      <AsyncPromptModal promptRef={promptRef} />
    </AsyncPromptModalContext.Provider>
  );
};

function AsyncPromptModal({
  promptRef,
}: PropsWithChildren<{
  promptRef: React.RefObject<PromptFn | null>;
}>) {
  const [value, setValue] = useState<string>("");
  const [open, setOpen] = useState(false);
  const [option, setOption] = useState<AsyncPromptModalParams | undefined>(
    undefined
  );

  const resolver = useRef<((v: string | boolean | null) => void) | null>(null);

  promptRef.current = async function (params: AsyncPromptModalParams) {
    setValue("");
    setOpen(true);
    setOption(params);

    return new Promise((resolve) => {
      resolver.current = resolve as any;
    });
  } as PromptFn;

  const onEnterClick = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      resolver.current?.(option?.confirmationPrompt || value);
      resolver.current = null;

      setOpen(false);
    },
    [value, option]
  );

  return (
    <Dialog
      open={open}
      onOpenChange={(open) => {
        setOpen(open);

        setTimeout(() => {
          resolver.current?.(null);
        }, 100);
      }}
    >
      <DialogContent
        className={cn(
          "sm:max-w-[425px]",
          option?.confirmationPrompt && "sm:max-w-[400px]"
        )}
        showCloseButton={false}
      >
        <DialogHeader>
          <DialogTitle>{option?.title}</DialogTitle>
          <DialogDescription>{option?.description}</DialogDescription>
        </DialogHeader>

        <form className="w-full" onSubmit={onEnterClick}>
          {!option?.confirmationPrompt && (
            <Input
              type="text"
              value={value}
              autoFocus
              onChange={(e) => setValue(e.target.value)}
              placeholder={option?.placeholder}
            />
          )}

          <DialogFooter className="justify-between sm:justify-between mt-2">
            <DialogClose asChild>
              <Button type="button" variant="outline">
                Cancel
              </Button>
            </DialogClose>

            <DialogClose asChild>
              <Button type="submit">OK</Button>
            </DialogClose>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
