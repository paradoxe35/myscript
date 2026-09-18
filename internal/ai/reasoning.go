// Copyright (c) 2024
// Licensed under the MIT License. See LICENSE file in the root directory.

package ai

import (
	"errors"
	"sync"
)

type reasoningStyle int

const (
	reasoningOpenAIEffort reasoningStyle = iota
	reasoningOpenRouter
)

// OpenRouter is the one OpenAI-compatible gateway with its own reasoning shape,
// and a custom provider can point at it too, so the host decides.
func reasoningStyleOf(baseURL string) reasoningStyle {
	if containsFold(baseURL, "openrouter.ai") {
		return reasoningOpenRouter
	}
	return reasoningOpenAIEffort
}

// refusedReasoning remembers endpoint/model pairs that rejected the parameter,
// so the wasted round trip happens once per launch instead of every request.
var refusedReasoning sync.Map

func withReasoningFallback(settings Settings, send func(lowReasoning bool) error) error {
	key := settings.BaseURL + "::" + settings.Model

	if _, refused := refusedReasoning.Load(key); !settings.LowReasoning || refused {
		return send(false)
	}

	err := send(true)
	if err == nil || !rejectedParameter(err) {
		return err
	}

	if retryErr := send(false); retryErr != nil {
		return retryErr
	}

	refusedReasoning.Store(key, struct{}{})
	return nil
}

func rejectedParameter(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == 400 || apiErr.StatusCode == 422
}
