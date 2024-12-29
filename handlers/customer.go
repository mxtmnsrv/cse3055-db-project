package handlers

import (
	"dbs-term-project/db"
	"dbs-term-project/models"
	"encoding/json"
	"net/http"
)

func GetCustomers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT CustomerID, PhoneNumber, EmailAddress, CustomerType FROM Customer")
	if err != nil {
		http.Error(w, "Error fetching customers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		if err := rows.Scan(&customer.CustomerID, &customer.PhoneNumber, &customer.EmailAddress, &customer.CustomerType); err != nil {
			http.Error(w, "Error scanning customer data", http.StatusInternalServerError)
			return
		}
		customers = append(customers, customer)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}
