package main

import (
	"errors"
	_ "fmt"
	_ "html"
	"log/slog"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/periface/cuba/handlers"
	"github.com/periface/cuba/internals/utils"
	"github.com/periface/cuba/views"
)

var renderers = utils.NewRenderers()

func mainHandler(c echo.Context) error {
	rfcQuery := c.QueryParam("rfc")
	component := views.Index(rfcQuery)
	return renderers.Render("Inicio", c, http.StatusOK, component)
}
func main() {
	e := echo.New()

	e.Static("/static", "assets")
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	handlers := handlers.NewMainHandler()
	e.GET("/", mainHandler)
	e.GET("/buscar", handlers.Proveedores.BuscarProveedor)
	if err := e.Start(":1634"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}
