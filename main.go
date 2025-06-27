package main

import (
	"errors"
	_ "fmt"
	_ "html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/periface/cuba/internals/llm"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/services/appsheets"
	"github.com/periface/cuba/internals/services/proveedores"
	"github.com/periface/cuba/internals/utils"
	"github.com/periface/cuba/views"
)

func buscarProveedorEnAtcom(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {
	query := `Filter(PADRON DE PROVEEDORES, [RFC]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)
	return instance.Search("PADRON DE PROVEEDORES", models.AppSheetsPayload{
		Action: "Find",
		Properties: map[string]string{
			"Selector": query,
		},
	})
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
	query := `Filter(EMPLEADOS, [RFC]=${rfc}")`
	query = strings.ReplaceAll(query, "${rfc}", rfc)
	return instance.SearchIn(APIKEY, SECRET, "EMPLEADOS", models.AppSheetsPayload{
		Action: "Find",
		Properties: map[string]string{
			"Selector": query,
		},
	})
}

// iterate all maps and gets only the valid Props with their values
func getOnlyThisProps(inputList []map[string]string, validProps []string) []map[string]string {
	// Crear un set para búsqueda rápida de props válidas
	validSet := make(map[string]struct{})
	for _, prop := range validProps {
		validSet[prop] = struct{}{}
	}

	var result []map[string]string

	// Iterar sobre cada mapa de la lista
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

func buscarContratos(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {
	query := `Filter(CONTRATOS, [Proveedor]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)
	return instance.Search("CONTRATOS", models.AppSheetsPayload{
		Action: "Find",
		Properties: map[string]string{
			"Selector": query,
		},
	})
}

func buscarHandler(c echo.Context) error {
	rfcQuery := c.QueryParam("rfc")

	if rfcQuery == "" {
		slog.Error("error")
		empty := []map[string]string{}
		component := views.Buscar(empty, "")
		return RenderNoLayout(c, http.StatusOK, component)
	} else {
		appsheetsInstance, err := appsheets.NewAppsheets()
		if err != nil {
			slog.Error("Instance appsheets")
			slog.Error(err.Error())
		}
		proveedorAnalisis, err := llm.NewProveedorReviewer()
		if err != nil {
			slog.Error(err.Error())
		}

		observacionesSat := proveedores.BuscarPorRfc(rfcQuery)
		empleadosDeGobierno, err := buscarEmpleadosDeGobierno(rfcQuery, appsheetsInstance)

		if err != nil {
			slog.Error(err.Error())
		}
		datosDelProveedor, err := buscarProveedorEnAtcom(rfcQuery, appsheetsInstance)

		if err != nil {
			slog.Error(err.Error())
		}
		contratos, err := buscarContratos(rfcQuery, appsheetsInstance)

		if err != nil {
			slog.Error(err.Error())
		}
		proveedorData := models.BuscarResponse{
			ObservacionesSat: observacionesSat,
			EmpleadosEncontrados: getOnlyThisProps(empleadosDeGobierno, []string{
				"Partida",
				"Departamento",
				"ape_pat",
				"ape_mat",
				"nombre",
				"RFC",
			}),
			ContratosEncontrados: getOnlyThisProps(contratos, []string{
				"Concepto / Objeto del Contrato",
				"No. de Contrato DGCYOP",
				"Monto Total del Contrato",
			}),
			InformacionDelProveedor: getOnlyThisProps(datosDelProveedor, []string{
				"RAZON SOCIAL",
				"NOMBRE DEL PROVEEDOR",
				"1ER. APELLIDO",
				"2O. APELLIDO",
				"GIRO",
				"FECHA ALTA",
				"FECHA VENCIMIENTO",
				"COORDENADAS",
			}),
		}
		llmResponse, err := proveedorAnalisis.ReviewProveedor(rfcQuery, proveedorData)
		if err != nil {
			slog.Error(err.Error())
		}
		appsheetsInstance.Insert("PROVEEDORES_ANALISIS", models.AppSheetsPayload{
			Action: "Add",
			Rows: []map[string]string{{
				"Proveedor":          rfcQuery,
				"Contenido":          llmResponse.Response,
				"Fecha":              time.Now().Format("2006-01-02 15:04:05"),
				"Pregunta Realizada": llmResponse.Prompt,
			}},
		})

		if err != nil {
			slog.Error("Instance appsheets")
			slog.Error(err.Error())
		}

		component := views.Buscar(observacionesSat, llmResponse.Response)
		return RenderNoLayout(c, http.StatusOK, component)
	}
}
func testAppsheets(c echo.Context) error {

	rfcQuery := c.QueryParam("rfc")
	appsheets, err := appsheets.NewAppsheets()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	proveedores, err := buscarProveedorEnAtcom(rfcQuery, appsheets)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, proveedores)
}
func mainHandler(c echo.Context) error {

	rfcQuery := c.QueryParam("rfc")
	component := views.Index(rfcQuery)
	return Render("Inicio", c, http.StatusOK, component)
}
func main() {
	e := echo.New()

	e.Static("/static", "assets")
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", mainHandler)
	e.GET("/buscar", buscarHandler)

	e.GET("/test", testAppsheets)

	if err := e.Start(":1634"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

func Render(title string, ctx echo.Context, statusCode int, t templ.Component) error {
	page := views.Layout(title, t)
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if err := page.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}
	return ctx.HTML(statusCode, buf.String())
}
func RenderNoLayout(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}
	return ctx.HTML(statusCode, buf.String())
}
