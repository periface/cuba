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
	"github.com/periface/cuba/internals/services/proveedores"
	searchxng "github.com/periface/cuba/internals/services/searchXNG"
	"github.com/periface/cuba/internals/utils"
	"github.com/periface/cuba/views"
)

var renderers = utils.NewRenderers()

const MIN_SCORE float64 = 1

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
		n := 0
		for _, result := range proveedorSearch.Results {
			fmt.Println(result.Score)
			if result.Score >= MIN_SCORE {
				proveedorSearch.Results[n] = result
				n++
			}
		}

		proveedorSearch.Results = proveedorSearch.Results[:n]
		proveedorInfo.InformacionDelProveedor[i].SearxngResponse = proveedorSearch
	}

	for i, representante := range proveedorInfo.RepresentantesLegales {
		representanteSearch, err := searchxngClient.AdvancedSearch(representante.Values["Concatenado"], categorias, motores, 3)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}
		n := 0
		for _, result := range representanteSearch.Results {

			fmt.Println(result.Score)
			if result.Score >= MIN_SCORE {
				representanteSearch.Results[n] = result
				n++
			}
		}
		representanteSearch.Results = representanteSearch.Results[:n]
		proveedorInfo.RepresentantesLegales[i].SearxngResponse = representanteSearch
	}

	for i, socio := range proveedorInfo.Socios {

		socioSearch, err := searchxngClient.AdvancedSearch(socio.Values["Nombre/Razón Social del Socio/Accionista"], categorias, motores, 3)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}

		n := 0

		for _, result := range socioSearch.Results {

			fmt.Println(result.Score)
			if result.Score >= MIN_SCORE {
				socioSearch.Results[n] = result
				n++
			}
		}
		socioSearch.Results = socioSearch.Results[:n]
		proveedorInfo.Socios[i].SearxngResponse = socioSearch

		fmt.Print("SOCIOS DONE")
	}

	for i, empleado := range proveedorInfo.EmpleadosEncontrados {
		fmt.Print("ENTRANDO DONE")

		fmt.Print(empleado.Values)

		nombre := empleado.Values["nombre"] + " " + empleado.Values["ape_pat"] + " " + empleado.Values["ape_mat"]

		empleadoSearch, err := searchxngClient.AdvancedSearch(
			nombre,
			categorias, motores, 3)
		if err != nil {
			slog.Error("error busqueda riesgo", "error", err.Error(), "rfc", rfc)
		}
		n := 0
		for _, result := range empleadoSearch.Results {

			fmt.Println(result.Score)
			if result.Score >= MIN_SCORE {
				empleadoSearch.Results[n] = result
				n++
			}
		}
		empleadoSearch.Results = empleadoSearch.Results[:n]
		proveedorInfo.EmpleadosEncontrados[i].SearxngResponse = empleadoSearch
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
