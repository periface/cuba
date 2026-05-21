package proveedores

import (
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo"

	"github.com/periface/cuba/internals/llm"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/prompts"
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
	// Extracción de Datos e Identidad
	// --------------------------------------------
	proveedorInfo := proveedores.FetchProveedorInfo(rfc)

	// --------------------------------------------
	// Extracción de Datos e Identidad
	// --------------------------------------------
	basePrompt, data := prompts.AnalisisDeProveedoresPrompt(rfc, proveedorInfo)

	// Configuración para la API de SearXNG
	categorias := []string{"news", "general"}
	motores := []string{"google_news", "bing_news", "google"}

	// --------------------------------------------
	// EJECUCIÓN DE BÚSQUEDA POR PROVEEDOR/REPRESENTANTE/SOCIOS [VOLVER ASYNC PARA EVITAR BLOQUEOS]
	// --------------------------------------------
	for i, proveedor := range proveedorInfo.InformacionDelProveedor {
		searchQuery := buildSingleCleanSearchQuery(rfc, proveedor)
		proveedorSearch, err := searchxngClient.AdvancedSearch(searchQuery, categorias, motores, 10)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}
		proveedorInfo.InformacionDelProveedor[i].SearxngResponse = proveedorSearch
	}

	for i, representante := range proveedorInfo.RepresentantesLegales {

		slog.Info(representante.Values["Concatenado"])

		representanteSearch, err := searchxngClient.AdvancedSearch(representante.Values["Concatenado"], categorias, motores, 3)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}
		proveedorInfo.RepresentantesLegales[i].SearxngResponse = representanteSearch
	}

	for i, socio := range proveedorInfo.Socios {

		slog.Info(socio.Values["Nombre/Razón Social del Socio/Accionista"])

		socioSearch, err := searchxngClient.AdvancedSearch(socio.Values["Nombre/Razón Social del Socio/Accionista"], categorias, motores, 3)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}
		proveedorInfo.Socios[i].SearxngResponse = socioSearch
	}

	// --------------------------------------------
	// Render response
	// --------------------------------------------
	return h.renderSuccess(
		c,
		data,
		basePrompt,
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
		razonSocial = data.InformacionDelProveedor[0].Values["RAZON SOCIAL"]
		nombreProveedor = data.InformacionDelProveedor[0].Values["NOMBRE DEL PROVEEDOR"]
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
func buildSingleCleanSearchQuery(rfc string, data models.InternalSearchResult) string {
	var razonSocial, nombreProveedor string

	razonSocial = data.Values["RAZON SOCIAL"]
	nombreProveedor = data.Values["NOMBRE DEL PROVEEDOR"] + " " +
		data.Values["1ER. APELLIDO"] + " " +
		data.Values["2O. APELLIDO"]
		//"1ER. APELLIDO", "2O. APELLIDO"

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

// buildCleanSearchQuery genera una consulta pura de identidad sin condicionales de peligro.
func buildCleanSearchQuery(rfc string, data models.BuscarResponse) string {
	var razonSocial, nombreProveedor string

	if len(data.InformacionDelProveedor) > 0 {
		razonSocial = data.InformacionDelProveedor[0].Values["RAZON SOCIAL"]
		nombreProveedor = data.InformacionDelProveedor[0].Values["NOMBRE DEL PROVEEDOR"] + " " +
			data.InformacionDelProveedor[0].Values["1ER. APELLIDO"] + " " +
			data.InformacionDelProveedor[0].Values["2O. APELLIDO"]
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
) error {
	component := views.Buscar(
		models.BuscarViewModel{
			Data:   data,
			Prompt: prompt,
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
