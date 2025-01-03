package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
)

type Ordero struct {
	OrderID     int
	OrderDate   string
	OrderNote   *string // Use a pointer to handle NULL values
	CustomerID  int
	LogisticsID *int // Use a pointer to handle NULL values
	EmployeeID  int
}

func viewOrders(w http.ResponseWriter, r *http.Request) {
	// Query to fetch only the required fields
	rows, err := shared.DB.Query(`
		SELECT OrderID, OrderDate, OrderNote, CustomerID, LogisticsID, EmployeeID
		FROM [Order]
	`)
	if err != nil {
		http.Error(w, "Error fetching orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Ordero
	for rows.Next() {
		var order Ordero
		// Scan values, if any field is NULL, it will be set to nil
		if err := rows.Scan(&order.OrderID, &order.OrderDate, &order.OrderNote, &order.CustomerID, &order.LogisticsID, &order.EmployeeID); err != nil {
			http.Error(w, "Error scanning order data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// If OrderNote is nil, set it to "NULL"
		if order.OrderNote == nil {
			placeholder := "NULL"
			order.OrderNote = &placeholder
		}

		// If LogisticsID is nil, set it to a default value or "NULL"
		if order.LogisticsID == nil {
			nullLogisticsID := -1 // Or you can use another indicator value like 0
			order.LogisticsID = &nullLogisticsID
		}

		orders = append(orders, order)
	}

	tmpl, err := template.ParseFiles("templates/orders.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, orders)
}
