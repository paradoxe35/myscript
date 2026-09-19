// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import "fmt"

type Task string

const (
	TaskContinue Task = "continue"
	TaskImprove  Task = "improve"
	TaskShorter  Task = "shorter"
	TaskLonger   Task = "longer"
	TaskFix      Task = "fix"
	TaskCommand  Task = "zap"
	TaskWrite    Task = "write"
)

const markdownNote = "Use Markdown formatting when appropriate."

const briefNote = "Limit your response to no more than 200 characters, but make sure to construct complete sentences."

var taskInstructions = map[Task]string{
	TaskContinue: "You are a writing assistant that continues existing text based on the context before it. " +
		"Give more weight to the later text than the earlier text. " + briefNote + " " + markdownNote,

	TaskImprove: "You are a writing assistant that improves existing text. " + briefNote + " " + markdownNote,

	TaskShorter: "You are a writing assistant that shortens existing text. " + markdownNote,

	TaskLonger: "You are a writing assistant that lengthens existing text. " + markdownNote,

	TaskFix: "You are a writing assistant that fixes grammar and spelling mistakes in existing text. " +
		briefNote + " " + markdownNote,

	TaskCommand: "You are a writing assistant that rewrites text according to the user's instruction. " + markdownNote,

	TaskWrite: "You are a writing assistant. Write what the user asks for, continuing naturally from " +
		"any text they give you. Reply with the writing only, no preamble. " + markdownNote,
}

func BuildRequest(task Task, text, command string) (Request, error) {
	instruction, ok := taskInstructions[task]
	if !ok {
		return Request{}, fmt.Errorf("unknown writing task: %s", task)
	}

	var prompt string
	switch {
	case task == TaskWrite && text == "":
		prompt = command
	case task == TaskWrite:
		prompt = fmt.Sprintf("What comes before:\n%s\n\nWrite: %s", text, command)
	case task == TaskCommand:
		prompt = fmt.Sprintf("For this text: %s\n\nFollow this instruction: %s", text, command)
	default:
		prompt = fmt.Sprintf("The existing text is: %s", text)
	}

	return Request{System: instruction, Prompt: prompt}, nil
}
