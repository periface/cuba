package prompts

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/services/appsheets"
	"github.com/periface/cuba/internals/services/proveedores"
	"github.com/periface/cuba/internals/utils"
)

func listToStringList(fallBackText string, input []map[string]string) string {
	if len(input) == 0 {
		return fallBackText
	}

	var sb strings.Builder
	for i, item := range input {
		sb.WriteString(fmt.Sprintf("%d:\n", i+1))
		for key, value := range item {
			sb.WriteString(fmt.Sprintf(" - %s: %s\n", key, value))
		}
	}
	return sb.String()
}

func makeParagraph(title string, onEmptyString string, data []map[string]string) string {

	var sb strings.Builder
	sb.WriteString(fmt.Sprintln(title))
	stringList := listToStringList(onEmptyString, data)
	sb.WriteString(fmt.Sprintln(stringList))
	return sb.String()
}
func buildProveedoresPrompt(rfc string, buscarResponse models.BuscarResponse) string {

	contratosStr := makeParagraph("**Contratos en Gobierno:**", "Sin datos de contratos", buscarResponse.ContratosEncontrados)
	observacionesStr := makeParagraph("**Observaciones SAT**", "No hay observaciones en el SAT", buscarResponse.ObservacionesSat)
	empleadosEncontradosStr := makeParagraph("**Empleados con Mismo RFC (plantilla laboral gubernamental):**", "Sin datos de empleados como proveedores", buscarResponse.EmpleadosEncontrados)
	informacionDelProveedorStr := makeParagraph("**Información DGCyOP:**", "DGCyOP no tiene datos del proveedor", buscarResponse.InformacionDelProveedor)
	representantesLegalesStr := makeParagraph("**Representantes Legales en DGCyOP:**", "No se encontraron representantes legales", buscarResponse.RepresentantesLegales)

	// Crear prompt claro para el agente
	prompt := fmt.Sprintf(`**Reporte de Integridad de Proveedores - Secretaría de Administración del Estado de Tamaulipas**
---
**Objetivo:** Evaluar riesgos y posibles conflictos de interés de un proveedor para proteger la integridad y la imagen del Gobierno Estatal de Tamaulipas. Este reporte está diseñado para la toma de decisiones directivas.

**Datos de Entrada del Proveedor:**
- %s
- %s
- %s
- %s
- %s
- %s

---

### **Instrucciones para la Generación del Reporte:**

1.  **Investigación Externa (Web):**
    * **Búsqueda Principal:** Inicia la búsqueda exhaustiva con la **razón social** del proveedor. Complementa con el **RFC "%s"** y los **nombres completos de los representantes legales**.
    * **Palabras Clave de Riesgo:** Incluye siempre: "corrupción", "lavado de dinero", "evasión fiscal", "fraude", "irregularidades", "polémica", "escándalo", "investigación", "denuncia", "demandas", "sanciones", "conflicto de interés", "soborno", "inhabilitado", "nepotismo", "desvío de recursos", "peculado", "enriquecimiento ilícito", "conflicto de intereses (ex-funcionario)", "vinculación con gobierno anterior".
    * **Fuentes Prioritarias:** Prioriza **noticias de investigación, comunicados oficiales de fiscalías/contralorías, sentencias, reportes de auditoría y listas de inhabilitados**. Incluye URLs y, si aplica, descripciones de imágenes/logos relevantes. Considera también archivos de prensa históricos y bases de datos periodísticas.
    * **Manejo de Coincidencias (RFC/Nombre):** Si los hallazgos web para el RFC o nombre corresponden a **empleados de gobierno** (y no al proveedor), identifica solo su **nombre completo y la dependencia donde trabaja**. Si hay hallazgos sobre el proveedor, priorízalos.
    * **Investigación de Personas Físicas Relacionadas (PRIORITARIO):**
        * Para cada **representante legal del proveedor**, y si el **RFC del proveedor corresponde a una persona física**, realiza una búsqueda web específica sobre ellos.
        * Utiliza sus **nombres completos** y, si está disponible, su **RFC personal** (incluso si no está en las bases de datos gubernamentales actuales).
        * Aplica las mismas "Palabras Clave de Riesgo" para esta búsqueda.
        * **Objetivo:** Identificar cualquier controversia, investigación, acusación (incluyendo delitos fiscales), o vínculo con administraciones gubernamentales anteriores, incluso si no hay una declaración oficial de conflicto de interés. Si no encuentras controversias, de igual forma **genera un perfil conciso de la persona con la información pública disponible.**

2.  **Análisis Interno y Cruce de Datos:**
    * **Observaciones SAT:** Determina si las observaciones son leves o indican irregularidades graves.
    * **Contratos Gubernamentales:** Evalúa el historial de contratos: número, montos totales (sumar si son varios), y la naturaleza de los servicios/productos.
    * **Conflictos de Interés (RFC y Representantes Legales):**
        * **RFC Coincidente:** Si "Empleados con Mismo RFC" contiene datos, reporta el **nombre del empleado y su dependencia**. Si es un directivo o funcionario, menciónalo.
        * **Representante Legal en Gobierno Actual:** Si algún "Representante Legal" es o fue empleado/funcionario de gobierno en la **administración actual**, identifica su **nombre, puesto y dependencia**.
        * **Representante Legal con Historial Gubernamental (Anterior):** Si la investigación web reveló que un "Representante Legal" fue empleado/funcionario de gobierno en **administraciones anteriores** y existen hallazgos de riesgo o controversias asociadas, detállalos aquí.

---

### **Estructura del Reporte Final (Máximo 3 Párrafos para el Cuerpo Principal):**

**a) Descripción General:**
    * Actividades principales del proveedor.
    * Resumen conciso de observaciones SAT (máximo 3 líneas).
    * Síntesis del historial de contratos gubernamentales (ej. "Ha celebrado X contratos por un total de Y pesos, principalmente para Z").

**b) Hallazgos Críticos y Riesgos:**
    * **Hallazgos Web del Proveedor:** Reporta las controversias, investigaciones o escándalos vinculados directamente al proveedor como entidad, citando fuentes. Si no hay, indica: "No se encontraron referencias públicas adversas relevantes para el proveedor como entidad en fuentes de riesgo."
    * **Perfiles de Personas Físicas Relacionadas y Posibles Controversias (PRIORITARIO):**
        * Para cada representante legal investigado (y la persona física si el RFC del proveedor corresponde a una):
            * **Perfil:** "Se generó un perfil público para **[Nombre Completo de la Persona]**. [Breve descripción de su actividad conocida o historial profesional público, si aplica]."
            * **Controversias/Vínculos:** "Para **[Nombre Completo de la Persona]**, se encontraron los siguientes hallazgos relevantes en fuentes web: [Detallar controversias, investigaciones, acusaciones (incluyendo delitos fiscales), o vínculos con administraciones gubernamentales anteriores. Citar URLs]. Si no hay controversias relevantes, indicar: "No se encontraron controversias o hallazgos públicos adversos de riesgo relevantes asociados a esta persona."
    * **Conflictos de Interés Detectados:**
        * Si se identificó un empleado de gobierno con el mismo RFC del proveedor: "Se encontró una coincidencia de RFC con un empleado de gobierno: **[Nombre del Empleado]** en **[Dependencia]**. Esto se informa por transparencia, sin vínculo directo probado con el proveedor en esta administración."
        * Si un representante legal es o fue empleado/funcionario de gobierno **en la administración actual**: "Uno de los representantes legales, **[Nombre del Representante Legal]**, es/fue **[Puesto y Dependencia]** en la administración actual."
        * Si un representante legal fue empleado/funcionario de gobierno en **administraciones anteriores** y existen hallazgos web de riesgo o controversias asociadas: "Se identificó que uno de los representantes legales, **[Nombre del Representante Legal]**, fue **[Puesto y Dependencia, si se conoce]** en administraciones gubernamentales anteriores. La investigación web sugiere posibles controversias/irregularidades vinculadas a su persona durante o después de su periodo en el servicio público: [Mencionar brevemente la naturaleza de los hallazgos]."

**c) Evaluación y Recomendación Directiva:**
    * **Nivel de Riesgo:** Clasifica el riesgo global como "Nulo", "Bajo", "Moderado" o "Alto", justificando brevemente, considerando especialmente los hallazgos de personas físicas y sus posibles vínculos o controversias, incluso si no hay confirmación oficial.
    * **Recomendación:** Proporciona una de las siguientes:
        * **APROBADO:** Si no hay riesgos significativos.
        * **APROBADO CON OBSERVACIONES:** Si existen puntos a considerar (especifícalos, incluyendo cualquier controversia de personas físicas que requiera seguimiento o mayor escrutinio).
        * **RECHAZADO:** Si los riesgos son inaceptables para la contratación.

---

### **Anexos:**

* **Enlaces de Evidencia:** Lista de URLs de todas las fuentes web consultadas, diferenciando entre las del proveedor y las de las personas físicas.
* **Imágenes/Fotos:** Incluye imágenes relevantes si se encontraron.

---

### **Consideraciones Adicionales:**

* **Datos No Disponibles:** Si alguna sección de los "Datos de Entrada" no aplica o está vacía, ignórala en el análisis o indícalo como "No aplica" en el reporte final.
* **Juicio Crítico:** El sistema debe aplicar un juicio crítico para interpretar los hallazgos web, especialmente los no oficiales, y presentarlos de manera objetiva, señalando si son acusaciones o investigaciones no concluyentes, pero sin omitirlos debido a su potencial relevancia para la integridad.
`,
		rfc,
		informacionDelProveedorStr,
		observacionesStr,
		contratosStr,
		empleadosEncontradosStr,
		representantesLegalesStr,
		// Se repite el RFC aquí para la búsqueda web
		rfc)
	return prompt
}
func AnalisisDeProveedoresPrompt(rfc string, appsheetsInstance appsheets.Appsheets) string {
	proveedorInfo := fetchProveedorInfo(rfc, &appsheetsInstance)
	prompt := buildProveedoresPrompt(rfc, proveedorInfo)
	return prompt
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
