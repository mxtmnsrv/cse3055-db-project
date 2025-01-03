package main

import (
	"database/sql"
	"dbs-term-project/shared"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	initDB()
	http.HandleFunc("/login", login)
	http.HandleFunc("/products", viewProducts)
	http.HandleFunc("/addProduct", addProduct)
	http.HandleFunc("/addCustomer", addCustomer)
	http.HandleFunc("/customers", viewCustomers)
	// !!!
	http.HandleFunc("/addOrder", addOrder)
	http.HandleFunc("/orders", viewOrders)
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	var err error
	connString := "server=ASUS-TUF-MKSTAK;user id=webAppUser;password=123;database=StoreDB;encrypt=disable"
	shared.DB, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	err = shared.DB.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Connected!")
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("templates/login.html")
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
