package proveedores

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/periface/cuba/internals/db"
	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/services/appsheets"
)

func filtraCamposVacios(obj any, etiquetas map[string]string) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := t.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		value := fmt.Sprintf("%v", field.Interface())
		if strings.TrimSpace(value) == "" || value == "0" {
			continue // omitimos campos vacíos o cero
		}

		nombreLegible := etiquetas[jsonTag]
		if nombreLegible == "" {
			nombreLegible = jsonTag // fallback al nombre crudo
		}

		result[nombreLegible] = value
	}

	return result
}
func buscaClasificacionDocumento(archivo string,
	clasificaciones []map[string]string) map[string]string {
	response := make(map[string]string)
	for _, clasificacion := range clasificaciones {
		if clasificacion["Archivo"] == archivo {
			response["nombre"] = clasificacion["nombre"]
			response["descripcion"] = clasificacion["Descripción"]
			return response
		}
	}
	return response
}
func BuscarPorRfc(rfc string) []map[string]string {
	proveedores := buscaProveedor(rfc)
	app, err := appsheets.NewAppsheets()
	if err != nil {
		fmt.Println("Error al cargar appsheets", err.Error())
	}
	data, err := app.GetTable("CLASIFICACION")
	if err != nil {
		fmt.Println("Error al cargar los datos", err.Error())
	}
	response := make([]map[string]string, 0)
	for _, val := range proveedores {
		clasificacion := buscaClasificacionDocumento(val.FileName, data)
		val.Clasificacion = clasificacion["nombre"]
		val.ClasificacionDescription = clasificacion["descripcion"]
		data := filtraCamposVacios(val, etiquetas)
		response = append(response, data)
	}
	return response
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

	fmt.Println("Proveedores:", response)
	return response
}

// Etiquetas encontradas dentro del json
var etiquetas = map[string]string{
	"id": "ID",
	"administracin_general_responsable_de_la_cancelacion": "Administración General Responsable de la Cancelación",
	"ao":                       "Año",
	"contribuyente":            "Contribuyente",
	"ejercicio":                "Ejercicio",
	"entidad_federativa":       "Estado",
	"fecha_de_autorizacin":     "Fecha de Autorización",
	"fecha_de_cancelacion":     "Fecha de Cancelación",
	"fecha_de_cancelacion_csd": "Fecha de Cancelación CSD",
	"fecha_de_publicacin":      "Fecha de Publicación",
	"fecha_de_publicacin_con_monto_de_acuerdo_a_la_ley_de_transparencia": "Fecha de Publicación con Monto (Ley de Transparencia)",
	"fechas_de_primera_publicacion":                                      "Fecha de Primera Publicación",
	"importe":                                                            "Importe",
	"importe_condonado":                                                  "Importe Condonado",
	"monto":                                                              "Monto",
	"motivo":                                                             "Motivo",
	"motivo_de_condonacin":                                               "Motivo de Condonación",
	"nmero_y_fecha_de_oficio_global_de_contribuyentes_que_desvirtuaron_dof": "Oficio Contribuyentes Desvirtuaron (DOF)",
	"nmero_y_fecha_de_oficio_global_de_contribuyentes_que_desvirtuaron_sat": "Oficio Contribuyentes Desvirtuaron (SAT)",
	"nmero_y_fecha_de_oficio_global_de_definitivos_dof":                     "Oficio Definitivos (DOF)",
	"nmero_y_fecha_de_oficio_global_de_definitivos_sat":                     "Oficio Definitivos (SAT)",
	"nmero_y_fecha_de_oficio_global_de_presuncin_dof":                       "Oficio Presunción (DOF)",
	"nmero_y_fecha_de_oficio_global_de_presuncin_sat":                       "Oficio Presunción (SAT)",
	"nmero_y_fecha_de_oficio_global_de_sentencia_favorable_dof":             "Oficio Sentencia Favorable (DOF)",
	"nmero_y_fecha_de_oficio_global_de_sentencia_favorable_sat":             "Oficio Sentencia Favorable (SAT)",
	"nmero_y_fecha_de_oficio_global_definitivo_dof":                         "Oficio Global Definitivo (DOF)",
	"nmero_y_fecha_de_oficio_global_definitivo_sat":                         "Oficio Global Definitivo (SAT)",
	"no":                                       "Número",
	"nombre_del_contribuyente":                 "Nombre del Contribuyente",
	"nombre_denominacin_o_razn_social":         "Nombre / Denominación / Razón Social",
	"nombre_o_razn_social":                     "Nombre o Razón Social",
	"periodo":                                  "Periodo",
	"publicacin_dof_definitivo":                "Publicación DOF Definitivo",
	"publicacin_dof_definitivos":               "Publicación DOF Definitivos",
	"publicacin_dof_desvirtuados":              "Publicación DOF Desvirtuados",
	"publicacin_dof_presuntos":                 "Publicación DOF Presuntos",
	"publicacin_dof_sentencia_favorable":       "Publicación DOF Sentencia Favorable",
	"publicacin_pgina_sat_definitivo":          "Publicación SAT Definitivo",
	"publicacin_pgina_sat_definitivos":         "Publicación SAT Definitivos",
	"publicacin_pgina_sat_desvirtuados":        "Publicación SAT Desvirtuados",
	"publicacin_pgina_sat_presuntos":           "Publicación SAT Presuntos",
	"publicacin_pgina_sat_sentencia_favorable": "Publicación SAT Sentencia Favorable",
	"razn_social":                              "Razón Social",
	"rfc":                                      "RFC",
	"situacin_del_contribuyente":               "Situación del Contribuyente",
	"supuesto":                                 "Supuesto",
	"supuesto_de_cancelacin_csd":               "Supuesto de Cancelación CSD",
	"tipo_de_persona":                          "Tipo de Persona",
	"tipo_persona":                             "Tipo Persona",
	"last_update":                              "Última Actualización",
	"file_name":                                "Archivo Fuente",
	"clasificacion":                            "Clasificación",
	"clasificacionDescription":                 "Descripción",
}
