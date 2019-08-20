package routes

import (
	"IngredientGrader/handler"

	"github.com/gorilla/mux"
)

// Router is the router for the site
var Router *mux.Router

// InitRoutes initializes the routers for the web server
// for this site. This must be called first OR after a
// connection to the database has been established
func InitRoutes() {
	Router = mux.NewRouter()
	// Attach Routes

	// Routes for public webpages
	Router.HandleFunc("/food", handler.HandleFood)
	Router.HandleFunc("/about", handler.HandleAbout)
	Router.HandleFunc("/", handler.HandleLanding)

	// Routes for Admin Pages
	Router.HandleFunc("/admin/food/create", handler.MakeFood)
	Router.HandleFunc("/admin/ingredient/create", handler.MakeIngredient)

	// Routes for misc
	Router.HandleFunc("/public/{dir}/{file}/", handler.HandlePublic)
	// Route for 404 Not Found
	var i handler.Foo
	Router.NotFoundHandler = i
}
