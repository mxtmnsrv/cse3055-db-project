package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
)

type Customer struct {
	CustomerID   int
	PhoneNumber  string
	EmailAddress string
	CustomerType string
	Details      string
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

	rows, err := shared.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var cust Customer
		if err := rows.Scan(&cust.CustomerID, &cust.PhoneNumber, &cust.EmailAddress, &cust.CustomerType, &cust.Details); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		customers = append(customers, cust)
	}

	tmpl, _ := template.ParseFiles("templates/customers.html")
	tmpl.Execute(w, customers)
}

func addCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("templates/addCustomer.html")
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
		_, err := shared.DB.Exec(query, customerID, phoneNumber, emailAddress, customerType)
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
			_, err = shared.DB.Exec(query, customerID, nationalID, name, dob)
		} else {
			companyName := r.FormValue("companyName")
			taxNumber := r.FormValue("taxNumber")
			query = "INSERT INTO Corporate (CustomerID, CompanyName, TaxNumber) VALUES (@p1, @p2, @p3)"
			_, err = shared.DB.Exec(query, customerID, companyName, taxNumber)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
	}
}
