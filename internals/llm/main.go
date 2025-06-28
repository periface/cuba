package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	customtools "github.com/periface/cuba/internals/custom_tools"
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

type ProveedorReviewer struct {
	llm      llms.Model
	executor *agents.Executor
}

func NewProveedorReviewer() (*ProveedorReviewer, error) {
	deepSeekApiKey, err := utils.GetEnvVariable("DEEPSEEK_API_KEY")
	if err != nil {
		return nil, err
	}
	//serpapiAPiKey, err := utils.GetEnvVariable("SERPAPI_API_KEY")
	//if err != nil {
	//	return nil, err
	//}
	// llm, err := openai.New(
	//
	//	openai.WithToken(deepSeekApiKey),
	//	openai.WithModel("deepseek-chat"),
	//	openai.WithBaseURL("https://api.deepseek.com/v1"),
	//
	// )
	llm, err := openai.New(
		openai.WithToken(deepSeekApiKey),
		openai.WithModel("gemini-2.0-flash"),
		openai.WithBaseURL("https://generativelanguage.googleapis.com/v1beta"),
	)
	//searchTool, err := serpapi.New(
	//	serpapi.WithAPIKey(serpapiAPiKey),
	//)

	googleSearchTool := customtools.NewGoogleSearchTool()

	if err != nil {
		return nil, err
	}

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
	return &ProveedorReviewer{
		llm:      llm,
		executor: executor,
	}, nil

}
func listToStr(fallBackText string, input []map[string]string) string {
	if len(input) == 0 {
		return fallBackText
	}

	var sb strings.Builder
	for i, item := range input {
		sb.WriteString(fmt.Sprintf("Observación %d:\n", i+1))
		for key, value := range item {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", key, value))
		}
	}
	return sb.String()
}
func (pr *ProveedorReviewer) ReviewProveedor(rfc string,
	buscarResponse models.BuscarResponse) (models.LLMResponse, error) {
	// Formatear observaciones
	observacionesStr := listToStr("No hay observaciones en el SAT", buscarResponse.ObservacionesSat)
	contratosStr := listToStr("Sin datos de contratos", buscarResponse.ContratosEncontrados)
	empleadosEncontradosStr := listToStr("Sin datos de empleados como proveedores", buscarResponse.EmpleadosEncontrados)
	informacionDelProveedorStr := listToStr("DGCyOP no tiene datos del proveedor", buscarResponse.InformacionDelProveedor)
	representantesLegalesStr := listToStr("No se encontraron representantes legales", buscarResponse.RepresentantesLegales)
	// Crear prompt claro para el agente
	prompt := fmt.Sprintf(`Revisión de proveedor para la Secretaría de Administración del Estado de Tamaulipas:

Datos del proveedor:
- RFC: %s
- Información que tiene la DIRECCION GENERAL DE COMPRAS Y OPERACIONES PATRIMONIALES (DGCyOP) del proveedor
%s
- Observaciones SAT:
%s
- Contratos en gobierno
%s
- Empleados con el mismo RFC
%s
- Representantes Legales en DGCyOP
%s

Instrucciones específicas:"
1. Busca en internet usando el RFC "%s" y la razón social relacionada
2. Analiza las observaciones del SAT
3. Genera un reporte con:
   a) Actividades que realiza la empreza/persona
   b) Resumen de observaciones del SAT (máximo 3 líneas)
   c) Información reelevante sobre los contratos en los que participa (Si hay)
   d) Hallazgos relevantes en la web (mencionar fuentes si hay),
		puedes usar el nombre de los representantes legales para buscar informacion sobre ellos
   e) Recomendación clara (aprobado/observaciones/rechazado)
   f) Si hay urls relevantes agregar links
4. Formato: Español claro, máximo 2 párrafos
5. Si hay imagenes o fotos relevantes (logos, notas, etc.) Agregalas
6. Si no encuentras información relevante, indica "No se encontraron referencias públicas
7. Si no hay datos sobre algo ignoralo
8. Da prioridad a notas periodisticas o temas que puedan prevenir posibles polemicas
9. Si solo hay datos de empleados de gobierno"`,

		rfc,
		informacionDelProveedorStr,
		observacionesStr,
		contratosStr,
		empleadosEncontradosStr,
		representantesLegalesStr,
		rfc)

	// Ejecutar con el agente (que usará la búsqueda web cuando sea necesario)
	response, err := chains.Run(context.Background(), pr.executor, prompt)
	slog.Debug("================")
	slog.Debug(response)
	slog.Debug("================")
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
