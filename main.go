package main

import (
	"errors"
	"fmt"
	_ "fmt"
	_ "html"
	"log/slog"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/periface/cuba/internals/db"
	"github.com/periface/cuba/internals/models"
)

func buscarHandler(c echo.Context) error {
	rfcQuery := c.QueryParam("rfc")
	if rfcQuery == "" {
		slog.Error("error")
		return c.JSON(http.StatusInternalServerError, nil)
	} else {
		proveedor := buscaProveedor(rfcQuery)
		return c.JSON(http.StatusOK, proveedor)
	}
}
func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/buscar", buscarHandler)
	if err := e.Start(":1634"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
func buscaProveedor(rfc string) []models.CondonacionSAT {
	dbInstance, err := db.GetDBInstance()
	if err != nil {
		fmt.Println("Error getting database instance:", err)
		return nil
	}
	// search proveedor in unified table
	query := "SELECT * FROM unified_table WHERE rfc = ?"
	rows, err := dbInstance.Database.Query(query, rfc)
	if err != nil {
		fmt.Println("Error querying unified table:", err)
		return nil
	}
	response := []models.CondonacionSAT{}
	defer rows.Close()
	for rows.Next() {
		proveedorEncontrado := models.CondonacionSAT{}
		if err := rows.Scan(
			&proveedorEncontrado.ID,
			&proveedorEncontrado.AdministracionGeneralResponsableDeLaCancelacion,
			&proveedorEncontrado.Ao,
			&proveedorEncontrado.Contribuyente,
			&proveedorEncontrado.Ejercicio,
			&proveedorEncontrado.EntidadFederativa,
			&proveedorEncontrado.FechaAutorizacion,
			&proveedorEncontrado.FechaCancelacion,
			&proveedorEncontrado.FechaCancelacionCSD,
			&proveedorEncontrado.FechaPublicacion,
			&proveedorEncontrado.FechaPublicacionConMontoLeyTransparencia,
			&proveedorEncontrado.FechasPrimeraPublicacion,
			&proveedorEncontrado.Importe,
			&proveedorEncontrado.ImporteCondonado,
			&proveedorEncontrado.Monto,
			&proveedorEncontrado.Motivo,
			&proveedorEncontrado.MotivoCondonacion,
			&proveedorEncontrado.NumeroFechaOficioGlobalContribuyentesDof,
			&proveedorEncontrado.NumeroFechaOficioGlobalContribuyentesSat,
			&proveedorEncontrado.NumeroFechaOficioGlobalDefinitivosDof,
			&proveedorEncontrado.NumeroFechaOficioGlobalDefinitivosSat,
			&proveedorEncontrado.NumeroFechaOficioGlobalPresuncionDof,
			&proveedorEncontrado.NumeroFechaOficioGlobalPresuncionSat,
			&proveedorEncontrado.NumeroFechaOficioGlobalSentenciaFavorableDof,
			&proveedorEncontrado.NumeroFechaOficioGlobalSentenciaFavorableSat,
			&proveedorEncontrado.NumeroFechaOficioGlobalDefinitivoDof,
			&proveedorEncontrado.NumeroFechaOficioGlobalDefinitivoSat,
			&proveedorEncontrado.No,
			&proveedorEncontrado.NombreContribuyente,
			&proveedorEncontrado.NombreDenominacionRazonSocial,
			&proveedorEncontrado.NombreRazonSocial,
			&proveedorEncontrado.Periodo,
			&proveedorEncontrado.PublicacionDofDefinitivo,
			&proveedorEncontrado.PublicacionDofDefinitivos,
			&proveedorEncontrado.PublicacionDofDesvirtuados,
			&proveedorEncontrado.PublicacionDofPresuntos,
			&proveedorEncontrado.PublicacionDofSentenciaFavorable,
			&proveedorEncontrado.PublicacionPaginaSatDefinitivo,
			&proveedorEncontrado.PublicacionPaginaSatDefinitivos,
			&proveedorEncontrado.PublicacionPaginaSatDesvirtuados,
			&proveedorEncontrado.PublicacionPaginaSatPresuntos,
			&proveedorEncontrado.PublicacionPaginaSatSentenciaFavorable,
			&proveedorEncontrado.RazonSocial,
			&proveedorEncontrado.RFC,
			&proveedorEncontrado.SituacionContribuyente,
			&proveedorEncontrado.Supuesto,
			&proveedorEncontrado.SupuestoCancelacionCSD,
			&proveedorEncontrado.TipoPersona,
			&proveedorEncontrado.TipoPersona2,
			&proveedorEncontrado.LastUpdate,
			&proveedorEncontrado.FileName,
		); err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}
		response = append(response, proveedorEncontrado)
	}
	return response
}
