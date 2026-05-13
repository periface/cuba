package llm

import (
	"context"
	"log/slog"

	customTools "github.com/periface/cuba/internals/llm/tools"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/utils"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type PromptRunner struct {
	llm llms.Model
}

func getLlm() (*openai.LLM, error) {

	deepSeekApiKey, err := utils.GetEnvVariable("DEEPSEEK_API_KEY")
	if err != nil {
		return nil, err
	}

	model, err := utils.GetEnvVariable("MODEL")
	if err != nil {
		model = "gemini-2.0-flash"
	}

	baseUrl, err := utils.GetEnvVariable("BASEURL")
	if err != nil {
		baseUrl = "https://generativelanguage.googleapis.com/v1beta"
	}

	return openai.New(
		openai.WithToken(deepSeekApiKey),
		openai.WithModel(model),
		openai.WithBaseURL(baseUrl),
	)
}

type LlmType = int

const (
	WithGoogleSearchTool = iota
)

func NewPromptRunner(_ LlmType) (*PromptRunner, error) {

	llm, err := getLlm()
	if err != nil {
		return nil, err
	}

	return &PromptRunner{
		llm: llm,
	}, nil
}

func (pr *PromptRunner) RunPrompt(prompt string) (models.LLMResponse, error) {

	response, err := llms.GenerateFromSinglePrompt(
		context.Background(),
		pr.llm,
		prompt,
	)

	if err != nil {
		slog.Error(
			"error generating llm response",
			"error", err.Error(),
		)

		return models.LLMResponse{}, err
	}

	slog.Info(
		"llm response generated",
		"response_length", len(response),
	)

	return models.LLMResponse{
		Prompt:   prompt,
		Response: response,
	}, nil
}

func (pr *PromptRunner) RunPromptWithGoogle(
	basePrompt string,
	searchQueries []string,
) (models.LLMResponse, error) {

	googleTool, err := customTools.NewSerpAPISearchTool()
	if err != nil {
		return models.LLMResponse{}, err
	}

	var googleResults string

	for _, query := range searchQueries {

		if query == "" {
			continue
		}

		slog.Info(
			"running google search",
			"query", query,
		)

		result, err := googleTool.Call(
			context.Background(),
			query,
		)

		if err != nil {

			slog.Error(
				"google search failed",
				"query", query,
				"error", err.Error(),
			)

			continue
		}

		googleResults += "\n\n=============================\n"
		googleResults += "GOOGLE SEARCH QUERY:\n"
		googleResults += query
		googleResults += "\n\nRESULTS:\n"
		googleResults += result
		googleResults += "\n=============================\n"
	}

	finalPrompt := basePrompt

	if googleResults != "" {
		finalPrompt += "\n\n"
		finalPrompt += "### RESULTADOS DE INVESTIGACIÓN WEB\n"
		finalPrompt += googleResults
	}

	response, err := llms.GenerateFromSinglePrompt(
		context.Background(),
		pr.llm,
		finalPrompt,
	)

	if err != nil {

		slog.Error(
			"error generating llm response",
			"error", err.Error(),
		)

		return models.LLMResponse{}, err
	}

	slog.Info(
		"llm response generated",
		"response_length", len(response),
	)

	return models.LLMResponse{
		Prompt:   finalPrompt,
		Response: response,
	}, nil
}
