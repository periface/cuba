package handlers

import (
	"github.com/periface/cuba/handlers/proveedores"
)

type CubaHttpHandlers struct {
	Proveedores *proveedores.ProveedoresHandlers
}

func NewMainHandler() *CubaHttpHandlers {
	proveedores := proveedores.NewProveedoresHandlers()
	return &CubaHttpHandlers{
		Proveedores: proveedores,
	}
}
