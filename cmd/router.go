package cmd

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/internal/controllers/rest"
)

// newRouter builds the complete HTTP route table for the REST API and
// documentation endpoints. chain is applied to every route registered here.
func newRouter(restHandler *rest.Handler, chain alice.Chain) http.Handler {
	r := httprouter.New()

	r.Handler(http.MethodGet, "/ping", chain.ThenFunc(restHandler.Ping))

	r.Handler(http.MethodPost, "/rest/v1/sheets", chain.ThenFunc(restHandler.CreateSheet))
	r.Handler(http.MethodGet, "/rest/v1/sheets/:id", chain.ThenFunc(restHandler.GetSheetById))
	r.Handler(http.MethodGet, "/rest/v1/sheets", chain.ThenFunc(restHandler.GetAllSheets))
	r.Handler(http.MethodPut, "/rest/v1/sheets/:id", chain.ThenFunc(restHandler.UpdateSheetById))
	r.Handler(http.MethodDelete, "/rest/v1/sheets/:id", chain.ThenFunc(restHandler.DeleteSheetById))

	r.Handler(http.MethodPost, "/rest/v1/roasters", chain.ThenFunc(restHandler.CreateRoaster))
	r.Handler(http.MethodGet, "/rest/v1/roasters/:id", chain.ThenFunc(restHandler.GetRoasterById))
	r.Handler(http.MethodGet, "/rest/v1/roasters", chain.ThenFunc(restHandler.GetAllRoasters))
	r.Handler(http.MethodPut, "/rest/v1/roasters/:id", chain.ThenFunc(restHandler.UpdateRoasterById))
	r.Handler(http.MethodDelete, "/rest/v1/roasters/:id", chain.ThenFunc(restHandler.DeleteRoasterById))

	r.Handler(http.MethodPost, "/rest/v1/beans", chain.ThenFunc(restHandler.CreateBeans))
	r.Handler(http.MethodGet, "/rest/v1/beans/:id", chain.ThenFunc(restHandler.GetBeansById))
	r.Handler(http.MethodGet, "/rest/v1/beans", chain.ThenFunc(restHandler.GetAllBeans))
	r.Handler(http.MethodPut, "/rest/v1/beans/:id", chain.ThenFunc(restHandler.UpdateBeanById))
	r.Handler(http.MethodDelete, "/rest/v1/beans/:id", chain.ThenFunc(restHandler.DeleteBeansById))

	r.Handler(http.MethodPost, "/rest/v1/shots", chain.ThenFunc(restHandler.CreateShot))
	r.Handler(http.MethodGet, "/rest/v1/shots/:id", chain.ThenFunc(restHandler.GetShotById))
	r.Handler(http.MethodGet, "/rest/v1/shots", chain.ThenFunc(restHandler.GetAllShots))
	r.Handler(http.MethodPut, "/rest/v1/shots/:id", chain.ThenFunc(restHandler.UpdateShotById))
	r.Handler(http.MethodDelete, "/rest/v1/shots/:id", chain.ThenFunc(restHandler.DeleteShotById))

	redocOpts := middleware.RedocOpts{Path: "redoc", SpecURL: "swagger.json"}
	swaggerUiOpts := middleware.SwaggerUIOpts{Path: "swagger", SpecURL: "swagger.json"}
	r.Handler(http.MethodGet, "/redoc", middleware.Redoc(redocOpts, nil))
	r.Handler(http.MethodGet, "/swagger", middleware.SwaggerUI(swaggerUiOpts, nil))
	r.Handler(http.MethodGet, "/swagger.json", chain.ThenFunc(restHandler.Swagger))

	return r
}
