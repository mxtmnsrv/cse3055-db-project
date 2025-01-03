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
		tmpl, err := template.ParseFiles("templates/addOrder.html")
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

		productID := r.FormValue("productID")
		quantityStr := r.FormValue("quantity")
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			http.Error(w, "Invalid quantity: "+err.Error(), http.StatusBadRequest)
			return
		}
		customerID := r.FormValue("customerID")
		paymentMethod := r.FormValue("paymentMethod")
		cashOwnerName := r.FormValue("cashOwnerName")
		cardOwnerName := r.FormValue("cardOwnerName")
		cardNumber := r.FormValue("cardNumber")
		expMonth := r.FormValue("expMonth")
		expYear := r.FormValue("expYear")
		ccv := r.FormValue("ccv")
		bankName := r.FormValue("bankName")
		checkDate := r.FormValue("checkDate")
		accountHolder := r.FormValue("accountHolderName")

		// Fetch the last inserted OrderID
		orderDate := time.Now().Format("2006-01-02")
		var orderID int
		orderQuery := `
			INSERT INTO [Order] (OrderDate, CustomerID, EmployeeID)
			OUTPUT INSERTED.OrderID
			VALUES (@p1, @p2, @p3)
		`
		err = shared.DB.QueryRow(orderQuery, orderDate, customerID, 1).Scan(&orderID)
		if err != nil {
			http.Error(w, "Error inserting order or fetching Order ID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Fetch the product price
		var price float64
		err = shared.DB.QueryRow("SELECT Price FROM Product WHERE ProductID = @p1", productID).Scan(&price)
		if err != nil {
			http.Error(w, "Error fetching product price: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate the payment amount
		paidAmount := price * float64(quantity)

		// Insert Payment
		paymentDate := time.Now().Format("2006-01-02")
		var paymentID int
		paymentQuery := "INSERT INTO Payment (PaymentDate, PaidAmount, PaymentStatus, OrderID) OUTPUT INSERTED.PaymentID VALUES (@p1, @p2, 'Completed', @p3)"
		err = shared.DB.QueryRow(paymentQuery, paymentDate, paidAmount, orderID).Scan(&paymentID)
		if err != nil {
			http.Error(w, "Error inserting payment: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Insert Payment Details based on Payment Method
		if paymentMethod == "Cash" {
			query := "INSERT INTO Cash (PaymentID, OwnerName) VALUES (@p1, @p2)"
			_, err = shared.DB.Exec(query, paymentID, cashOwnerName)
		} else if paymentMethod == "CreditCard" {
			query := "INSERT INTO CreditCard (PaymentID, OwnerName, CardNumber, ExpMonth, ExpYear, CCV) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)"
			_, err = shared.DB.Exec(query, paymentID, cardOwnerName, cardNumber, expMonth, expYear, ccv)
		} else if paymentMethod == "Check" {
			query := "INSERT INTO [Check] (PaymentID, BankName, CheckDate, AccountHolderName) VALUES (@p1, @p2, @p3, @p4)"
			_, err = shared.DB.Exec(query, paymentID, bankName, checkDate, accountHolder)
		}

		if err != nil {
			http.Error(w, "Error inserting payment details: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect to orders page
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
	}
}
