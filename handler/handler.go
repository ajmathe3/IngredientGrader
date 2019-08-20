package handler

/* Package that contains the handler functions for the router */

import (
	"IngredientGrader/data"
	"IngredientGrader/server"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var db = data.DB

// HandleFood is the page handler for the Search Food (/food) page
/* The page requires a get variable named barcode.
   The page should return alerts if one of the following conditions
   have been met:
	  - The barcode cannot be parsed into an integer
	  - The food does not exist within the database
	  - The food does exist within the database, but has ingredients
		with no assigned grades
*/
func HandleFood(w http.ResponseWriter, r *http.Request) {
	// This page can only handle get requests
	if r.Method != "GET" {
		return
	}
	// Load the layout
	t, _ := template.ParseFiles("public/templates/layout.html")
	// Now check if values can be parsed from the query string
	vals, ok := r.URL.Query()["barcode"]

	var content data.Content
	c := &content
	c.Source = "HandleFood"
	// If ok is false, or length vals is 0, the page is being loaded
	if !ok || len(vals) == 0 {
		templ, _ := template.ParseFiles("public/templates/food.html")
		t.AddParseTree("content", templ.Tree)
		c.Success = false
		t.ExecuteTemplate(w, "layout", c)
		return
	}

	bar := string(vals[0])

	// If this point is reached, then a get request with barcode defined was submitted

	// Check if the barcode is numeric
	_, isnum := strconv.Atoi(bar)
	if isnum != nil {
		c.AddError("Barcode must be an integer")
		templ, _ := template.ParseFiles("public/templates/food.html")
		t.AddParseTree("content", templ.Tree)
		c.Success = false
		t.ExecuteTemplate(w, "layout", c)
		return
	}

	// Retrieve food from the DB
	tempFood, exists := server.GetFood(bar)

	// Check if it exists, and if food is missing ingredients
	if !exists {
		c.AddError(fmt.Sprintf("There is no food associated with barcode: %s", bar))
	}
	if tempFood.Grade == "missing" {
		c.AddError(fmt.Sprintf("%s is missing graded ingredients", tempFood.Name))
	}

	// From here, the barcode is valid
	list := strings.Split(tempFood.Ingredients, ",")
	for i := 0; i < len(list); i++ {
		// Format the str to be usable by server.GetIngredient
		str := strings.ToLower(list[i])
		str = strings.Trim(str, " ")
		// Retreive the ingredient
		ingredient := server.GetIngredient(str)
		// Add it to the content struct
		c.AddIngredient(ingredient)
	}
	c.PageFood = tempFood
	if !c.HasErrors() {
		c.Success = true
	}

	templ, _ := template.ParseFiles("public/templates/food.html")
	t.AddParseTree("content", templ.Tree)
	t.ExecuteTemplate(w, "layout", c)
}

// MakeFood is the handler function for /food route. Allows for the creation
/* of a Food struct from checked Form data. The data of the Food struct will
   then be added to the database
*/
func MakeFood(w http.ResponseWriter, r *http.Request) {
	// Create content struct to hold content
	var content data.Content
	// Pointer to content struct for method use
	c := &content
	c.Source = "MakeFood"

	// Load the layout template
	t, _ := template.ParseFiles("public/templates/layout.html")
	// If code reaches here, the page is loading
	if r.Method == "GET" {
		templ, _ := template.ParseFiles("public/templates/makeFood.html")
		t.AddParseTree("content", templ.Tree)
		t.ExecuteTemplate(w, "layout", c)
		return
	}
	// If the code executes this, then an illegal http method was used
	if r.Method != "POST" {
		return
	}

	// If the code has reached here, process form data from the http POST
	r.ParseForm()
	barcode := r.Form["barcode"][0]
	name := r.Form["name"][0]
	ingred := strings.ToLower(r.Form["ingred"][0])

	// Check if the food searched for exists. Duplicate foods cannot be made
	_, foodExists := server.GetFood(barcode)
	if foodExists {
		c.AddError(fmt.Sprintf("Food with barcode: %s already exists", barcode))
	}

	// Check if any of the form's fields were submitted empty
	if len(barcode) == 0 {
		c.AddError("Barcode Field cannot be empty")
	}
	if len(name) == 0 {
		c.AddError("Name Field cannot be empty")
	}
	if len(ingred) == 0 {
		c.AddError("Name Field cannot be empty")
	}

	// Calculate Grade
	list := strings.Split(ingred, ",")
	var total int
	var allFound = true

	for i := 0; i < len(list); i++ {
		oneIngred := list[i]
		oneIngred = strings.Trim(oneIngred, " ")
		oneIngred = strings.ToLower(oneIngred)

		tempIngred := server.GetIngredient(oneIngred)
		if tempIngred.Grade == -10 {
			total = 0
			allFound = false
			server.RecordMissingIngredient(tempIngred.Name)
		} else {
			total += tempIngred.Grade
		}
	}
	var avgGrade = float64(total) / float64(len(list))
	var grade = "missing"
	if allFound {
		if avgGrade < -3 {
			grade = "very bad"
		} else if avgGrade < -1 {
			grade = "bad"
		} else if avgGrade < 1 {
			grade = "neutral"
		} else if avgGrade < 3 {
			grade = "good"
		} else if avgGrade < 5 {
			grade = "very good"
		}
	}

	if !c.HasErrors() {
		server.CreateFood(barcode, name, ingred, grade, avgGrade)
		c.Success = true
		c.PageFood = data.Food{Barcode: barcode, Name: name, Ingredients: ingred, Grade: grade, NumGrade: avgGrade}
		// create food and display success template
	}
	templ, _ := template.ParseFiles("public/templates/makeFood.html")
	t.AddParseTree("content", templ.Tree)
	t.ExecuteTemplate(w, "layout", c)
}

// MakeIngredient is the handler for the admin page that allows for the creation
/* of ingredients and add them to the database
 */
func MakeIngredient(w http.ResponseWriter, r *http.Request) {
	// Content struct for holding data, pointer for using methods
	var content data.Content
	c := &content
	c.Source = "MakeIngredient"

	t, _ := template.ParseFiles("public/templates/layout.html")
	if r.Method == "GET" {
		templ, _ := template.ParseFiles("public/templates/makeIngredient.html")
		t.AddParseTree("content", templ.Tree)
		t.ExecuteTemplate(w, "layout", c)
		return
	}
	if r.Method != "POST" {
		return
	}

	// Parse the frontend for the name and grade entered by the user
	r.ParseForm()
	name := r.Form["name"][0]
	grade := r.Form["grade"][0]

	//formatting stuff
	name = strings.Trim(name, " ")
	name = strings.ToLower(name)
	grade = strings.Trim(grade, " ")

	// Check if the name and grade fields are empty
	if len(name) == 0 {
		c.AddError("Name Field cannot be empty|")
	}
	if len(grade) == 0 {
		c.AddError("Grade Field cannot be empty|")
	}

	// Create ingredient struct
	g, e := strconv.Atoi(grade)
	if e != nil {
		c.AddError("Grade could not be parsed. Check to make sure it is a number")
	}

	// Check if the grade is in range
	if g < -5 || g > 5 || g%1 != 0 {
		c.AddError("The grade must be an integer between -5 and 5, inclusive")
	}

	// Check if the ingredient already exists
	in := server.GetIngredient(name)

	if in.Grade != -10 {
		c.AddError(fmt.Sprintf("Ingredient %s already exists", name))
	}

	if !c.HasErrors() {
		server.CreateIngredient(name, g)
		var temp = data.Ingredient{Name: name, Grade: g}
		c.AddIngredient(temp)
		c.Success = true
	}
	templ, _ := template.ParseFiles("public/templates/makeIngredient.html")
	t.AddParseTree("content", templ.Tree)
	t.ExecuteTemplate(w, "layout", c)
}

// HandleAbout is a function that displays the About page
func HandleAbout(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("public/templates/layout.html")
	if err != nil {
		log.Println(err)
		return
	}
	inner, err := template.ParseFiles("public/templates/about.html")
	if err != nil {
		log.Println(err)
		return
	}
	t.AddParseTree("content", inner.Tree)
	var content data.Content
	c := &content
	c.Source = "HandleAbout"
	c.Success = true
	fmt.Println(t.DefinedTemplates())
	t.ExecuteTemplate(w, "layout", c)
}

// HandleLanding loads the landing page when the domain is visited */
func HandleLanding(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("public/templates/layout.html")
	templ, _ := template.ParseFiles("public/templates/index2.html")
	t.AddParseTree("content", templ.Tree)

	t.ExecuteTemplate(w, "layout", nil)
}

// HandlePublic serves public assets of the website to allow for the
/* site to access them. Examples of this are JS files, CSS files, HTML
   templates, and images
*/
func HandlePublic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dir := vars["dir"]
	file := vars["file"]

	url := fmt.Sprintf("public/%s/%s", dir, file)
	http.ServeFile(w, r, url)
}

// Foo implements ServeHTTP, so it can be used as a Handler
type Foo int

// ServeHTTP is part of an interface implemented by foo to allow
// for 404 redirects
func (m Foo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("public/templates/notFound.html")
	t.Execute(w, nil)
}
