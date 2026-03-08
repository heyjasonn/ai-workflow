package orchestrator

import "github.com/heyjasonn/ai-workflow/internal/llm"

func NewRunner(baseDir string, provider ProviderAdapter) Runner {
	return Runner{
		PromptBuilder: DefaultPromptBuilder{},
		Provider:      provider,
		Validator:     JSONContractValidator{},
		Artifacts:     FileArtifactStore{BaseDir: baseDir},
		MaxRetries:    1,
	}
}

func NewLLMProvider(client llm.LLM) ProviderAdapter {
	return LLMProviderAdapter{LLM: client}
}
