package prompts

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/services/appsheets"
	"github.com/periface/cuba/internals/services/proveedores"
	"github.com/periface/cuba/internals/utils"
)

type PromptContext struct {
	RFC string `json:"rfc"`

	InformacionProveedor []map[string]string `json:"informacion_proveedor"`

	RepresentantesLegales []map[string]string `json:"representantes_legales"`

	Contratos []map[string]string `json:"contratos"`

	EmpleadosCoincidentes []map[string]string `json:"empleados_coincidentes"`

	ObservacionesSAT []map[string]string `json:"observaciones_sat"`
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
	appsheetsInstance appsheets.Appsheets,
) string {

	proveedorInfo := fetchProveedorInfo(
		rfc,
		&appsheetsInstance,
	)

	return buildProveedoresPrompt(
		rfc,
		proveedorInfo,
	)
}

func buscarProveedorEnAtcom(
	rfc string,
	instance *appsheets.Appsheets,
) ([]map[string]string, error) {

	query := `Filter(PADRON DE PROVEEDORES, [RFC]=${rfc})`

	query = strings.ReplaceAll(
		query,
		"${rfc}",
		rfc,
	)

	return instance.Search(
		"PADRON DE PROVEEDORES",
		models.AppSheetsPayload{
			Action: "Find",
			Properties: map[string]string{
				"Selector": query,
			},
		},
	)
}

func buscarRepresentantesLegales(
	rfc string,
	instance *appsheets.Appsheets,
) ([]map[string]string, error) {

	query := `Filter(REPRESENTANTES LEGALES, [RFC]=${rfc})`

	query = strings.ReplaceAll(
		query,
		"${rfc}",
		rfc,
	)

	return instance.Search(
		"REPRESENTANTES LEGALES",
		models.AppSheetsPayload{
			Action: "Find",
			Properties: map[string]string{
				"Selector": query,
			},
		},
	)
}

func buscarEmpleadosDeGobierno(
	rfc string,
	instance *appsheets.Appsheets,
) ([]map[string]string, error) {

	APIKEY, err := utils.GetEnvVariable(
		"APPSHEETSID_RH",
	)

	if err != nil {
		return nil, err
	}

	SECRET, err := utils.GetEnvVariable(
		"APPSHEETSSECRET_RH",
	)

	if err != nil {
		return nil, err
	}

	query := `Filter(EMPLEADOS, [RFC]=${rfc})`

	query = strings.ReplaceAll(
		query,
		"${rfc}",
		rfc,
	)

	return instance.SearchIn(
		APIKEY,
		SECRET,
		"EMPLEADOS",
		models.AppSheetsPayload{
			Action: "Find",
			Properties: map[string]string{
				"Selector": query,
			},
		},
	)
}

func buscarContratos(
	rfc string,
	instance *appsheets.Appsheets,
) ([]map[string]string, error) {

	query := `Filter(CONTRATOS, [Proveedor]=${rfc})`

	query = strings.ReplaceAll(
		query,
		"${rfc}",
		rfc,
	)

	return instance.Search(
		"CONTRATOS",
		models.AppSheetsPayload{
			Action: "Find",
			Properties: map[string]string{
				"Selector": query,
			},
		},
	)
}

// iterate all maps and gets only the valid Props with their values
func getOnlyThisProps(
	inputList []map[string]string,
	validProps []string,
) []map[string]string {

	validSet := make(map[string]struct{})

	for _, prop := range validProps {
		validSet[prop] = struct{}{}
	}

	var result []map[string]string

	for _, item := range inputList {

		filtered := make(map[string]string)

		for key, value := range item {

			if _, ok := validSet[key]; ok {
				filtered[key] = value
			}
		}

		result = append(result, filtered)
	}

	return result
}

func fetchProveedorInfo(
	rfcQuery string,
	appsheetsInstance *appsheets.Appsheets,
) models.BuscarResponse {

	observacionesSat := proveedores.BuscarPorRfc(
		rfcQuery,
	)

	empleadosDeGobierno, err := buscarEmpleadosDeGobierno(
		rfcQuery,
		appsheetsInstance,
	)

	if err != nil {
		slog.Error(err.Error())
	}

	datosDelProveedor, err := buscarProveedorEnAtcom(
		rfcQuery,
		appsheetsInstance,
	)

	if err != nil {
		slog.Error(err.Error())
	}

	representantesLegales, err := buscarRepresentantesLegales(
		rfcQuery,
		appsheetsInstance,
	)

	if err != nil {
		slog.Error(err.Error())
	}

	contratos, err := buscarContratos(
		rfcQuery,
		appsheetsInstance,
	)

	if err != nil {
		slog.Error(err.Error())
	}

	return models.BuscarResponse{
		ObservacionesSat: observacionesSat,

		EmpleadosEncontrados: getOnlyThisProps(
			empleadosDeGobierno,
			[]string{
				"Partida",
				"Departamento",
				"ape_pat",
				"ape_mat",
				"nombre",
				"RFC",
			},
		),

		ContratosEncontrados: getOnlyThisProps(
			contratos,
			[]string{
				"Concepto / Objeto del Contrato",
				"No. de Contrato DGCYOP",
				"Concepto detallado del contrato",
				"Monto Total del Contrato",
			},
		),

		InformacionDelProveedor: getOnlyThisProps(
			datosDelProveedor,
			[]string{
				"RAZON SOCIAL",
				"NOMBRE DEL PROVEEDOR",
				"1ER. APELLIDO",
				"2O. APELLIDO",
				"GIRO",
				"FECHA ALTA",
				"FECHA VENCIMIENTO",
				"COORDENADAS",
			},
		),

		RepresentantesLegales: getOnlyThisProps(
			representantesLegales,
			[]string{
				"Concatenado",
			},
		),
	}
}
