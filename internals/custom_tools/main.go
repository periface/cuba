package customtools

import (
	"context"
	"fmt"
	"strings"

	"github.com/periface/cuba/internals/utils"
	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

// GoogleSearchTool implementa la interfaz tools.Tool para la API de Búsqueda Personalizada de Google.
type GoogleSearchTool struct {
	// Puedes añadir un cliente de búsqueda aquí si quieres reutilizarlo,
	// o inicializarlo en cada llamada para simplificar.
	// Por simplicidad, lo inicializaremos en cada llamada en este ejemplo.
}

// Name devuelve el nombre de la herramienta.
func (t GoogleSearchTool) Name() string {
	return "Google Search"
}

// Description devuelve una descripción de la herramienta para que el LLM la entienda.
func (t GoogleSearchTool) Description() string {
	return "Útil para buscar información general en Google. La entrada debe ser una consulta de búsqueda simple."
}

// Call ejecuta la búsqueda de Google y devuelve los resultados.
func (t GoogleSearchTool) Call(ctx context.Context, input string) (string, error) {
	apiKey, err := utils.GetEnvVariable("GOOGLESEARCH_API_KEY")

	if err != nil {
		return "", fmt.Errorf("Falta api key")
	}
	cseID, err := utils.GetEnvVariable("GOOGLESENGINE_ID")

	if err != nil {
		return "", fmt.Errorf("Falta search engine id")
	}

	if apiKey == "" || cseID == "" {
		return "", fmt.Errorf("GOOGLE_API_KEY y GOOGLE_CSE_ID deben estar configurados en las variables de entorno")
	}

	// Inicializar el servicio de Búsqueda Personalizada
	svc, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("fallo al crear el servicio de búsqueda: %w", err)
	}


	// Realizar la consulta de búsqueda
	resp, err := svc.Cse.List().Cx(cseID).Q(input).Do()
	if err != nil {
		return "", fmt.Errorf("fallo al ejecutar la consulta de búsqueda: %w", err)
	}

	// Procesar y formatear los resultados
	if len(resp.Items) == 0 {
		return "No se encontraron resultados para tu búsqueda.", nil
	}

	var sb strings.Builder
	for i, item := range resp.Items {
		if i >= 5 { // Limitar a los 5 primeros resultados para evitar respuestas muy largas
			break
		}
		sb.WriteString(fmt.Sprintf("Título: %s\n", item.Title))
		sb.WriteString(fmt.Sprintf("Snippet: %s\n", item.Snippet))
		sb.WriteString(fmt.Sprintf("Enlace: %s\n", item.Link))
		sb.WriteString("---\n") // Separador entre resultados
	}

	return sb.String(), nil
}

// NewGoogleSearchTool crea una nueva instancia de GoogleSearchTool.
func NewGoogleSearchTool() GoogleSearchTool {
	return GoogleSearchTool{}
}
