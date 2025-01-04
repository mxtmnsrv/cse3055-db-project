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
	http.HandleFunc("/addOrder", addOrder)
	http.HandleFunc("/orders", viewOrders)
	http.HandleFunc("/addEmployee", addEmployee)
	http.HandleFunc("/employees", viewEmployees)
	http.HandleFunc("/setLogistics", setLogisticsToOrder)
	http.HandleFunc("/addProductToWarehouse", addProductToWarehouse)
	http.HandleFunc("/viewProductsInWarehouse", viewProductsInWarehouse)

	http.HandleFunc("/supervisorPage", supervisorPage)
	http.HandleFunc("/managerPage", managerPage)

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

	if r.Method == http.MethodPost {
		r.ParseForm()
		employeeID := r.FormValue("employeeID")
		password := r.FormValue("password")

		var role string
		// Query to get the role of the employee
		query := `SELECT Role FROM Employee WHERE EmployeeID = @p1 AND Password = @p2`
		err := shared.DB.QueryRow(query, employeeID, password).Scan(&role)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Redirect based on the role
		if role == "Supervisor" {
			http.Redirect(w, r, "/supervisorPage", http.StatusSeeOther)
		} else if role == "Manager" {
			http.Redirect(w, r, "/managerPage", http.StatusSeeOther)
		} else {
			http.Error(w, "Access Denied", http.StatusForbidden)
		}
	}
}

func supervisorPage(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated (you can use sessions/cookies for actual authentication)
	tmpl, err := template.ParseFiles("templates/supervisorPage.html")
	if err != nil {
		http.Error(w, "Error loading supervisor page template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func managerPage(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated (you can use sessions/cookies for actual authentication)
	tmpl, err := template.ParseFiles("templates/managerPage.html")
	if err != nil {
		http.Error(w, "Error loading manager page template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
