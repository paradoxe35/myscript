// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package utils

import (
	"context"
	"os"
	"os/signal"
)

// ContextForSignal returns a context object which is cancelled when a signal
// is received. It returns nil if no signal parameter is provided
func ContextForSignal(signals ...os.Signal) context.Context {
	if len(signals) == 0 {
		return nil
	}

	ch := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())

	// Send message on channel when signal received
	signal.Notify(ch, signals...)

	// When any signal received, call cancel
	go func() {
		<-ch
		cancel()
	}()

	// Return success
	return ctx
}
