package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type Order struct {
	ProductID     int
	Quantity      int
	CustomerID    int
	PaymentMethod string
	OwnerName     string
	CardNumber    string
	ExpMonth      string
	ExpYear       string
	CCV           string
	BankName      string
	CheckDate     string
	AccountHolder string
}

func addOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// For GET requests, render the order creation form
		tmpl, err := template.ParseFiles("templates/addOrder.html")
		if err != nil {
			http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		// Handle POST request for creating an order

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form data: "+err.Error(), http.StatusBadRequest)
			return
		}

		productID, err := strconv.Atoi(r.FormValue("productID"))
		if err != nil {
			http.Error(w, "Invalid Product ID: "+err.Error(), http.StatusBadRequest)
			return
		}

		quantity, err := strconv.Atoi(r.FormValue("quantity"))
		if err != nil {
			http.Error(w, "Invalid Quantity: "+err.Error(), http.StatusBadRequest)
			return
		}

		customerID, err := strconv.Atoi(r.FormValue("customerID"))
		if err != nil {
			http.Error(w, "Invalid Customer ID: "+err.Error(), http.StatusBadRequest)
			return
		}

		paymentMethod := r.FormValue("paymentMethod")
		ownerName := r.FormValue("ownerName")
		cardNumber := r.FormValue("cardNumber")
		expMonth := r.FormValue("expMonth")
		expYear := r.FormValue("expYear")
		ccv := r.FormValue("ccv")
		bankName := r.FormValue("bankName")
		checkDate := r.FormValue("checkDate")
		accountHolder := r.FormValue("accountHolderName")

		// Insert the order into the "Order" table
		orderDate := time.Now().Format("2006-01-02")
		orderQuery := `
			INSERT INTO [Order] (OrderDate, CustomerID, EmployeeID) 
			VALUES (@p1, @p2, @p3)`
		_, err = shared.DB.Exec(orderQuery, orderDate, customerID, 1) // Assuming EmployeeID is 1 (you can adjust this)
		if err != nil {
			http.Error(w, "Error inserting order: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Fetch the last inserted OrderID
		var orderID int
		err = shared.DB.QueryRow("SELECT SCOPE_IDENTITY()").Scan(&orderID)
		if err != nil {
			http.Error(w, "Error fetching Order ID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert OrderDetail (Product and Quantity)
		orderDetailQuery := `
			INSERT INTO OrderDetail (OrderID, ProductID, Quantity) 
			VALUES (@p1, @p2, @p3)`
		_, err = shared.DB.Exec(orderDetailQuery, orderID, productID, quantity)
		if err != nil {
			http.Error(w, "Error inserting order detail: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert Payment based on Payment Method
		paymentDate := time.Now().Format("2006-01-02")
		paymentQuery := `
			INSERT INTO Payment (PaymentDate, PaidAmount, PaymentStatus, OrderID)
			VALUES (@p1, @p2, 'Completed', @p3)`
		// Assume PaidAmount = quantity * price (price should be fetched from the product table)
		var price float64
		err = shared.DB.QueryRow("SELECT Price FROM Product WHERE ProductID = @p1", productID).Scan(&price)
		if err != nil {
			http.Error(w, "Error fetching product price: "+err.Error(), http.StatusInternalServerError)
			return
		}
		paidAmount := price * float64(quantity)

		_, err = shared.DB.Exec(paymentQuery, paymentDate, paidAmount, orderID)
		if err != nil {
			http.Error(w, "Error inserting payment: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Now insert payment details based on the payment method
		if paymentMethod == "Cash" {
			query := "INSERT INTO Cash (PaymentID, OwnerName) VALUES (SCOPE_IDENTITY(), @p1)"
			_, err = shared.DB.Exec(query, ownerName)
		} else if paymentMethod == "CreditCard" {
			query := `
				INSERT INTO CreditCard (PaymentID, OwnerName, CardNumber, ExpMonth, ExpYear, CCV)
				VALUES (SCOPE_IDENTITY(), @p1, @p2, @p3, @p4, @p5)`
			_, err = shared.DB.Exec(query, ownerName, cardNumber, expMonth, expYear, ccv)
		} else if paymentMethod == "Check" {
			query := `
				INSERT INTO [Check] (PaymentID, BankName, CheckDate, AccountHolderName)
				VALUES (SCOPE_IDENTITY(), @p1, @p2, @p3)`
			_, err = shared.DB.Exec(query, bankName, checkDate, accountHolder)
		}

		if err != nil {
			http.Error(w, "Error inserting payment details: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to the orders page or success message
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
	}
}
