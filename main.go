package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
)

type Product struct {
	ProductID   int
	Name        string
	Category    string
	Price       float64
	Description string
}

var db *sql.DB

func main() {
	initDB()
	http.HandleFunc("/login", login)
	http.HandleFunc("/products", viewProducts)
	http.HandleFunc("/addProduct", addProduct)
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	var err error
	connString := "server=ASUS-TUF-MKSTAK;user id=webAppUser;password=okey123;database=ConstructionStoreDB;encrypt=disable"
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Connected!")
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("login.html")
		tmpl.Execute(w, nil)
		return
	}

	// Simple username/password check
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Replace with actual authentication logic
		if username == "manager" && password == "123" {
			http.Redirect(w, r, "/products", http.StatusSeeOther)
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	}
}

func viewProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT ProductID, Name, Category, Price, Description FROM Product")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ProductID, &product.Name, &product.Category, &product.Price, &product.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	tmpl, _ := template.ParseFiles("products.html")
	tmpl.Execute(w, products)
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("addProduct.html")
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		productID := r.FormValue("productID")
		name := r.FormValue("name")
		category := r.FormValue("category")
		price := r.FormValue("price")
		description := r.FormValue("description")

		// SQL query to insert product
		query := "INSERT INTO Product (ProductID, Name, Category, Price, Description) VALUES (@p1, @p2, @p3, @p4, @p5)"
		_, err := db.Exec(query, productID, name, category, price, description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Product added successfully!")
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}
