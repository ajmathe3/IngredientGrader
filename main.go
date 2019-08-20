package main

import (
	"IngredientGrader/data"
	"IngredientGrader/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	routes.InitRoutes()
	router := routes.Router

	// Connect to the database
	if dbError := data.Init(); dbError != nil {
		log.Fatalln(dbError)
	}

	// Serve the website
	if ServeErr := http.ListenAndServe(":8000", router); ServeErr != nil {
		log.Fatalln(ServeErr)
	}
}

func testRouter(w http.ResponseWriter, t *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

/* To Do
Add restrictions that limit who can use admin pages and api
SQL Injection protection
Implement api for public consumption
Periodically extract all foods with grade 'missing' to recalculate grade, OR do so every time an
	ingredient is added to the database
When printing to a table for /food, make the header printing better
Admin functionality to alter database without needing to check DB itself

Edge Cases
	Fix the number of sig figs when printing the grade to a table
*/
