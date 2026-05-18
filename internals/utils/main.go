package utils

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/periface/cuba/views"
)

func get_os() string {
	return runtime.GOOS
}

func IsLinux() bool {
	return get_os() == "linux"
}
func IsWindows() bool {
	return get_os() == "windows"
}
func IsMac() bool {
	return get_os() == "darwin"
}
func GetEnvVariable(key string) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", err
	}
	return os.Getenv(key), nil
}

func ReadCsvFile(fileName string) ([][]string, error) {

	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", fileName, err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ',' // Asegurarse de que el separador sea coma
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file %s: %w", fileName, err)
	}

	return rows, nil
}

type Renderers struct {
}

func NewRenderers() *Renderers {
	return &Renderers{}
}

func (r *Renderers) Render(title string, ctx echo.Context, statusCode int, t templ.Component) error {
	page := views.Layout(title, t)
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if err := page.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}
	return ctx.HTML(statusCode, buf.String())
}
func (r *Renderers) RenderNoLayout(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}
	return ctx.HTML(statusCode, buf.String())
}

type HttpTools struct {
	baseUrl string
	client  *http.Client
}

func NewHttpTools(baseUrl string) *HttpTools {
	return &HttpTools{
		baseUrl: baseUrl,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}
func (h *HttpTools) RunHttp(
	method string,
	route string,
	body []byte,
	target any,
) error {

	time.Sleep(1 * time.Second)

	url := h.baseUrl + route

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	fmt.Println("========== REQUEST ==========")
	fmt.Println("METHOD:", method)
	fmt.Println("URL:", url)
	fmt.Println("BODY:", string(body))
	fmt.Println("=============================")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	fmt.Println("========== RESPONSE ==========")
	fmt.Println("STATUS:", resp.Status)
	fmt.Println("HEADERS:", resp.Header)
	fmt.Println("BODY:")
	fmt.Println(string(resBody))
	fmt.Println("==============================")

	if resp.StatusCode >= 400 {
		return fmt.Errorf(
			"http error %d: %s",
			resp.StatusCode,
			string(resBody),
		)
	}

	if len(resBody) == 0 {
		return fmt.Errorf("empty response")
	}

	err = json.Unmarshal(resBody, target)
	if err != nil {
		return fmt.Errorf("parsing json: %w", err)
	}

	return nil
}
