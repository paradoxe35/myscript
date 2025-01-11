import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const wait = (ms: number): Promise<void> => {
  return new Promise((resolve) => setTimeout(resolve, ms));
};

export function splitWithDelimiters(str: string, regexPattern: RegExp) {
  // Ensure the regex has the global flag
  const regex = new RegExp(
    regexPattern,
    regexPattern.flags.includes("g")
      ? regexPattern.flags
      : regexPattern.flags + "g"
  );

  const result = [];
  let lastIndex = 0;
  let match;

  while ((match = regex.exec(str)) !== null) {
    const chunk = str.slice(lastIndex, match.index);
    if (chunk) {
      result.push({
        text: chunk,
        delimiter: match[0],
        position: lastIndex,
      });
    }

    lastIndex = regex.lastIndex;
  }

  // Add the remaining chunk if there is any
  if (lastIndex < str.length) {
    result.push({
      text: str.slice(lastIndex),
      delimiter: null,
      position: lastIndex,
    });
  }

  return result;
}

export const reorder = <T>(list: T[], startIndex: number, endIndex: number) => {
  const result = Array.from(list);
  const [removed] = result.splice(startIndex, 1);
  result.splice(endIndex, 0, removed);

  return result;
};
