package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

type ProveedorReviewer struct {
	llm      llms.Model
	executor *agents.Executor
}

func NewProveedorReviewer() (*ProveedorReviewer, error) {
	deepSeekApiKey, err := utils.GetEnvVariable("DEEPSEEK_API_KEY")
	if err != nil {
		return nil, err
	}
	llm, err := openai.New(
		openai.WithToken(deepSeekApiKey),
		openai.WithModel("gemini-2.0-flash"),
		openai.WithBaseURL("https://generativelanguage.googleapis.com/v1beta"),
	)
	googleSearchTool := customTools.NewGoogleSearchTool()

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
	prompt := fmt.Sprintf(`Revisión integral de proveedor para la Secretaría de Administración del Estado de Tamaulipas: Herramienta Anticorrupción

**Objetivo del Reporte:** Proporcionar información crucial para la toma de decisiones directivas, identificando riesgos y posibles conflictos de interés de un proveedor para salvaguardar la integridad y la imagen del Gobierno Estatal de Tamaulipas.

**Datos del Proveedor a Analizar:**
- **RFC del Proveedor:** %s
- **Información DGCyOP:** %s
- **Observaciones SAT:** %s
- **Contratos en Gobierno:** %s
- **Empleados con Mismo RFC (en plantilla laboral de gobierno):** %s
- **Representantes Legales en DGCyOP:** %s

**Instrucciones para la Generación del Reporte (TODO EN ESPAÑOL):**

---

### 1. Recolección y Priorización de Información Externa (Web)

Realiza una búsqueda exhaustiva en internet siguiendo estas prioridades:
* **Prioridad 1: Razón Social del Proveedor.** Inicia con búsquedas amplias usando la razón social.
* **Prioridad 2: RFC del Proveedor y Nombres de Representantes Legales.** Complementa con el **RFC "%s"** y los nombres completos de los representantes legales.
* **Palabras Clave de Riesgo:** Incorpora proactivamente términos como "corrupción", "lavado de dinero", "evasión fiscal", "fraude", "irregularidades", "polémica", "escándalo", "investigación", "denuncia", "demandas", "sanciones", "conflicto de interés", "soborno", "inhabilitado", "nepotismo".
* **Fuentes Prioritarias:** Da máxima relevancia a **notas periodísticas de investigación, comunicados oficiales de fiscalías o contralorías, sentencias judiciales, reportes de auditoría, listas de proveedores inhabilitados, y documentos de dominio público** que vinculen al proveedor o sus representantes legales con actividades ilícitas, controversias o prácticas poco éticas. Incluye imágenes o fotos relevantes (logos, capturas de notas, etc.).

---

### 2. Análisis Integrado de Datos Internos y Externos

* **Observaciones SAT:** Analiza las observaciones proporcionadas. ¿Son leves o sugieren irregularidades graves?
* **Contratos Gubernamentales:** Evalúa el historial de contratos. ¿Hay patrones inusuales? ¿Montos significativos?
* **Empleados con Mismo RFC (Potenciales Conflictos de Interés):** Si los datos de "Empleados con Mismo RFC" revelan coincidencias entre el RFC del proveedor y empleados en la plantilla laboral de gobierno, identifica el **nombre del empleado y la dependencia donde trabaja**. Esto es crítico para detectar posibles conflictos de interés. **Indica si el empleado es o ha sido directivo o funcionario público.**
* **Representantes Legales como Empleados de Gobierno:** Cruza la información de los "Representantes Legales en DGCyOP" con los registros de empleados gubernamentales. Si un representante legal del proveedor es, o ha sido recientemente, un empleado o funcionario de gobierno, proporciona su **nombre, puesto y dependencia**. Esto señala otro posible conflicto de interés.
* **Síntesis de Hallazgos Web:** Agrupa y sintetiza los hallazgos web más relevantes, priorizando aquellos que involucren directamente al proveedor o sus representantes legales en situaciones de riesgo o controversia.

---

### 3. Generación del Reporte Final para Directivos

El reporte debe ser claro, conciso y orientado a la acción, con un máximo de **tres párrafos** para el cuerpo principal y la siguiente estructura:

a)  **Descripción del Proveedor:** Presenta brevemente las actividades principales de la empresa/persona.
b)  **Resumen de Observaciones SAT:** Máximo 3 líneas, destacando la implicación.
c)  **Análisis de Contratos en Gobierno:** Resalta la relevancia de los contratos existentes (ej. "Proveedor con X contratos, destacando uno de Y pesos para Z").
d)  **Hallazgos Clave y Posibles Conflictos de Interés:**
    * **Hallazgos Externos (Web):** Menciona explícitamente las fuentes periodísticas o de autoridades que vinculen al proveedor con controversias, escándalos o investigaciones. Si no hay, indica "No se encontraron referencias públicas relevantes para el proveedor en fuentes de riesgo".
    * **Coincidencias de RFC/Nombre con Empleados de Gobierno (Plantilla Laboral):** Si se encontró un empleado de gobierno con el mismo RFC del proveedor, reporta: "Se identificó una coincidencia de RFC con un empleado de gobierno: **[Nombre del Empleado]** trabajando en **[Dependencia donde trabaja]**. Este hallazgo no se vincula directamente al proveedor bajo revisión, pero se informa para transparencia."
    * **Representantes Legales con Rol en Gobierno:** Si un representante legal del proveedor es o fue empleado de gobierno, indica: "Uno de los representantes legales del proveedor, **[Nombre del Representante Legal]**, se identifica/ó como empleado en **[Puesto y Dependencia]**."
e)  **Evaluación de Riesgo y Recomendación:**
    * **Riesgo Identificado:** Clasifica el nivel de riesgo con base en los hallazgos: "Nulo", "Bajo", "Moderado", "Alto". Justifica brevemente.
    * **Recomendación Directiva Clara:** Categoriza como "APROBADO", "APROBADO CON OBSERVACIONES" (especificando cuáles), o "RECHAZADO". Esta recomendación debe ser directa y fundamentada en el análisis.

f)  **Enlaces de Evidencia:** Incluye los URLs relevantes de todas las fuentes consultadas.
g)  **Imágenes/Fotos:** Si existen imágenes relevantes (logos, recortes de noticias, etc.), agrégalas.

---

### 4. Consideraciones Adicionales

* **Ausencia de Información:** Si no se encuentra información relevante sobre el proveedor en la web (y no hay coincidencias de RFC/nombre con empleados de gobierno que justifiquen una mención específica), indica "No se encontraron referencias públicas relevantes para el proveedor."
* **Ignorar Datos No Disponibles:** Si algún dato del proveedor en la entrada no está presente, simplemente ignóralo o indica que "No aplica" en el reporte.

`,
		rfc,
		informacionDelProveedorStr,
		observacionesStr,
		contratosStr,
		empleadosEncontradosStr,
		representantesLegalesStr,
		// Se repite el RFC aquí para la búsqueda web
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
