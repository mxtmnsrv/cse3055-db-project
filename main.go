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
	http.HandleFunc("/addCustomer", addCustomer)
	http.HandleFunc("/customers", viewCustomers)
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

		//fmt.Fprintf(w, "Product added successfully!")
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}

func viewCustomers(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT 
			C.CustomerID, 
			C.PhoneNumber, 
			C.EmailAddress, 
			C.CustomerType,
			CASE 
				WHEN C.CustomerType = 'Individual' THEN I.Name + ' (National ID: ' + CAST(I.NationalID AS NVARCHAR) + ')'
				WHEN C.CustomerType = 'Corporate' THEN Co.CompanyName + ' (Tax: ' + Co.TaxNumber + ')'
			END AS Details
		FROM Customer C
		LEFT JOIN Individual I ON C.CustomerID = I.CustomerID
		LEFT JOIN Corporate Co ON C.CustomerID = Co.CustomerID`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Customer struct {
		CustomerID   int
		PhoneNumber  string
		EmailAddress string
		CustomerType string
		Details      string
	}

	var customers []Customer
	for rows.Next() {
		var cust Customer
		if err := rows.Scan(&cust.CustomerID, &cust.PhoneNumber, &cust.EmailAddress, &cust.CustomerType, &cust.Details); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		customers = append(customers, cust)
	}

	tmpl, _ := template.ParseFiles("customers.html")
	tmpl.Execute(w, customers)
}

func addCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("addCustomer.html")
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		customerID := r.FormValue("customerID")
		phoneNumber := r.FormValue("phoneNumber")
		emailAddress := r.FormValue("emailAddress")
		customerType := r.FormValue("customerType")

		// Insert into Customer table
		query := "INSERT INTO Customer (CustomerID, PhoneNumber, EmailAddress, CustomerType) OUTPUT INSERTED.CustomerID VALUES (@p1, @p2, @p3, @p4)"
		_, err := db.Exec(query, customerID, phoneNumber, emailAddress, customerType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert into Individual or Corporate table
		if customerType == "Individual" {
			nationalID := r.FormValue("nationalID")
			name := r.FormValue("name")
			dob := r.FormValue("dob")
			query = "INSERT INTO Individual (CustomerID, NationalID, Name, DateOfBirth) VALUES (@p1, @p2, @p3, @p4)"
			_, err = db.Exec(query, customerID, nationalID, name, dob)
		} else {
			companyName := r.FormValue("companyName")
			taxNumber := r.FormValue("taxNumber")
			query = "INSERT INTO Corporate (CustomerID, CompanyName, TaxNumber) VALUES (@p1, @p2, @p3)"
			_, err = db.Exec(query, customerID, companyName, taxNumber)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
	}
}
