package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// DB is s typing of sql.DB to allow readability of code
type DB sql.DB

// Food is a typed struct that is used to decode and encode
// data for http requests
type Food struct {
	Barcode     string  `json:"barcode"`
	Name        string  `json:"title"`
	Ingredients string  `json:"ingredients"`
	Grade       string  `json:"grade"`
	NumGrade    float64 `json:"numgrade"`
}

// Ingredient is a typed struct that is used to decode and
// encode data for http requests
type Ingredient struct {
	Name  string `json:"title"`
	Grade int    `json:"grade"`
}

// FoodPage To print out both the Food Data and Ingredient Data on the
// /food page, a specialized struct will be defined for it
type FoodPage struct {
	AFood       Food         `json:"food"`
	Ingredients []Ingredient `json:"ingredients"`
}

// Implements http.Handler to allow for custom 404 error page
type foo int

// Allows for dynamic absolute addresses when testing/deploying
var root = "http://localhost:8000/"

var db *sql.DB

func main() {
	//Router to handle all routings
	router := mux.NewRouter()

	//serving external files
	router.HandleFunc("/js/{fileName}", handleJS)
	router.HandleFunc("images/{fileName}", handleImages)

	//endpoints for web pages
	router.HandleFunc("/food", handleFood)
	router.HandleFunc("/about", handleAbout)
	router.HandleFunc("/", landingFunc)

	//endpoints for admin pages
	router.HandleFunc("/admin/food/create", makeFood)
	router.HandleFunc("/admin/ingredient/create", makeIngredient)

	//endpoints for the api
	router.HandleFunc("/api/food/{bar}", getFood).Methods("GET")
	router.HandleFunc("/api/food/{bar}", updateFood).Methods("POST")
	router.HandleFunc("/api/food/{bar}", deleteFood).Methods("DELETE")
	router.HandleFunc("/api/food", getAllFoods).Methods("GET")
	router.HandleFunc("/api/food", createFood).Methods("POST")
	router.HandleFunc("/api/ingredient", createIngredient).Methods("POST")
	router.HandleFunc("/api/ingredient/{name}", getIngredient).Methods("GET")

	//custom 404 page
	var m foo
	router.NotFoundHandler = m
	// Open and Test DB connection
	openConn()

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatalln(err)
	}
}

/* Opens a connection to the database and checks if the connection is working */
func openConn() error {
	if db == nil {
		tempdb, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/foodstuff")
		if err != nil {
			return err
		}
		db = tempdb
		err = db.Ping()
		if err != nil {
			log.Fatalln(err)
		}
		return nil

	}
	log.Println("db already defined")
	err := db.Ping()
	if err != nil {
		log.Fatalln(err)
	}
	return nil
}

// PUBLIC WEB PAGES

/* Creates a custom 404 error page */
func (m foo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/notFound.html")
	t.Execute(w, nil)
}

/* Handler function for the landing page of the site */
func landingFunc(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index2.html")
	t.Execute(w, nil)
}

/* Retrieves food info from the database. Current implementation writes data to page */
func handleFood(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		vals, ok := r.URL.Query()["barcode"]
		var errors []string
		t, _ := template.ParseFiles("templates/food.html")
		t.Execute(w, nil)

		if ok && len(vals[0]) > 0 {
			// Check whether barcode is numeric
			_, err := strconv.Atoi(string(vals[0]))
			if err != nil {
				errors = append(errors, "barcode must be numerical")
			} else {
				str := fmt.Sprintf("%sapi/food/%s", root, string(vals[0]))
				resp, err := http.Get(str)
				if err != nil {
					log.Println(err)
				}
				defer resp.Body.Close()
				body, erros := ioutil.ReadAll(resp.Body)
				if erros != nil {
					log.Println(erros)
				}

				var tempFood Food
				var foundFood = true
				errMsg := json.Unmarshal(body, &tempFood)
				if errMsg != nil {
					foundFood = false
				}

				// This checks if the food has any missing ingredients
				if tempFood.Grade == "missing" {
					errors = append(errors, "Food has ingredients with undefined ingredients")
				} else if !foundFood {
					foundErr := fmt.Sprintf("There is no Food associated with barcode: %s", string(vals[0]))
					errors = append(errors, foundErr)
				} else {
					// Get all the ingredient information
					var tempList []Ingredient
					item := tempFood.Ingredients
					items := strings.Split(item, ",")

					// Create printing object struct
					var tempPage = FoodPage{tempFood, tempList}
					for i := 0; i < len(items); i++ {
						inName := items[i]
						inName = strings.ToLower(inName)
						inName = strings.Trim(inName, " ")

						ingredStr := fmt.Sprintf("%sapi/ingredient/%s", root, inName)
						inResp, inErr := http.Get(ingredStr)
						if inErr != nil {
							log.Println(1, i, inName, inErr)
						}
						defer inResp.Body.Close()
						inBody, inErrors := ioutil.ReadAll(inResp.Body)
						if inErrors != nil {
							log.Println(2, inErrors)
						}
						var tempIngred Ingredient
						inErrMsg := json.Unmarshal(inBody, &tempIngred)
						if inErrMsg != nil {
							log.Println(3, inErrMsg)
						}
						tempPage.Ingredients = append(tempPage.Ingredients, tempIngred)
					}

					// Now that we have our printing page, print it to the template
					newT, e := template.ParseFiles("templates/food2.html")
					if e != nil {
						log.Println(e)
					}
					newT.Execute(w, tempPage)
				}

			}

		} else if ok && len(vals[0]) == 0 {
			// Block should be reached if barcode was not defined
			errors = append(errors, "barcode field cannot be empty")
		}
		// If any errors exist, return an error code and message
		if len(errors) > 0 && ok {
			errorTemplate, err := template.ParseFiles("templates/foodError.html")
			if err != nil {
				log.Println(err)
			}
			errorTemplate.Execute(w, errors)
		}
	}
}

func handleJS(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["fileName"]
	url := fmt.Sprintf("js/%s", file)
	http.ServeFile(w, r, url)
}

func handleImages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["fileName"]
	url := fmt.Sprintf("images/%s", file)
	http.ServeFile(w, r, url)
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/about.html")
	t.Execute(w, nil)
}

// ADMIN PAGES
/* Admin page. Allows for food to be added to the database */
func makeFood(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("templates/makeFood.html")
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		r.ParseForm()
		barcode := r.Form["barcode"][0]
		name := r.Form["name"][0]
		ingred := strings.ToLower(r.Form["ingred"][0])

		//Reprint the makeFood template
		t, _ := template.ParseFiles("templates/makeFood.html")
		t.Execute(w, nil)

		if checkFoodExists(barcode) {
			templ, err := template.ParseFiles("templates/dupeFood.html")
			if err != nil {
				log.Println("Dupe Food ", err)
			}
			templ.Execute(w, barcode)
		} else {
			// Used to store the names of ingredients that are missing
			var missingIngredients []string

			// Calculate Grade
			list := strings.Split(ingred, ",")
			var total int
			var allFound = true

			for i := 0; i < len(list); i++ {
				oneIngred := list[i]
				oneIngred = strings.Trim(oneIngred, " ")
				oneIngred = strings.ToLower(oneIngred)
				url := fmt.Sprintf("%sapi/ingredient/%s", root, oneIngred)
				resp, err := http.Get(url)
				if err != nil {
					log.Println(err)
				}
				body, er := ioutil.ReadAll(resp.Body)
				if er != nil {
					log.Println(er)
				}

				var tempIngred Ingredient
				json.Unmarshal(body, &tempIngred)
				log.Println(tempIngred)
				if tempIngred.Grade == -10 {
					total = 0
					allFound = false
					recordMissingIngredient(tempIngred.Name)
					missingIngredients = append(missingIngredients, tempIngred.Name)
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

			// Create the food struct
			var temp = Food{barcode, name, ingred, grade, avgGrade}
			body, err := json.Marshal(temp)
			str := fmt.Sprintf("%sapi/food", root)
			if err != nil {
				log.Println(err)
			}
			_, er := http.Post(str, "application/json", bytes.NewBuffer(body))
			if er != nil {
				log.Println(er)
			}

			// Print out error or success message depending on if there are missing ingredients
			if len(missingIngredients) > 0 {
				templ, err := template.ParseFiles("templates/missingIngredients.html")
				if err != nil {
					log.Println(err)
				}
				templ.Execute(w, missingIngredients)
			} else {
				templ, err := template.ParseFiles("templates/foodSuccess.html")
				if err != nil {
					log.Println("main.makeFood", err)
				}
				templ.Execute(w, temp)
			}
		}
	}
}

func makeIngredient(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("templates/makeIngredient.html")
		log.Println(t)
		log.Println(1)
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		var errors []string

		// Parse the frontend for the name and grade entered by the user
		r.ParseForm()
		name := r.Form["name"][0]
		grade := r.Form["grade"][0]

		// Reserve html file
		t, _ := template.ParseFiles("templates/makeIngredient.html")
		t.Execute(w, nil)
		//formatting stuff
		name = strings.Trim(name, " ")
		name = strings.ToLower(name)
		grade = strings.Trim(grade, " ")

		// Check if the name and grade fields are empty
		if len(name) < 1 {
			errors = append(errors, "Name Field cannot be empty|")
		}
		if len(grade) < 1 {
			errors = append(errors, "Grade Field cannot be empty|")
		}

		// Create ingredient struct
		g, e := strconv.Atoi(grade)
		if e != nil {
			log.Println(e)
			errors = append(errors, "Grade could not be parsed. Check to make sure it is a number|")
		}

		// Check if the grade is in range
		if g < -5 || g > 5 || g%1 != 0 {
			errors = append(errors, "The grade must be an integer between -5 and 5, inclusive|")
		}

		// Check if the ingredient already exists
		stmt, err := db.Prepare("select * from ingredients where title=?;")
		if err != nil {
			log.Println(err)
		}
		rows, _ := stmt.Query(name)
		var count int
		for rows.Next() {
			count++
		}
		if count > 0 {
			dupIngred := fmt.Sprintf("Ingredient %s already exists|", name)
			errors = append(errors, dupIngred)
		}

		if len(errors) == 0 {
			var temp = Ingredient{name, g}
			body, err := json.Marshal(temp)
			str := fmt.Sprintf("%sapi/ingredient", root)
			if err != nil {
				log.Println(err)
			}
			_, er := http.Post(str, "application/json", bytes.NewBuffer(body))
			if er != nil {
				log.Println(er)
			}
			succString := fmt.Sprintf("Ingredient %s created with grade %d", temp.Name, temp.Grade)
			w.Write([]byte(succString))
		} else {
			errorString := strings.Join(errors, "|")
			log.Println(errorString)
			http.Error(w, errorString, 500)
		}
	}
}

// REST API handlers
// CRUD Operations Controllers - Food
func getAllFoods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("Select * from food;")
	if err != nil {
		log.Fatalln(err)
	}
	sel, err := stmt.Query()
	if err != nil {
		log.Fatalln(err)
	}
	for sel.Next() {
		var (
			bar      string
			name     string
			ins      string
			grade    string
			numgrade float64
		)
		sel.Scan(&bar, &name, &ins, &grade, &numgrade)
		var f = Food{bar, name, ins, grade, numgrade}
		json.NewEncoder(w).Encode(f)
	}
}

// Create - Current implentation allows for a single creation per call
func createFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newFood Food
	json.NewDecoder(r.Body).Decode(&newFood)
	log.Println(newFood)
	// Set up DB query
	stmt, err := db.Prepare("insert into food values(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	// Fire query string
	_, err = stmt.Query(newFood.Barcode, newFood.Name, newFood.Ingredients, newFood.Grade, newFood.NumGrade)
	if err != nil {
		log.Fatalln(err)
	}
}

// Read
func getFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("Select * from food where barcode=?;")
	if err != nil {
		log.Fatalln(err)
	}
	vars := mux.Vars(r)
	sel, err := stmt.Query(vars["bar"])
	if err != nil {
		log.Fatalln(err)
	}
	for sel.Next() {
		var (
			bar      string
			name     string
			ins      string
			grade    string
			numgrade float64
		)
		sel.Scan(&bar, &name, &ins, &grade, &numgrade)
		var f = Food{bar, name, ins, grade, numgrade}
		json.NewEncoder(w).Encode(f)
	}
}

// Update
func updateFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("Select * from food where barcode=?;")
	if err != nil {
		log.Fatalln(err)
	}
	rem, err := db.Prepare("update food set title=?, ingredients=?, grade=?, numgrade=? where barcode=?;")
	if err != nil {
		log.Fatalln(err)
	}
	vars := mux.Vars(r)
	sel, err := stmt.Query(vars["bar"])
	var (
		bar      string
		name     string
		ins      string
		grade    string
		numgrade float64
	)

	sel.Scan(&bar, &name, &ins, &grade, &numgrade)
	rem.Query(name, ins, grade, numgrade, bar)
	var f = Food{bar, name, ins, grade, numgrade}
	json.NewEncoder(w).Encode(f)
}

// Delete
func deleteFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("delete from food where bar=?")
	if err != nil {
		log.Fatalln(err)
	}
	vars := mux.Vars(r)
	stmt.Query(vars["bar"])
}

// CRUD Operation Controllers - Ingredients

// Create
func createIngredient(w http.ResponseWriter, r *http.Request) {
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	// Get ingredient data from http request
	var newIngred Ingredient
	json.NewDecoder(r.Body).Decode(&newIngred)

	// Set up DB Query
	stmt, err := db.Prepare("insert into ingredients values(?, ?);")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = stmt.Query(newIngred.Name, newIngred.Grade)
	if err != nil {
		log.Fatalln(err)
	}
}

// Read
func getIngredient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stmt, err := db.Prepare("select grade from ingredients where title=?;")
	if err != nil {
		log.Println(err)
	}

	vars := mux.Vars(r)
	sel, err := stmt.Query(vars["name"])
	if err != nil {
		log.Fatalln(err)
	}
	var (
		grade int
	)
	if sel.Next() {
		sel.Scan(&grade)
		if err != nil {
			log.Println(err)
		}
		var i = Ingredient{vars["name"], grade}
		json.NewEncoder(w).Encode(i)
	} else {
		var i = Ingredient{vars["name"], -10}
		json.NewEncoder(w).Encode(i)
	}
}

// Update
func updateIngredient(w http.ResponseWriter, r *http.Request) {

}

// Delete
func deleteIngredient(w http.ResponseWriter, r *http.Request) {

}

// HELPER FUNCTIONS

// Function used to record the name of any ingredients not currently in the ingredients table
func recordMissingIngredient(name string) {
	stmt, err := db.Prepare("insert into missing values(?)")
	if err != nil {
		log.Println(err)
	}
	stmt.Query(name)
}

func checkFoodExists(bar string) bool {
	str := fmt.Sprintf("%sapi/food/%s", root, bar)
	resp, err := http.Get(str)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, erros := ioutil.ReadAll(resp.Body)
	if erros != nil {
		log.Println("checkFoodExists: ", erros)
	}

	var tempFood Food
	var exists = true
	errMsg := json.Unmarshal(body, &tempFood)
	if errMsg != nil {
		exists = false
	}
	return exists
}

/* To Do
Add restrictions that limit who can use admin pages and api
SQL Injection protection
Implement and test update/delete methods for ingredients and food

Edge Cases
	DONE!!! (For now)
*/
