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

type ProveedoresHandlers struct {
}

func NewProveedoresHandlers() *ProveedoresHandlers {
	return &ProveedoresHandlers{}
}
func (chh *ProveedoresHandlers) ProveedoresIndex(c echo.Context) error {
	rfcQuery := c.QueryParam("rfc")
	component := views.Index(rfcQuery)
	return renderers.RenderNoLayout(c, http.StatusOK, component)
}

func (chh *ProveedoresHandlers) BuscarProveedor(c echo.Context) error {
	rfcQuery := c.QueryParam("rfc")
	if rfcQuery == "" {
		slog.Error("error")
		component := views.Buscar("")
		return renderers.RenderNoLayout(c, http.StatusOK, component)
	} else {
		appsheetsInstance, err := appsheets.NewAppsheets()
		if err != nil {
			slog.Error("Instance appsheets")
			slog.Error(err.Error())
		}
		prompRunner, err := llm.NewPromptRunner(llm.WithGoogleSearchTool)
		if err != nil {
			slog.Error(err.Error())
		}

		prompt := prompts.AnalisisDeProveedoresPrompt(rfcQuery, *appsheetsInstance)

		llmResponse, err := prompRunner.RunPrompt(prompt)
		if err != nil {
			slog.Error(err.Error())
		}
		_, err = appsheetsInstance.Insert("PROVEEDORES_ANALISIS", models.AppSheetsPayload{
			Action: "Add",
			Rows: []map[string]string{{
				"Proveedor":          rfcQuery,
				"Contenido":          llmResponse.Response,
				"Fecha":              time.Now().Format("2006-01-02 15:04:05"),
				"Pregunta Realizada": llmResponse.Prompt,
			}},
		})

		if err != nil {
			slog.Error("Error notificando respuesta")
			slog.Error(err.Error())
		}

		component := views.Buscar(llmResponse.Response)

		return renderers.RenderNoLayout(c, http.StatusOK, component)
	}
}
