package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type Employee struct {
	EmployeeID  int
	FirstName   string
	LastName    string
	Salary      int
	PhoneNumber string
	Role        string
	Details     string
}

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

// Distribution Agent
func setLogisticsToOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		orderID := r.FormValue("orderID")
		employeeID := r.FormValue("employeeID")

		// Insert a new logistics record and use OUTPUT to get the LogisticsID
		insertLogisticsQuery := `
			INSERT INTO Logistics (Date, EmployeeID)
			OUTPUT INSERTED.LogisticsID
			VALUES (GETDATE(), @p1)
		`

		// Execute the insert query and retrieve the LogisticsID using OUTPUT
		var logisticsID int64
		err := shared.DB.QueryRow(insertLogisticsQuery, employeeID).Scan(&logisticsID)
		if err != nil {
			http.Error(w, "Error inserting logistics record: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the order with the new LogisticsID
		updateOrderQuery := `
			UPDATE [Order]
			SET LogisticsID = @p1
			WHERE OrderID = @p2
		`

		_, err = shared.DB.Exec(updateOrderQuery, logisticsID, orderID)
		if err != nil {
			http.Error(w, "Error updating order with logistics ID: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Redirect or show a success message
		http.Redirect(w, r, "/orders", http.StatusSeeOther)
	}

	// Get all orders with NULL LogisticsID
	rows, err := shared.DB.Query(`
        SELECT OrderID, OrderDate, OrderNote, CustomerID, LogisticsID, EmployeeID
        FROM [Order]
        WHERE LogisticsID IS NULL
    `)
	if err != nil {
		http.Error(w, "Error fetching orders: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []Ordero
	for rows.Next() {
		var order Ordero
		if err := rows.Scan(&order.OrderID, &order.OrderDate, &order.OrderNote, &order.CustomerID, &order.LogisticsID, &order.EmployeeID); err != nil {
			http.Error(w, "Error scanning order data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		orders = append(orders, order)
	}

	tmpl, err := template.ParseFiles("templates/setLogistics.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, orders)
}

// Field Marketer
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

func viewEmployees(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT 
			E.EmployeeID, 
			E.FirstName, 
			E.LastName, 
			E.Salary, 
			E.PhoneNumber, 
			E.Role,
			CASE 
				WHEN E.Role = 'Manager' THEN CONCAT('Team Size: ', M.TeamSize)
				WHEN E.Role = 'Supervisor' THEN CONCAT('Manager ID: ', S.ManagerID, ', Team Size: ', S.TeamSize)
				WHEN E.Role = 'SalesRepresentative' THEN CONCAT('Supervisor ID: ', SR.SupervisorID, ', Shift Duration: ', SR.ShiftDuration)
				WHEN E.Role = 'Cashier' THEN CONCAT('Supervisor ID: ', C.SupervisorID, ', Shift Duration: ', C.ShiftDuration)
				WHEN E.Role = 'FieldMarketer' THEN CONCAT('Marketing Area: ', FM.MarketingArea)
				WHEN E.Role = 'Accountant' THEN CONCAT('Accounting Field: ', A.AccountingField)
				WHEN E.Role = 'DistributionAgent' THEN CONCAT('Delivery Vehicle: ', DA.DeliveryVehicle)
				ELSE 'No Additional Info'
			END AS Details
		FROM Employee E
		LEFT JOIN Manager M ON E.EmployeeID = M.EmployeeID
		LEFT JOIN Supervisor S ON E.EmployeeID = S.EmployeeID
		LEFT JOIN SalesRepresentative SR ON E.EmployeeID = SR.EmployeeID
		LEFT JOIN Cashier C ON E.EmployeeID = C.EmployeeID
		LEFT JOIN FieldMarketer FM ON E.EmployeeID = FM.EmployeeID
		LEFT JOIN Accountant A ON E.EmployeeID = A.EmployeeID
		LEFT JOIN DistributionAgent DA ON E.EmployeeID = DA.EmployeeID`

	rows, err := shared.DB.Query(query)
	if err != nil {
		http.Error(w, "Error fetching employees: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.EmployeeID, &emp.FirstName, &emp.LastName, &emp.Salary, &emp.PhoneNumber, &emp.Role, &emp.Details); err != nil {
			http.Error(w, "Error scanning employee data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		employees = append(employees, emp)
	}

	tmpl, err := template.ParseFiles("templates/employees.html")
	if err != nil {
		http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, employees)
}

func addEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/addEmployee.html")
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

		employeeID := r.FormValue("employeeID")
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		salary := r.FormValue("salary")
		phoneNumber := r.FormValue("phoneNumber")
		password := r.FormValue("password")
		role := r.FormValue("role")

		query := "INSERT INTO Employee (EmployeeID, FirstName, LastName, Salary, PhoneNumber, Password, Role) VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7)"
		_, err := shared.DB.Exec(query, employeeID, firstName, lastName, salary, phoneNumber, password, role)
		if err != nil {
			http.Error(w, "Error inserting employee: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switch role {
		case "Manager":
			teamSize := r.FormValue("teamSize")
			query = "INSERT INTO Manager (EmployeeID, TeamSize) VALUES (@p1, @p2)"
			_, err = shared.DB.Exec(query, employeeID, teamSize)

		case "Supervisor":
			managerID := r.FormValue("managerID")
			teamSize := r.FormValue("supervisorTeamSize")
			query = "INSERT INTO Supervisor (EmployeeID, ManagerID, TeamSize) VALUES (@p1, @p2, @p3)"
			_, err = shared.DB.Exec(query, employeeID, managerID, teamSize)

		case "SalesRepresentative":
			shiftDuration := r.FormValue("shiftDuration")
			supervisorID := r.FormValue("supervisorID")
			query = "INSERT INTO SalesRepresentative (EmployeeID, ShiftDuration, SupervisorID) VALUES (@p1, @p2, @p3)"
			_, err = shared.DB.Exec(query, employeeID, shiftDuration, supervisorID)

		case "Cashier":
			cashierShiftDuration := r.FormValue("cashierShiftDuration")
			csupervisorID := r.FormValue("csupervisorID")
			query = "INSERT INTO Cashier (EmployeeID, ShiftDuration, SupervisorID) VALUES (@p1, @p2, @p3)"
			_, err = shared.DB.Exec(query, employeeID, cashierShiftDuration, csupervisorID)

		case "FieldMarketer":
			marketingArea := r.FormValue("marketingArea")
			query = "INSERT INTO FieldMarketer (EmployeeID, MarketingArea) VALUES (@p1, @p2)"
			_, err = shared.DB.Exec(query, employeeID, marketingArea)

		case "Accountant":
			accountingField := r.FormValue("accountingField")
			query = "INSERT INTO Accountant (EmployeeID, AccountingField) VALUES (@p1, @p2)"
			_, err = shared.DB.Exec(query, employeeID, accountingField)

		case "DistributionAgent":
			deliveryVehicle := r.FormValue("deliveryVehicle")
			query = "INSERT INTO DistributionAgent (EmployeeID, DeliveryVehicle) VALUES (@p1, @p2)"
			_, err = shared.DB.Exec(query, employeeID, deliveryVehicle)
		}

		if err != nil {
			http.Error(w, "Error inserting role-specific data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/employees", http.StatusSeeOther)
	}
}
