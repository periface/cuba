package appsheets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/periface/cuba/internals/models"
	"github.com/periface/cuba/internals/utils"
)

type Appsheets struct {
	appId     string
	appSecret string
	baseUrl   string
}

const ORIGINAL_URL = "https://www.appsheet.com/api/v2/apps/${APPSHEETSID}/tables/${TABLENAME}/Action?applicationAccessKey=${APPSHEETSSECRET}"

func NewAppsheets() (*Appsheets, error) {

	APPSHEETSID, err := utils.GetEnvVariable("APPSHEETSID")
	if err != nil {
		fmt.Println("Error getting APPSHEETSID:", err)
		return nil, err
	}
	APPSHEETSSECRET, err := utils.GetEnvVariable("APPSHEETSSECRET")
	if err != nil {
		fmt.Println("Error getting APPSHEETSSECRET:", err)
		return nil, err
	}

	API_URL := ORIGINAL_URL
	API_URL = buildApiUrl(API_URL, APPSHEETSID, APPSHEETSSECRET)
	return &Appsheets{
		appId:     APPSHEETSID,
		appSecret: APPSHEETSSECRET,
		baseUrl:   API_URL,
	}, nil
}
func altQueryScape(originalString string) string {
	return strings.ReplaceAll(originalString, " ", "%20")
}
func buildApiUrl(apiUrl string, key string, secret string) string {
	apiUrl = strings.ReplaceAll(apiUrl, "${APPSHEETSID}", key)
	apiUrl = strings.ReplaceAll(apiUrl, "${APPSHEETSSECRET}", secret)
	return apiUrl
}

func (as *Appsheets) SearchIn(apiKey string, apiSecret string, tableName string, input models.AppSheetsPayload) ([]map[string]string, error) {
	endPoint := buildApiUrl(ORIGINAL_URL, apiKey, apiSecret)
	tableName = altQueryScape(tableName)
	endPoint = strings.ReplaceAll(endPoint, "${TABLENAME}", tableName)
	body, err := json.Marshal(input)
	if err != nil {
		log.Fatal("Error en search")
		log.Fatal(err)
	}

	return RunHttpQuery(endPoint, body)
}
func (as *Appsheets) Search(tableName string, input models.AppSheetsPayload) ([]map[string]string, error) {
	tableName = altQueryScape(tableName)
	endPoint := strings.ReplaceAll(as.baseUrl, "${TABLENAME}", tableName)
	body, err := json.Marshal(input)
	if err != nil {
		log.Fatal("Error en search")
		log.Fatal(err)
	}

	return RunHttpQuery(endPoint, body)
}
func (as *Appsheets) GetTable(tableName string) ([]map[string]string, error) {
	tableName = altQueryScape(tableName)
	endPoint := strings.ReplaceAll(as.baseUrl, "${TABLENAME}", tableName)
	body := []byte(`{"Action": "Find"}`)
	return RunHttpQuery(endPoint, body)
}

func (as *Appsheets) Insert(tableName string,
	input models.AppSheetsPayload) ([]map[string]string, error) {

	tableName = altQueryScape(tableName)
	endPoint := strings.ReplaceAll(as.baseUrl, "${TABLENAME}", tableName)
	body, err := json.Marshal(input)
	if err != nil {
		log.Fatal("Error en insert")
		log.Fatal(err)
	}

	return RunHttpQuery(endPoint, body)
}

func RunHttpQuery(url string, body []byte) ([]map[string]string, error) {
	time.Sleep(1 * time.Second)
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	if err != nil {
		fmt.Println("Request Error")
		return nil, err
	}

	httpResponse, err := client.Do(request)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer httpResponse.Body.Close()
	resBody, err := io.ReadAll(httpResponse.Body)

	fmt.Println("===========")
	fmt.Println(url)
	fmt.Println("===========")
	if err != nil {
		fmt.Println("Error leyendo respuesta:", err)
		return nil, err
	}
	// Parse the JSON response
	var response []map[string]string
	if string(resBody) == "" {
		return nil, fmt.Errorf("No hay datos en la respuesta")
	}
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	return response, nil
}
