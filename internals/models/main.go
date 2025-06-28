package models

type AppSheetsPayload struct {
	Action     string              `json:"Action"`
	Properties map[string]string   `json:"Properties"`
	Rows       []map[string]string `json:"Rows"`
}
type LLMResponse struct {
	Prompt   string
	Response string
}
type BuscarResponse struct {
	ObservacionesSat        []map[string]string
	ContratosEncontrados    []map[string]string
	EmpleadosEncontrados    []map[string]string
	InformacionDelProveedor []map[string]string
	RepresentantesLegales   []map[string]string
	AnalisisPreventivo      string
}
type CondonacionSAT struct {
	ID                                              string `json:"id" csv:"id"`
	AdministracionGeneralResponsableDeLaCancelacion string `json:"administracin_general_responsable_de_la_cancelacion"`
	Ao                                              string `json:"ao"`
	Contribuyente                                   string `json:"contribuyente"`
	Ejercicio                                       string `json:"ejercicio"`
	EntidadFederativa                               string `json:"entidad_federativa"`
	FechaAutorizacion                               string `json:"fecha_de_autorizacin"`
	FechaCancelacion                                string `json:"fecha_de_cancelacin"`
	FechaCancelacionCSD                             string `json:"fecha_de_cancelacin_csd"`
	FechaPublicacion                                string `json:"fecha_de_publicacin"`
	FechaPublicacionConMontoLeyTransparencia        string `json:"fecha_de_publicacin_con_monto_de_acuerdo_a_la_ley_de_transparencia"`
	FechasPrimeraPublicacion                        string `json:"fechas_de_primera_publicacion"`
	Importe                                         string `json:"importe"`
	ImporteCondonado                                string `json:"importe_condonado"`
	Monto                                           string `json:"monto"`
	Motivo                                          string `json:"motivo"`
	MotivoCondonacion                               string `json:"motivo_de_condonacin"`
	NumeroFechaOficioGlobalContribuyentesDof        string `json:"nmero_y_fecha_de_oficio_global_de_contribuyentes_que_desvirtuaron_dof"`
	NumeroFechaOficioGlobalContribuyentesSat        string `json:"nmero_y_fecha_de_oficio_global_de_contribuyentes_que_desvirtuaron_sat"`
	NumeroFechaOficioGlobalDefinitivosDof           string `json:"nmero_y_fecha_de_oficio_global_de_definitivos_dof"`
	NumeroFechaOficioGlobalDefinitivosSat           string `json:"nmero_y_fecha_de_oficio_global_de_definitivos_sat"`
	NumeroFechaOficioGlobalPresuncionDof            string `json:"nmero_y_fecha_de_oficio_global_de_presuncin_dof"`
	NumeroFechaOficioGlobalPresuncionSat            string `json:"nmero_y_fecha_de_oficio_global_de_presuncin_sat"`
	NumeroFechaOficioGlobalSentenciaFavorableDof    string `json:"nmero_y_fecha_de_oficio_global_de_sentencia_favorable_dof"`
	NumeroFechaOficioGlobalSentenciaFavorableSat    string `json:"nmero_y_fecha_de_oficio_global_de_sentencia_favorable_sat"`
	NumeroFechaOficioGlobalDefinitivoDof            string `json:"nmero_y_fecha_de_oficio_global_definitivo_dof"`
	NumeroFechaOficioGlobalDefinitivoSat            string `json:"nmero_y_fecha_de_oficio_global_definitivo_sat"`
	No                                              string `json:"no"`
	NombreContribuyente                             string `json:"nombre_del_contribuyente"`
	NombreDenominacionRazonSocial                   string `json:"nombre_denominacin_o_razn_social"`
	NombreRazonSocial                               string `json:"nombre_o_razn_social"`
	Periodo                                         string `json:"periodo"`
	PublicacionDofDefinitivo                        string `json:"publicacin_dof_definitivo"`
	PublicacionDofDefinitivos                       string `json:"publicacin_dof_definitivos"`
	PublicacionDofDesvirtuados                      string `json:"publicacin_dof_desvirtuados"`
	PublicacionDofPresuntos                         string `json:"publicacin_dof_presuntos"`
	PublicacionDofSentenciaFavorable                string `json:"publicacin_dof_sentencia_favorable"`
	PublicacionPaginaSatDefinitivo                  string `json:"publicacin_pgina_sat_definitivo"`
	PublicacionPaginaSatDefinitivos                 string `json:"publicacin_pgina_sat_definitivos"`
	PublicacionPaginaSatDesvirtuados                string `json:"publicacin_pgina_sat_desvirtuados"`
	PublicacionPaginaSatPresuntos                   string `json:"publicacin_pgina_sat_presuntos"`
	PublicacionPaginaSatSentenciaFavorable          string `json:"publicacin_pgina_sat_sentencia_favorable"`
	RazonSocial                                     string `json:"razn_social"`
	RFC                                             string `json:"rfc"`
	SituacionContribuyente                          string `json:"situacin_del_contribuyente"`
	Supuesto                                        string `json:"supuesto"`
	SupuestoCancelacionCSD                          string `json:"supuesto_de_cancelacin_csd"`
	TipoPersona                                     string `json:"tipo_de_persona"`
	TipoPersona2                                    string `json:"tipo_persona"` // Nota: parece haber un typo en el nombre original
	LastUpdate                                      string `json:"last_update"`
	FileName                                        string `json:"file_name"`
	Clasificacion                                   string `json:"clasificacion"`

	ClasificacionDescription string `json:"clasificacionDescription"`
}
