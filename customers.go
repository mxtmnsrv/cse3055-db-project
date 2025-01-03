package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
)

type Customer struct {
	CustomerID   int
	PhoneNumber  string
	Email        string
	CustomerType string
	Details      string
}

func viewCustomers(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT 
			C.CustomerID, 
			C.PhoneNumber, 
			C.Email, 
			C.CustomerType,
			CASE 
				WHEN C.CustomerType = 'Individual' THEN CONCAT(I.Name, ' (National ID: ', I.NationalID, ')')
				WHEN C.CustomerType = 'Corporate' THEN CONCAT(Co.CompanyName, ' (Tax: ', Co.TaxNumber, ')')
				ELSE 'Unknown'
			END AS Details
		FROM Customer C
		LEFT JOIN Individual I ON C.CustomerID = I.CustomerID
		LEFT JOIN Corporate Co ON C.CustomerID = Co.CustomerID`

	rows, err := shared.DB.Query(query)
	if err != nil {
		http.Error(w, "Error fetching customers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var cust Customer
		if err := rows.Scan(&cust.CustomerID, &cust.PhoneNumber, &cust.Email, &cust.CustomerType, &cust.Details); err != nil {
			http.Error(w, "Error scanning customer data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		customers = append(customers, cust)
	}

	tmpl, err := template.ParseFiles("templates/customers.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, customers)
}

func addCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/addCustomer.html")
		if err != nil {
			http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
			return
		}

		customerID := r.FormValue("customerID")
		phoneNumber := r.FormValue("phoneNumber")
		emailAddress := r.FormValue("emailAddress")
		customerType := r.FormValue("customerType")

		query := "INSERT INTO Customer (CustomerID, PhoneNumber, Email, CustomerType) VALUES (@p1, @p2, @p3, @p4)"
		_, err := shared.DB.Exec(query, customerID, phoneNumber, emailAddress, customerType)
		if err != nil {
			http.Error(w, "Error inserting customer: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if customerType == "Individual" {
			nationalID := r.FormValue("nationalID")
			name := r.FormValue("name")
			dob := r.FormValue("dob")
			query = "INSERT INTO Individual (CustomerID, NationalID, Name, DateOfBirth) VALUES (@p1, @p2, @p3, @p4)"
			_, err = shared.DB.Exec(query, customerID, nationalID, name, dob)
		} else if customerType == "Corporate" {
			companyName := r.FormValue("companyName")
			taxNumber := r.FormValue("taxNumber")
			query = "INSERT INTO Corporate (CustomerID, CompanyName, TaxNumber) VALUES (@p1, @p2, @p3)"
			_, err = shared.DB.Exec(query, customerID, companyName, taxNumber)
		} else {
			http.Error(w, "Invalid customer type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Error inserting customer details: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
	}
}
