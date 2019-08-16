package data

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Food is a struct that contains the information of a single Food
/*	Barcode - The UPC-A code of the food. The barcode of the food must be unique
	Name - The name of the food
	Ingredients - A comma separated list of all the ingredients of the food
	Grade - The categorical grade of the food - Very Bad, Bad, Neutral, Good, Very Good
	NumGrade - The numerical grade of the food. This is a floating point number between -5
	and 5, inclusive.
*/
type Food struct {
	Barcode     string  `json:"barcode"`
	Name        string  `json:"title"`
	Ingredients string  `json:"ingredients"`
	Grade       string  `json:"grade"`
	NumGrade    float64 `json:"numgrade"`
}

// Ingredient is a struct that contains the information of a single ingredient
/*	Name - The name of the Ingredient. The name of the ingredient must be unique
	Grade - The grade of the Ingredient. This is a int between -5 and 5, inclusive
*/
type Ingredient struct {
	Name  string `json:"title"`
	Grade int    `json:"grade"`
}

// Content is a struct that contains any dynamic information that is printed
/* to the page. It is assumed that this struct will be used to populate an
html template. There is no guarantee of the states that Content will be initialized
with. It is up to the user to account for nil Food and Ingredients in the
Content struct.
	PageFood - The Food object that will be printed to the template. This will
	most likely be nil in the case that no Food is to be printed to the page
	PageIngredients - The Ingredients objects that will be printed to the template.
	This will most likely be nil in the case that there are no Ingredients to
	be printed to the page
	PageErrors - If there is anything wrong will the user's input, these errors
	will be printed as Bootstrap 4 Alert boxes. Even if PageFood and PageIngredients
	are defined, only PageErrors should be used if errors exist
	Success - If no errors were thrown, Success is true
	Source - The handling function that was executed associated with this struct
*/
type Content struct {
	PageFood        Food         `json:"food"`
	PageIngredients []Ingredient `json:"ingredients"`
	PageErrors      []string     `json:"errors"`
	Success         bool         `json:"success"`
	Source          string       `json:"source"`
}

// AddError adds an error to the PageErrors slice in a Content object
func (c *Content) AddError(err string) {
	c.PageErrors = append(c.PageErrors, err)
}

// HasErrors returns true if there is at least a single error associated
// with the Content struct
func (c *Content) HasErrors() bool {
	if len(c.PageErrors) > 0 {
		return true
	}
	return false
}

// AddIngredient adds an ingredient to PageIngredients slice in a Content object
func (c *Content) AddIngredient(ingred Ingredient) {
	c.PageIngredients = append(c.PageIngredients, ingred)
}

// DB is a database handle that will be used to read and write data to/from the database
var DB *sql.DB

// Init must be called before anything else in main.go to establish connection
/* to the database. If this is not done, no webpage will work correctly. functions
   in admin and server packages will not work correctly without this being run first
*/
func Init() error {
	// First open the connection
	creds := fmt.Sprintf("%s:%s@%s", os.Getenv("GRADER_USER"), os.Getenv("GRADER_PASS"), os.Getenv("GRADER_LOC"))
	tempdb, err := sql.Open("mysql", creds)
	if err != nil {
		return err
	}
	// If all checks out, assign it to DB
	DB = tempdb
	// Now check if it works
	err = DB.Ping()
	if err != nil {
		return err
	}
	return nil
}
