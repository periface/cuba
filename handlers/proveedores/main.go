package proveedores

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo"

	"github.com/periface/cuba/internals/llm"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/prompts"
	"github.com/periface/cuba/internals/services/appsheets"
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
		return h.renderSuccess(c, "", "")
	}

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
	// LLM
	// --------------------------------------------

	promptRunner, err := llm.NewPromptRunner(
		llm.WithGoogleSearchTool,
	)

	if err != nil {

		slog.Error(
			"prompt runner init error",
			"error", err.Error(),
			"rfc", rfc,
		)

		return h.renderError(
			c,
			http.StatusInternalServerError,
			"Error inicializando motor de análisis",
		)
	}

	// --------------------------------------------
	// Prompt
	// --------------------------------------------

	basePrompt := prompts.AnalisisDeProveedoresPrompt(
		rfc,
		*appsheetClient,
	)

	searchQueries := buildSearchQueries(rfc)

	// --------------------------------------------
	// Execute analysis
	// --------------------------------------------

	slog.Info(
		"starting provider analysis",
		"rfc", rfc,
	)

	llmResponse, err := promptRunner.RunPromptWithGoogle(
		basePrompt,
		searchQueries,
	)

	if err != nil {

		slog.Error(
			"llm execution error",
			"error", err.Error(),
			"rfc", rfc,
		)

		return h.renderError(
			c,
			http.StatusInternalServerError,
			"Error ejecutando análisis del proveedor",
		)
	}

	// --------------------------------------------
	// Persist analysis
	// --------------------------------------------

	go saveAnalysis(
		appsheetClient,
		rfc,
		llmResponse,
	)

	// --------------------------------------------
	// Render response
	// --------------------------------------------

	return h.renderSuccess(
		c,
		llmResponse.Response,
		llmResponse.Prompt,
	)
}

func buildSearchQueries(rfc string) []string {

	return []string{
		rfc + " corrupción",
		rfc + " fraude",
		rfc + " SAT",
		rfc + " contratos gobierno",
		rfc + " lavado de dinero",
		rfc + " conflicto de interés",
		rfc + " inhabilitado",
		rfc + " observaciones ASF",
	}
}

func saveAnalysis(
	appsheetClient *appsheets.Appsheets,
	rfc string,
	response models.LLMResponse,
) {

	_, err := appsheetClient.Insert(
		"PROVEEDORES_ANALISIS",
		models.AppSheetsPayload{
			Action: "Add",
			Rows: []map[string]string{
				{
					"Proveedor":          rfc,
					"Contenido":          response.Response,
					"Fecha":              time.Now().Format("2006-01-02 15:04:05"),
					"Pregunta Realizada": response.Prompt,
				},
			},
		},
	)

	if err != nil {

		slog.Error(
			"error saving provider analysis",
			"error", err.Error(),
			"rfc", rfc,
		)

		return
	}

	slog.Info(
		"provider analysis saved",
		"rfc", rfc,
	)
}

func (h *ProveedoresHandlers) renderSuccess(
	c echo.Context,
	response string,
	prompt string,
) error {

	component := views.Buscar(
		response,
		prompt,
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
	message string,
) error {

	component := views.Buscar(
		message,
		"",
	)

	return renderers.RenderNoLayout(
		c,
		status,
		component,
	)
}

