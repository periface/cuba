package searchxng

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/utils"
)

type SearXNGClient struct {
	httpTools *utils.HttpTools
}

func NewSearXNGClient(serverPath string) *SearXNGClient {
	return &SearXNGClient{
		httpTools: utils.NewHttpTools(serverPath),
	}
}

func (c *SearXNGClient) BasicSearch(query string) (models.SearxngResponse, error) {
	var resp models.SearxngResponse
	route := "/search?q=" + url.QueryEscape(query) + "&format=json"
	err := c.httpTools.RunHttp(http.MethodGet, route, nil, &resp)

	if err != nil {
		return models.SearxngResponse{}, err
	}

	fmt.Println(resp.Query)
	fmt.Println(len(resp.Results))
	return resp, nil
}

func (c *SearXNGClient) AdvancedSearch(query string,
	categories []string,
	engines []string, limitResults int) (models.SearxngResponse, error) {
	var resp models.SearxngResponse
	route := "/search?q=" + url.QueryEscape(query) + "&format=json"
	formData := url.Values{}
	formData.Set("q", query)
	formData.Set("format", "json")
	formData.Set("language", "es-MX")
	formData.Set("number_of_results", string(limitResults))

	if len(categories) > 0 {
		formData.Set("categories", strings.Join(categories, ","))
	}
	if len(engines) > 0 {
		formData.Set("engines", strings.Join(engines, ","))
	}

	// Convertimos el formulario a bytes
	bodyBytes := []byte(formData.Encode())

	err := c.httpTools.RunHttp(http.MethodPost, route, bodyBytes, &resp)

	if err != nil {
		return models.SearxngResponse{}, err
	}

	fmt.Println("Query procesado:", resp.Query)
	fmt.Println("Resultados encontrados:", len(resp.Results))
	return resp, nil
}
