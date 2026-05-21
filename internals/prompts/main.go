package prompts

import (
	"encoding/json"
	"fmt"

	"github.com/periface/cuba/internals/models"
)

type PromptContext struct {
	RFC string `json:"rfc"`

	InformacionProveedor []models.InternalSearchResult `json:"informacion_proveedor"`

	RepresentantesLegales []models.InternalSearchResult `json:"representantes_legales"`

	Contratos []models.InternalSearchResult `json:"contratos"`

	EmpleadosCoincidentes []models.InternalSearchResult `json:"empleados_coincidentes"`

	ObservacionesSAT []models.InternalSearchResult `json:"observaciones_sat"`
}

func toPrettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}

	return string(b)
}

func buildProveedoresPrompt(
	rfc string,
	buscarResponse models.BuscarResponse,
) string {

	ctx := PromptContext{
		RFC: rfc,

		InformacionProveedor: buscarResponse.InformacionDelProveedor,

		RepresentantesLegales: buscarResponse.RepresentantesLegales,

		Contratos: buscarResponse.ContratosEncontrados,

		EmpleadosCoincidentes: buscarResponse.EmpleadosEncontrados,

		ObservacionesSAT: buscarResponse.ObservacionesSat,
	}

	contextJSON := toPrettyJSON(ctx)

	prompt := fmt.Sprintf(`
Eres un analista de integridad y riesgo gubernamental.

OBJETIVO:
Evaluar riesgos reputacionales,
conflictos de interés,
riesgos de corrupción
y posibles irregularidades relacionadas con un proveedor.

Debes generar un reporte ejecutivo claro,
objetivo y útil para toma de decisiones directivas.

REGLAS IMPORTANTES:
- No inventes información.
- No asumas relaciones sin evidencia.
- Distingue claramente entre:
  - hechos confirmados
  - investigaciones
  - acusaciones
  - notas periodísticas
- Si no existe evidencia suficiente:
  indicar explícitamente:
  "No se encontraron hallazgos públicos relevantes."
- Prioriza:
  - fuentes oficiales
  - fiscalías
  - auditorías
  - listas de sanciones
  - medios periodísticos confiables

INVESTIGACIÓN WEB:
Realiza búsquedas usando:
- RFC
- razón social
- representantes legales

Buscar señales relacionadas con:
- corrupción
- fraude
- lavado de dinero
- evasión fiscal
- sanciones
- inhabilitaciones
- conflicto de interés
- desvío de recursos
- sobornos
- demandas relevantes

ANÁLISIS INTERNO:
Evalúa:
- observaciones SAT
- historial de contratos
- coincidencias con empleados gubernamentales
- posibles vínculos gubernamentales relevantes

SCORING DE RIESGO:
Usa estas reglas de referencia:
- +50 investigación penal
- +40 inhabilitación o sanción oficial
- +30 observaciones SAT graves
- +20 conflicto de interés documentado
- +10 múltiples notas periodísticas negativas

CLASIFICACIÓN:
- 0-9 -> NULO
- 10-29 -> BAJO
- 30-59 -> MODERADO
- 60+ -> ALTO

DATOS DEL PROVEEDOR:
%s

FORMATO DE RESPUESTA:

# Resumen Ejecutivo

# Información General del Proveedor

# Hallazgos Relevantes

# Riesgos Detectados

# Evaluación Final
Nivel de riesgo: NULO | BAJO | MODERADO | ALTO

# Recomendación
APROBADO |
APROBADO CON OBSERVACIONES |
RECHAZADO

# Fuentes Consultadas
- listar URLs

# Observaciones Finales

`, contextJSON)

	return prompt
}

func AnalisisDeProveedoresPrompt(
	rfc string,
	proveedorInfo models.BuscarResponse,
) (string, models.BuscarResponse) {

	response := buildProveedoresPrompt(
		rfc,
		proveedorInfo,
	)
	return response, proveedorInfo
}
