package models

type KeyValue struct {
	Key   string
	Value string
}
type AppSheetsPayload struct {
	Action     string              `json:"Action"`
	Properties map[string]string   `json:"Properties"`
	Rows       []map[string]string `json:"Rows"`
}
type LLMResponse struct {
	Prompt   string
	Response string
}

type BuscarViewModel struct {
	Data              BuscarResponse  `json:"Data"`
	Prompt            string          `json:"Prompt"`
	SearchEngine      SearxngResponse `json:"SearchEngine"`
	SearchEngineClean SearxngResponse `json:"SearchEngineClean"`
	QueryClean        string          `json:"QueryClean"`
	QueryRiesgo       string          `json:"QueryRiesgo"`
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

type SearxngResponse struct {
	Query               string        `json:"query"`
	NumberOfResults     int           `json:"number_of_results"`
	Results             []Result      `json:"results"`
	Answers             []Answer      `json:"answers"`
	Corrections         []interface{} `json:"corrections"`
	Infoboxes           []Infobox     `json:"infoboxes"`
	Suggestions         []interface{} `json:"suggestions"`
	UnresponsiveEngines [][]string    `json:"unresponsive_engines"`
}

type Result struct {
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Thumbnail     *string  `json:"thumbnail"`
	Engine        string   `json:"engine"`
	Template      string   `json:"template"`
	ParsedURL     []string `json:"parsed_url"`
	ImgSrc        string   `json:"img_src"`
	Priority      string   `json:"priority"`
	Engines       []string `json:"engines"`
	Positions     []int    `json:"positions"`
	Score         float64  `json:"score"`
	Category      string   `json:"category"`
	PublishedDate *string  `json:"publishedDate"`

	// Campos opcionales
	IframeSrc *string `json:"iframe_src,omitempty"`
	AudioSrc  *string `json:"audio_src,omitempty"`
	Pubdate   *string `json:"pubdate,omitempty"`
	Length    *int    `json:"length,omitempty"`
	Views     *string `json:"views,omitempty"`
	Author    *string `json:"author,omitempty"`
	Metadata  *string `json:"metadata,omitempty"`

	OpenGroup  *bool `json:"open_group,omitempty"`
	CloseGroup *bool `json:"close_group,omitempty"`
}

type Answer struct {
	URL       string   `json:"url"`
	Engine    string   `json:"engine"`
	ParsedURL []string `json:"parsed_url"`
	Template  string   `json:"template"`
	Answer    string   `json:"answer"`
}

type Infobox struct {
	Infobox    string             `json:"infobox"`
	ID         string             `json:"id"`
	Content    string             `json:"content"`
	ImgSrc     *string            `json:"img_src"`
	URLs       []InfoboxURL       `json:"urls"`
	Attributes []InfoboxAttribute `json:"attributes"`

	Engine    string      `json:"engine"`
	URL       *string     `json:"url"`
	Template  string      `json:"template"`
	ParsedURL interface{} `json:"parsed_url"`

	Title         string      `json:"title"`
	Thumbnail     string      `json:"thumbnail"`
	Priority      string      `json:"priority"`
	Engines       []string    `json:"engines"`
	Positions     interface{} `json:"positions"`
	Score         float64     `json:"score"`
	Category      string      `json:"category"`
	PublishedDate *string     `json:"publishedDate"`
}

type InfoboxURL struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Official bool   `json:"official,omitempty"`
}

type InfoboxAttribute struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Entity string `json:"entity"`
}
