package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"runtime"

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
