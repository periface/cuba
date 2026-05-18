package proveedores

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"github.com/periface/cuba/internals/llm"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/prompts"
	"github.com/periface/cuba/internals/services/appsheets"
	"github.com/periface/cuba/internals/services/proveedores"
	searchxng "github.com/periface/cuba/internals/services/searchXNG"
	"github.com/periface/cuba/internals/utils"
	"github.com/periface/cuba/views"
)

var renderers = utils.NewRenderers()

type ProveedoresHandlers struct{}

func NewProveedoresHandlers() *ProveedoresHandlers {
	return &ProveedoresHandlers{}
}

func (h *ProveedoresHandlers) ProveedoresIndex(c echo.Context) error {
	rfc := c.QueryParam("rfc")

	return renderers.RenderNoLayout(
		c,
		http.StatusOK,
		views.Index(rfc),
	)
}

func (h *ProveedoresHandlers) BuscarProveedor(c echo.Context) error {
	rfc := c.QueryParam("rfc")

	if rfc == "" {
		return h.renderSuccess(c, models.BuscarResponse{},
			"Error, no RFC",
			models.SearxngResponse{},
			models.SearxngResponse{},
			"",
			"",
		)
	}

	// --------------------------------------------
	// SEARCHXNG CLIENT
	// --------------------------------------------
	serverPath, err := utils.GetEnvVariable("SEARCH_SERVER")
	if err != nil {
		log.Print("Defaulting a local")
		serverPath = "http://localhost:1991"
	}
	searchxngClient := searchxng.NewSearXNGClient(serverPath)

	// --------------------------------------------
	// AppSheets
	// --------------------------------------------
	appsheetClient, err := appsheets.NewAppsheets()
	if err != nil {
		slog.Error(
			"appsheets init error",
			"error", err.Error(),
			"rfc", rfc,
		)
		return h.renderError(
			c,
			http.StatusInternalServerError,
			"Error inicializando AppSheets",
		)
	}

	// --------------------------------------------
	// Extracción de Datos e Identidad
	// --------------------------------------------
	proveedorInfo := fetchProveedorInfo(rfc, appsheetClient)
	basePrompt, data := prompts.AnalisisDeProveedoresPrompt(rfc, proveedorInfo)

	// Configuración para la API de SearXNG
	categorias := []string{"news", "general"}
	motores := []string{"google_news", "bing_news", "google"}

	// --------------------------------------------
	// EJECUCIÓN DE BÚSQUEDA DOBLE
	// --------------------------------------------

	// 1. Búsqueda de Alertas de Riesgo (Filtros estrictos con AND)
	queryRiesgo := buildSearchQueryString(rfc, proveedorInfo)
	var responseRiesgo models.SearxngResponse
	if queryRiesgo != "" {
		responseRiesgo, err = searchxngClient.AdvancedSearch(queryRiesgo, categorias, motores)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}
	}

	// 2. Búsqueda Limpia (Sólo nombres/razón social para capturar notas de prensa normales)
	queryLimpio := buildCleanSearchQuery(rfc, proveedorInfo)
	var responseLimpia models.SearxngResponse
	if queryLimpio != "" {
		responseLimpia, err = searchxngClient.AdvancedSearch(queryLimpio, categorias, motores)
		if err != nil {
			slog.Error("error busqueda limpia", "error", err.Error(), "rfc", rfc)
		}
	}

	// 3. Unificación inteligente de resultados sin duplicar URLs

	// --------------------------------------------
	// Render response
	// --------------------------------------------
	return h.renderSuccess(
		c,
		data,
		basePrompt,
		responseLimpia,
		responseRiesgo,
		queryLimpio,
		queryRiesgo,
	)
}

type AnalisisInput struct {
	rfc           string
	prompt        string
	proveedorInfo models.BuscarResponse
}

func (h *ProveedoresHandlers) CorrerAnalisis(c echo.Context) error {
	input := new(AnalisisInput)

	if err := c.Bind(input); err != nil {
		return err
	}

	// --------------------------------------------
	// LLM
	// --------------------------------------------
	promptRunner, err := llm.NewPromptRunner(
		llm.WithGoogleSearchTool,
	)
	if err != nil {
		slog.Error(
			"prompt runner init error",
			"error", err.Error(),
			"rfc", input.rfc,
		)
		return c.JSON(http.StatusInternalServerError, nil)
	}

	// --------------------------------------------
	// Execute analysis
	// --------------------------------------------
	slog.Info(
		"starting provider analysis",
		"rfc", input.rfc,
	)

	llmResponse, err := promptRunner.RunPrompt(
		input.prompt,
	)
	if err != nil {
		slog.Error(
			"llm execution error",
			"error", err.Error(),
			"rfc", input.rfc,
		)
		return c.JSON(http.StatusInternalServerError, nil)
	}
	return c.JSON(http.StatusOK, llmResponse)
}

// buildSearchQueryString genera la consulta restrictiva de riesgos usando operadores lógicos llimpios.
func buildSearchQueryString(rfc string, data models.BuscarResponse) string {
	var razonSocial, nombreProveedor string

	if len(data.InformacionDelProveedor) > 0 {
		razonSocial = data.InformacionDelProveedor[0]["RAZON SOCIAL"]
		nombreProveedor = data.InformacionDelProveedor[0]["NOMBRE DEL PROVEEDOR"]
	}

	var identities []string
	if rfc != "" {
		identities = append(identities, `"`+rfc+`"`)
	}
	if razonSocial != "" {
		identities = append(identities, `"`+razonSocial+`"`)
	}
	if nombreProveedor != "" && nombreProveedor != razonSocial {
		identities = append(identities, `"`+nombreProveedor+`"`)
	}

	if len(identities) == 0 {
		return ""
	}
	identityQuery := "(" + strings.Join(identities, " OR ") + ")"

	keywords := []string{
		"corrupción", "fraude", "SAT", `"lavado de dinero"`,
		`"conflicto de interés"`, "inhabilitado", "ASF", "multa",
		"desvío", `"empresas fantasma"`, "investigación", "denuncia",
		`"lista negra"`, "SFP",
	}
	keywordsQuery := "(" + strings.Join(keywords, " OR ") + ")"

	return identityQuery + " AND " + keywordsQuery
}

// buildCleanSearchQuery genera una consulta pura de identidad sin condicionales de peligro.
func buildCleanSearchQuery(rfc string, data models.BuscarResponse) string {
	var razonSocial, nombreProveedor string

	if len(data.InformacionDelProveedor) > 0 {
		razonSocial = data.InformacionDelProveedor[0]["RAZON SOCIAL"]
		nombreProveedor = data.InformacionDelProveedor[0]["NOMBRE DEL PROVEEDOR"] + " " +
			data.InformacionDelProveedor[0]["1ER. APELLIDO"] + " " +
			data.InformacionDelProveedor[0]["2O. APELLIDO"]
		//"1ER. APELLIDO", "2O. APELLIDO"
	}

	var identities []string
	if razonSocial != "" {
		identities = append(identities, `"`+razonSocial+`"`)
	}
	if nombreProveedor != "" && nombreProveedor != razonSocial {
		identities = append(identities, `"`+nombreProveedor+`"`)
	}
	if len(identities) == 0 && rfc != "" {
		identities = append(identities, `"`+rfc+`"`)
	}

	return strings.Join(identities, " OR ")
}

func (h *ProveedoresHandlers) renderSuccess(
	c echo.Context,
	data models.BuscarResponse,
	prompt string,
	searchEngine models.SearxngResponse,
	searchEngineClean models.SearxngResponse,
	queryClean string,
	queryRiesgo string,
) error {
	component := views.Buscar(
		models.BuscarViewModel{
			Data:              data,
			Prompt:            prompt,
			SearchEngine:      searchEngine,
			SearchEngineClean: searchEngineClean,
			QueryClean:        queryClean,
			QueryRiesgo:       queryRiesgo,
		},
	)

	return renderers.RenderNoLayout(
		c,
		http.StatusOK,
		component,
	)
}

func (h *ProveedoresHandlers) renderError(
	c echo.Context,
	status int,
	_ string,
) error {
	component := views.Buscar(
		models.BuscarViewModel{},
	)
	return renderers.RenderNoLayout(
		c,
		status,
		component,
	)
}

func fetchProveedorInfo(
	rfcQuery string,
	appsheetsInstance *appsheets.Appsheets,
) models.BuscarResponse {
	observacionesSat := proveedores.BuscarPorRfc(rfcQuery)

	empleadosDeGobierno, err := buscarEmpleadosDeGobierno(rfcQuery, appsheetsInstance)
	if err != nil {
		slog.Error(err.Error())
	}

	datosDelProveedor, err := buscarProveedorEnAtcom(rfcQuery, appsheetsInstance)
	if err != nil {
		slog.Error(err.Error())
	}

	representantesLegales, err := buscarRepresentantesLegales(rfcQuery, appsheetsInstance)
	if err != nil {
		slog.Error(err.Error())
	}

	contratos, err := buscarContratos(rfcQuery, appsheetsInstance)
	if err != nil {
		slog.Error(err.Error())
	}

	fmt.Println(empleadosDeGobierno)

	return models.BuscarResponse{
		ObservacionesSat: observacionesSat,

		EmpleadosEncontrados: getOnlyThisProps(
			empleadosDeGobierno,
			[]string{"Partida", "Departamento", "ape_pat", "ape_mat", "nombre", "RFC"},
		),

		ContratosEncontrados: getOnlyThisProps(
			contratos,
			[]string{"Concepto / Objeto del Contrato", "No. de Contrato DGCYOP", "Concepto detallado del contrato", "Monto Total del Contrato"},
		),

		InformacionDelProveedor: getOnlyThisProps(
			datosDelProveedor,
			[]string{"RAZON SOCIAL", "NOMBRE DEL PROVEEDOR", "1ER. APELLIDO", "2O. APELLIDO", "GIRO", "FECHA ALTA", "FECHA VENCIMIENTO", "COORDENADAS"},
		),

		RepresentantesLegales: getOnlyThisProps(
			representantesLegales,
			[]string{"Concatenado"},
		),
	}
}

func buscarProveedorEnAtcom(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {
	query := `Filter(PADRON DE PROVEEDORES, [RFC]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)

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

func buscarRepresentantesLegales(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {
	query := `Filter(REPRESENTANTES LEGALES, [RFC]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)

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

func buscarEmpleadosDeGobierno(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {
	APIKEY, err := utils.GetEnvVariable("APPSHEETSID_RH")
	if err != nil {
		return nil, err
	}

	SECRET, err := utils.GetEnvVariable("APPSHEETSSECRET_RH")
	if err != nil {
		return nil, err
	}

	query := `Filter(EMPLEADOS, [RFC]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)

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

func buscarContratos(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {
	query := `Filter(CONTRATOS, [Proveedor]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)

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

func getOnlyThisProps(inputList []map[string]string, validProps []string) []map[string]string {
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
