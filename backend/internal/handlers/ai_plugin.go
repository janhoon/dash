package handlers

import (
	_ "github.com/aceobservability/ace-llm-anthropic"
	_ "github.com/aceobservability/ace-llm-copilot"
	_ "github.com/aceobservability/ace-llm-openai-compat"

	"github.com/aceobservability/ace/backend/pkg/llm"
)

type (
	AIProvider  = llm.AIProvider
	AIModel     = llm.AIModel
	ChatRequest = llm.ChatRequest
	LLMConfig   = llm.LLMConfig
	LLMFactory  = llm.LLMFactory
)

var ErrUnknownProviderType = llm.ErrUnknownProviderType
