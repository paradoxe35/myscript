import { useRef } from "react";

export const useSyncRef = <T>(value: T) => {
  const ref = useRef<T>(value);

  ref.current = value;

  return ref;
};
