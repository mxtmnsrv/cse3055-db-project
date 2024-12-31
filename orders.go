package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
)

type Order struct {
	OrderID    int
	OrderDate  string
	OrderNote  string
	CustomerID int
}

func viewOrders(w http.ResponseWriter, r *http.Request) {
	// SQL query to fetch orders, now without CustomerName
	query := `
		SELECT o.OrderID, o.OrderDate, o.OrderNote, o.CustomerID
		FROM [Order] o
	`
	rows, err := shared.DB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.OrderID, &order.OrderDate, &order.OrderNote, &order.CustomerID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}

	tmpl, err := template.ParseFiles("templates/orders.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, orders)
}

func addOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Load the list of customers
		rows, err := shared.DB.Query("SELECT CustomerID FROM Customer")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var customers []struct {
			CustomerID int
		}
		for rows.Next() {
			var customer struct {
				CustomerID int
			}
			if err := rows.Scan(&customer.CustomerID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			customers = append(customers, customer)
		}

		tmpl, err := template.ParseFiles("templates/addOrder.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, customers)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		customerID := r.FormValue("customerID")
		orderNote := r.FormValue("orderNote")
		orderDate := r.FormValue("orderDate")

		// SQL query to insert order
		query := "INSERT INTO [Order] (OrderDate, OrderNote, CustomerID) VALUES (@p1, @p2, @p3)"
		_, err := shared.DB.Exec(query, orderDate, orderNote, customerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/orders", http.StatusSeeOther)
	}
}
