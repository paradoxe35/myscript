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

type AsyncPromptModalParams = {
  title?: string;
  description?: string;
  placeholder?: string;
};

type PromptFn = AsyncPromptModalContext["prompt"];

type AsyncPromptModalContext = {
  prompt(params: AsyncPromptModalParams): Promise<string | null>;
};

const AsyncPromptModalContext = createContext<AsyncPromptModalContext>({
  prompt: () => Promise.resolve(""),
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
  }, []);

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
  const [value, setValue] = useState("");
  const [open, setOpen] = useState(false);
  const [option, setOption] = useState<AsyncPromptModalParams | null>(null);

  const resolver = useRef<((v: string | null) => void) | null>(null);

  promptRef.current = async (params: AsyncPromptModalParams) => {
    setValue("");
    setOpen(true);
    setOption(params);

    return new Promise<string | null>((resolve) => {
      resolver.current = resolve;
    });
  };

  const onEnterClick = useCallback(
    (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      resolver.current?.(value);
      resolver.current = null;

      setOpen(false);
    },
    [value]
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
      <DialogContent className="sm:max-w-[425px]" showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>{option?.title}</DialogTitle>
          <DialogDescription>{option?.description}</DialogDescription>
        </DialogHeader>

        <form className="w-full" onSubmit={onEnterClick}>
          <Input
            type="text"
            value={value}
            autoFocus
            onChange={(e) => setValue(e.target.value)}
            placeholder={option?.placeholder}
          />

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
