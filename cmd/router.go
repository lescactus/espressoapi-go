package cmd

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/internal/controllers/rest"
	"github.com/lescactus/espressoapi-go/internal/controllers/web"
)

// newRouter builds the complete HTTP route table for the REST API, the web
// UI, and documentation endpoints. chain is applied to every route
// registered here.
func newRouter(restHandler *rest.Handler, webHandler *web.Handler, chain alice.Chain) http.Handler {
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
	r.Handler(http.MethodGet, "/rest/v1/sheets/:id/shots", chain.ThenFunc(restHandler.GetShotsBySheetId))

	redocOpts := middleware.RedocOpts{Path: "redoc", SpecURL: "swagger.json"}
	swaggerUiOpts := middleware.SwaggerUIOpts{Path: "swagger", SpecURL: "swagger.json"}
	r.Handler(http.MethodGet, "/redoc", middleware.Redoc(redocOpts, nil))
	r.Handler(http.MethodGet, "/swagger", middleware.SwaggerUI(swaggerUiOpts, nil))
	r.Handler(http.MethodGet, "/swagger.json", chain.ThenFunc(restHandler.Swagger))

	// Web UI
	r.Handler(http.MethodGet, "/", chain.ThenFunc(webHandler.Home))

	r.Handler(http.MethodGet, "/sheets", chain.ThenFunc(webHandler.ListSheets))
	r.Handler(http.MethodGet, "/sheets/add", chain.ThenFunc(webHandler.AddSheetForm))
	r.Handler(http.MethodPost, "/sheets/add", chain.ThenFunc(webHandler.CreateSheet))
	r.Handler(http.MethodGet, "/sheets/get/:id", chain.ThenFunc(webHandler.GetSheet))
	r.Handler(http.MethodGet, "/sheets/update/:id", chain.ThenFunc(webHandler.EditSheetForm))
	r.Handler(http.MethodPut, "/sheets/update/:id", chain.ThenFunc(webHandler.UpdateSheet))
	r.Handler(http.MethodDelete, "/sheets/delete/:id", chain.ThenFunc(webHandler.DeleteSheet))

	r.Handler(http.MethodGet, "/roasters", chain.ThenFunc(webHandler.ListRoasters))
	r.Handler(http.MethodGet, "/roasters/add", chain.ThenFunc(webHandler.AddRoasterForm))
	r.Handler(http.MethodPost, "/roasters/add", chain.ThenFunc(webHandler.CreateRoaster))
	r.Handler(http.MethodGet, "/roasters/get/:id", chain.ThenFunc(webHandler.GetRoaster))
	r.Handler(http.MethodGet, "/roasters/update/:id", chain.ThenFunc(webHandler.EditRoasterForm))
	r.Handler(http.MethodPut, "/roasters/update/:id", chain.ThenFunc(webHandler.UpdateRoaster))
	r.Handler(http.MethodDelete, "/roasters/delete/:id", chain.ThenFunc(webHandler.DeleteRoaster))

	r.Handler(http.MethodGet, "/beans", chain.ThenFunc(webHandler.ListBeans))
	r.Handler(http.MethodGet, "/beans/add", chain.ThenFunc(webHandler.AddBeanForm))
	r.Handler(http.MethodPost, "/beans/add", chain.ThenFunc(webHandler.CreateBean))
	r.Handler(http.MethodGet, "/beans/get/:id", chain.ThenFunc(webHandler.GetBean))
	r.Handler(http.MethodGet, "/beans/update/:id", chain.ThenFunc(webHandler.EditBeanForm))
	r.Handler(http.MethodPut, "/beans/update/:id", chain.ThenFunc(webHandler.UpdateBean))
	r.Handler(http.MethodDelete, "/beans/delete/:id", chain.ThenFunc(webHandler.DeleteBean))

	return r
}
