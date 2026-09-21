// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package main

import (
	"strings"
	"testing"
)

// Wails never settles the promise after a panic in a binding, so bindings
// must return the failure instead.
func TestRecoverAsErrorReturnsThePanic(t *testing.T) {
	err := func() (err error) {
		defer recoverAsError(&err)
		var providers map[string]string
		providers["boom"] = "nil map write"
		return nil
	}()

	if err == nil {
		t.Fatal("the panic should surface as an error")
	}
	if !strings.Contains(err.Error(), "nil map") {
		t.Errorf("error should say what happened, got %q", err)
	}
}

func TestRecoverAsErrorLeavesASuccessAlone(t *testing.T) {
	err := func() (err error) {
		defer recoverAsError(&err)
		return nil
	}()

	if err != nil {
		t.Errorf("got %v", err)
	}
}
