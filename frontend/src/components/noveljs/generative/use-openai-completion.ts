import OpenAI from "openai";
import { type Option } from "./ai-selector-commands";
import { useState } from "react";

interface Messages {
  role: "system" | "user";
  content: string;
}

const OPENAI_MODEL = "gpt-4o-mini";

const generateMessage = (
  option: Option,
  prompt: string,
  command?: string
): Messages[] => {
  const systemMessages = {
    continue:
      "You are an AI writing assistant that continues existing text based on context from prior text. " +
      "Give more weight/priority to the later characters than the beginning ones. " +
      "Limit your response to no more than 200 characters, but make sure to construct complete sentences." +
      "Use Markdown formatting when appropriate.",

    improve:
      "You are an AI writing assistant that improves existing text. " +
      "Limit your response to no more than 200 characters, but make sure to construct complete sentences." +
      "Use Markdown formatting when appropriate.",

    shorter:
      "You are an AI writing assistant that shortens existing text. " +
      "Use Markdown formatting when appropriate.",

    longer:
      "You are an AI writing assistant that lengthens existing text. " +
      "Use Markdown formatting when appropriate.",

    fix:
      "You are an AI writing assistant that fixes grammar and spelling errors in existing text. " +
      "Limit your response to no more than 200 characters, but make sure to construct complete sentences." +
      "Use Markdown formatting when appropriate.",

    zap:
      "You area an AI writing assistant that generates text based on a prompt. " +
      "You take an input from the user and a command for manipulating the text" +
      "Use Markdown formatting when appropriate.",
  };

  return [
    {
      role: "system",
      content: systemMessages[option],
    },
    {
      role: "user",
      content:
        option === "zap"
          ? `For this text: ${prompt}. You have to respect the command: ${command}`
          : `The existing text is: ${prompt}`,
    },
  ];
};

export const useOpenAICompletion = (apiKey: string) => {
  const [completion, setCompletion] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const generateCompletion = async (
    prompt: string,
    option: Option,
    command?: string
  ) => {
    if (!apiKey) {
      throw new Error("OpenAI API key is not set");
    }

    setCompletion("");
    setIsLoading(true);
    setError(null);

    const openai = new OpenAI({
      apiKey: apiKey,
      // Since it's a desktop app, we can use the browser's fetch API
      dangerouslyAllowBrowser: true,
    });

    const messages = generateMessage(option, prompt, command);

    try {
      const response = await openai.chat.completions.create({
        model: OPENAI_MODEL,
        messages,
        top_p: 1,
        frequency_penalty: 0,
        presence_penalty: 0,
        n: 1,
        temperature: 0.7,
        stream: true,
      });

      for await (const chunk of response) {
        const content = chunk.choices[0]?.delta?.content || "";
        setCompletion((prev) => prev + content);
      }
    } catch (err) {
      const error = err instanceof Error ? err : new Error("An error occurred");
      setError(error);
    } finally {
      setIsLoading(false);
    }
  };

  return {
    generateCompletion,
    completion,
    isLoading,
    error,
  };
};
