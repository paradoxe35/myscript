// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"strings"
	"testing"
)

func TestBuildRequestCarriesTheText(t *testing.T) {
	request, err := BuildRequest(TaskImprove, "some prose", "")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(request.Prompt, "some prose") {
		t.Errorf("prompt = %q", request.Prompt)
	}
	if !strings.Contains(request.System, "improves") {
		t.Errorf("system = %q", request.System)
	}
}

func TestBuildRequestIncludesTheCommandForAdHocTasks(t *testing.T) {
	request, err := BuildRequest(TaskCommand, "some prose", "make it rhyme")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(request.Prompt, "make it rhyme") {
		t.Errorf("prompt = %q", request.Prompt)
	}
}

func TestBuildRequestRejectsUnknownTasks(t *testing.T) {
	if _, err := BuildRequest("sing", "text", ""); err == nil {
		t.Fatal("expected an error")
	}
}

func TestWriteTaskUsesTheInstructionAlone(t *testing.T) {
	request, err := BuildRequest(TaskWrite, "", "a haiku about rain")
	if err != nil {
		t.Fatal(err)
	}

	if request.Prompt != "a haiku about rain" {
		t.Errorf("prompt = %q", request.Prompt)
	}
}

func TestWriteTaskCarriesThePrecedingText(t *testing.T) {
	request, err := BuildRequest(TaskWrite, "The meeting opened at nine.", "summarise the rest")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(request.Prompt, "The meeting opened at nine.") {
		t.Errorf("prompt = %q", request.Prompt)
	}
	if !strings.Contains(request.Prompt, "summarise the rest") {
		t.Errorf("prompt = %q", request.Prompt)
	}
}
