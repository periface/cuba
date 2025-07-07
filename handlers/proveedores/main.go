package proveedores

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/periface/cuba/internals/llm"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/services/appsheets"
	"github.com/periface/cuba/internals/services/proveedores"
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
		empty := []map[string]string{}
		component := views.Buscar(empty, "")
		return renderers.RenderNoLayout(c, http.StatusOK, component)
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
		proveedorInfo := fetchProveedorInfo(rfcQuery, appsheetsInstance)
		llmResponse, err := proveedorAnalisis.ReviewProveedor(rfcQuery, proveedorInfo)
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

		component := views.Buscar(proveedorInfo.ObservacionesSat, llmResponse.Response)

		return renderers.RenderNoLayout(c, http.StatusOK, component)
	}
}

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
func buscarRepresentantesLegales(rfc string, instance *appsheets.Appsheets) ([]map[string]string, error) {

	query := `Filter(REPRESENTANTES LEGALES, [RFC]=${rfc})`
	query = strings.ReplaceAll(query, "${rfc}", rfc)
	return instance.Search("REPRESENTANTES LEGALES", models.AppSheetsPayload{
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
	query := `Filter(EMPLEADOS, [RFC]=${rfc})`
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

func fetchProveedorInfo(rfcQuery string, appsheetsInstance *appsheets.Appsheets) models.BuscarResponse {
	observacionesSat := proveedores.BuscarPorRfc(rfcQuery)
	empleadosDeGobierno, err := buscarEmpleadosDeGobierno(rfcQuery, appsheetsInstance)
	fmt.Println(empleadosDeGobierno)
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
	proveedorInfo := models.BuscarResponse{
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
			"Concepto detallado del contrato",
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
		RepresentantesLegales: getOnlyThisProps(representantesLegales, []string{
			"Concatenado",
		}),
	}
	return proveedorInfo
}
