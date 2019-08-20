package server

import (
	"IngredientGrader/data"
	"log"

	"golang.org/x/crypto/bcrypt"
)

var db = data.DB

/* This package deals with server-side functions needed for the site
   to operate correctly.
*/

// GetFood is a function that retrieves Food data from a database,
/* Constructs a food object from it, and returns it. If the food
   does not exist within the database, a zero Food and false is
   returned
   db - Database that is to be searched
   barcode - The barcode associated with the food
*/
func GetFood(barcode string) (data.Food, bool) {
	// Create statement
	stmt, err := db.Prepare("Select * from food where barcode=?;")
	if err != nil {
		log.Println("server.GetFood: ", err)
	}

	sel, err := stmt.Query(barcode)
	if err != nil {
		log.Println("server.GetFood: ", err)
	}

	if sel.Next() {
		var (
			bar      string
			name     string
			ins      string
			grade    string
			numgrade float64
		)
		sel.Scan(&bar, &name, &ins, &grade, &numgrade)
		var f = data.Food{Barcode: bar, Name: name, Ingredients: ins, Grade: grade, NumGrade: numgrade}
		return f, true
	}
	return data.Food{}, false
}

// CreateFood adds an entry to the database with the associated food data
/*
   db - The database that the food is being added to
   barcode - The barcode of the food being added to the database
   name - THe name of the food being added, lowercase
   ingredients - The names of all the ingredients of the food, lowercase in
	   a comma-separated list
   grade - The categorical grade of the food - very bad, bad, neutral, good, and very good
   avgGrade - The average numerical grade of the food being added
*/
func CreateFood(barcode, name, ingredients, grade string, avgGrade float64) {
	stmt, err := db.Prepare("insert into food values(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalln("server.CreateFood: ", err)
	}
	// Fire query string
	_, err = stmt.Query(barcode, name, ingredients, grade, avgGrade)
	if err != nil {
		log.Fatalln("server.CreateFood: ", err)
	}
}

// GetIngredient retrieves an ingredient with a matching name from the
/* database, constructs an ingredient object. name does not need to be
   checked for existence, as it is up to the user to check that
   db - The database that the ingredient is being retrieved from
   name - The name of the ingredient being retrieved
*/
func GetIngredient(name string) data.Ingredient {
	// Create sql statement
	stmt, err := db.Prepare("select grade from ingredients where title=?;")
	if err != nil {
		log.Println("server.GetIngredient: ", err)
	}

	sel, err := stmt.Query(name)
	if err != nil {
		log.Println("server.GetIngredient: ", err)
	}

	// Grade used to temp hold data from database, i to hold constructed
	// Ingredient struct
	var (
		grade int
		i     data.Ingredient
	)

	// Check if the io.Reader has anything in it. If so, there was a match
	if sel.Next() {
		sel.Scan(&grade)
		i = data.Ingredient{Name: name, Grade: grade}
	} else {
		// Grade of -10 signals that no ingredient was found
		i = data.Ingredient{Name: name, Grade: -10}
	}
	return i
}

// CreateIngredient takes in the name and grade of a prospective ingredient,
/* and adds it to the database. Currently, it is up to the user to check
   that name and grade are valid inputs
   db - The database that the ingredient will be added to
   name - The name of the ingredient being added, lowercase
   grade - The grade of the ingredient, -5 to 5 inclusive integer
*/
func CreateIngredient(name string, grade int) {
	// Set up DB Query
	stmt, err := db.Prepare("insert into ingredients values(?, ?);")
	if err != nil {
		log.Fatalln("server.MakeIngredient: ", err)
	}
	_, err = stmt.Query(name, grade)
	if err != nil {
		log.Fatalln("server.MakeIngredient: ", err)
	}
}

// RecordMissingIngredient records the name of a missing ingredient to the
/* database for administration to grade later
   db - The database that the ingredient name is being added to
   name - THe name of the ingredient missing from the database
*/
func RecordMissingIngredient(name string) {
	stmt, err := db.Prepare("insert into missing values(?)")
	if err != nil {
		log.Println(err)
	}
	stmt.Query(name)
}

// Obfuscate hashes and salts a potential password. This can be used for both
/* logging in and registering an account, as writing the hash to database is
   not done with this function
   pass - The byte slice equivalent of the string password
   return - Returns the hashed and salted password as a string
*/
func Obfuscate(pass []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		log.Println("Obfuscate: ", err)
	}
	return string(hash)
}

// GetHashedPassword retrieves a salted and hashed password with a matching
/* username from the database. In the scenario that there is no matching
   username, an empty string is returned
   username - the username associated with the password
   return - the salted and hashed password associated with the username
*/
func GetHashedPassword(username string) string {
	db = data.DB
	stmt, err := db.Prepare("select hashedPass from users where username=?;")
	if err != nil {
		log.Println("getHashedPassword: ", err)
	}

	sel, err := stmt.Query(username)
	if err != nil {
		log.Println("getHashedPassword/Query: ", err)
	}

	var pass string
	if sel.Next() {
		sel.Scan(&pass)
		return pass
	}
	return ""
}

// PasswordMatch checks if the hash is equivalent to the password
func PasswordMatch(hash, pass []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, pass)
	if err != nil {
		return false
	}
	return true
}

// Init initializes the db field of the server package. Can only be
/* called after calling data.Init
 */
func Init() {
	db = data.DB
}
