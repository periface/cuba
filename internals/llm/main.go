package llm

import (
	"context"

	customTools "github.com/periface/cuba/internals/llm/tools"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/utils"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"

	_ "github.com/tmc/langchaingo/tools/serpapi"
)

type PromptRunner struct {
	llm      llms.Model
	executor *agents.Executor
}

func getLlm() (*openai.LLM, error) {

	deepSeekApiKey, err := utils.GetEnvVariable("DEEPSEEK_API_KEY")
	if err != nil {
		return nil, err
	}
	return openai.New(
		openai.WithToken(deepSeekApiKey),
		openai.WithModel("gemini-2.0-flash"),
		openai.WithBaseURL("https://generativelanguage.googleapis.com/v1beta"),
	)
}

type LlmType = int

const (
	WithGoogleSearchTool = iota
)

var Llms = map[LlmType]string{
	WithGoogleSearchTool: "googleSearch",
}

func WithGoogleSearchExecutor(llm *openai.LLM) *agents.Executor {

	googleSearchTool := customTools.NewGoogleSearchTool()

	agentTools := []tools.Tool{WithFallback(googleSearchTool)}

	agent := agents.NewOneShotAgent(llm,
		agentTools,
		agents.WithMaxIterations(3),
		agents.WithCallbacksHandler(callbacks.LogHandler{}),
	)
	executor := agents.NewExecutor(
		agent,
		agents.WithCallbacksHandler(callbacks.LogHandler{}),
	)
	return executor
}

func NewPromptRunner(llmType LlmType) (*PromptRunner, error) {

	llm, err := getLlm()

	if err != nil {
		return nil, err
	}
	var executor *agents.Executor
	switch llmType {
	case WithGoogleSearchTool:
		return &PromptRunner{
			llm:      llm,
			executor: WithGoogleSearchExecutor(llm),
		}, nil
	}
	// set new tools
	return &PromptRunner{
		llm:      llm,
		executor: executor,
	}, nil

}

func (pr *PromptRunner) RunPrompt(prompt string) (models.LLMResponse, error) {
	// Formatear observaciones
	// Ejecutar con el agente (que usará la búsqueda web cuando sea necesario)
	response, err := chains.Run(context.Background(), pr.executor, prompt)
	if err != nil {
		// Si el error contiene la respuesta útil (común con algunos LLMs)
		if err.Error() != "" && len(err.Error()) > 50 { // 50 es arbitrario
			return models.LLMResponse{}, err
		}

		return models.LLMResponse{}, err
	}
	return models.LLMResponse{
		Prompt:   prompt,
		Response: response,
	}, nil
}

type fallbackTool struct {
	tool tools.Tool
}

func (f *fallbackTool) Name() string {
	return f.tool.Name()
}

func (f *fallbackTool) Description() string {
	return f.tool.Description()
}
func (f *fallbackTool) Call(ctx context.Context, input string) (string, error) {
	result, err := f.tool.Call(ctx, input)
	if err != nil {
		// Devuelve una cadena vacía y nil para indicar que no hay resultado pero no es un error fatal
		return "No se pudo obtener información de búsqueda", nil
	}
	return result, nil
}

func WithFallback(t tools.Tool) tools.Tool {
	return &fallbackTool{tool: t}
}
