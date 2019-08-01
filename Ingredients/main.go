package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// DB is s typing of sql.DB to allow readability of code
type DB sql.DB

// Food is a typed struct that is used to decode and encode
// data for http requests
type Food struct {
	Barcode     string `json:"barcode"`
	Name        string `json:"title"`
	Ingredients string `json:"ingredients"`
	Grade       string `json:"grade"`
}

// Ingredient is a typed struct that is used to decode and
// encode data for http requests
type Ingredient struct {
	Name  string `json:"title"`
	Grade int    `json:"grade"`
}

var db *sql.DB

func main() {
	//Router to handle all routings
	router := mux.NewRouter()
	//endpoints for the api
	router.HandleFunc("/api/food/{bar}", getFood).Methods("GET")
	router.HandleFunc("/api/food/{bar}", updateFood).Methods("POST")
	router.HandleFunc("/api/food/{bar}", deleteFood).Methods("DELETE")
	router.HandleFunc("/api/food", getAllFoods).Methods("GET")
	router.HandleFunc("/api/food", createFood).Methods("POST")
	router.HandleFunc("/api/ingredient", createIngredient).Methods("POST")
	router.HandleFunc("/api/ingredient/{name}", getIngredient).Methods("GET")
	router.HandleFunc("/", landingFunc).Methods("GET")

	// Open and Test DB connection
	openConn()

	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatalln(err)
	}
}

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

func landingFunc(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

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
			bar   string
			name  string
			ins   string
			grade string
		)
		sel.Scan(&bar, &name, &ins, &grade)
		var f = Food{bar, name, ins, grade}
		json.NewEncoder(w).Encode(f)
	}
}

// Create - Current implentation allows for a single creation per call
func createFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newFood Food
	json.NewDecoder(r.Body).Decode(&newFood)

	// Set up DB query
	stmt, err := db.Prepare("insert into food values(?, ?, ?, ?)")
	if err != nil {
		log.Fatalln(err)
	}
	// Fire query string
	_, err = stmt.Query(newFood.Barcode, newFood.Name, newFood.Ingredients, newFood.Grade)
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
			bar   string
			name  string
			ins   string
			grade string
		)
		sel.Scan(&bar, &name, &ins, &grade)
		var f = Food{bar, name, ins, grade}
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
	rem, err := db.Prepare("update food set title=?, ingredients=?, grade=? where barcode=?;")
	if err != nil {
		log.Fatalln(err)
	}
	vars := mux.Vars(r)
	sel, err := stmt.Query(vars["bar"])
	var (
		bar   string
		name  string
		ins   string
		grade string
	)

	sel.Scan(&bar, &name, &ins, &grade)
	rem.Query(name, ins, grade, bar)
	var f = Food{bar, name, ins, grade}
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
		log.Fatalln(err)
	}
	vars := mux.Vars(r)
	sel, err := stmt.Query(vars["name"])
	if err != nil {
		log.Fatalln(err)
	}
	var (
		grade int
	)
	sel.Scan(&grade)
	var i = Ingredient{vars["name"], grade}
	json.NewEncoder(w).Encode(i)
}

// Update
func updateIngredient(w http.ResponseWriter, r *http.Request) {

}

// Delete
func deleteIngredient(w http.ResponseWriter, r *http.Request) {

}
