package ai

import (
	"context"
	"fmt"

	"github.com/VedRatan/remediation-server/ai/gemini"
	"github.com/VedRatan/remediation-server/types"
)

type AIClient interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

func GetAiClient(ai string) (AIClient, error) {
	switch ai {
	case "gemini":
		return gemini.NewGeminiClient(types.AiAgentKey), nil
	default:
		return nil, fmt.Errorf("specified ai backend is not supported yet: %v", ai)
	}
}
