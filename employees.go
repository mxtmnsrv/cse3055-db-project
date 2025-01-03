package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
)

type Employee struct {
	EmployeeID  int
	FirstName   string
	LastName    string
	Salary      float64
	PhoneNumber string
	Role        string
	RoleDetails interface{} // This will hold the role-specific detail (e.g., ShiftDuration, TeamSize, etc.)
	// Role-specific fields (using interface{} for dynamic handling)
	ShiftDuration   int
	TeamSize        int
	MarketingArea   string
	DeliveryVehicle string
	AccountingField string
	ManagerID       int
	SupervisorID    int
}

func viewEmployees(w http.ResponseWriter, r *http.Request) {
	rows, err := shared.DB.Query(`
        SELECT EmployeeID, FirstName, LastName, Salary, PhoneNumber, Role 
        FROM Employee;
    `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.EmployeeID, &emp.FirstName, &emp.LastName, &emp.Salary, &emp.PhoneNumber, &emp.Role); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		employees = append(employees, emp)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/employees.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, employees)
}

func addEmployee(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// show html form to add employee
		tmpl, err := template.ParseFiles("templates/addEmployee.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		// getting data
		employeeID := r.FormValue("employeeID")
		firstName := r.FormValue("firstName")
		lastName := r.FormValue("lastName")
		salary := r.FormValue("salary")
		phoneNumber := r.FormValue("phoneNumber")
		role := r.FormValue("role")

		// checking fields
		if employeeID == "" || firstName == "" || lastName == "" || salary == "" || phoneNumber == "" || role == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// inserting employee sql query
		query := `
            INSERT INTO Employee (EmployeeID, FirstName, LastName, Salary, PhoneNumber, Role)
            VALUES (@p1, @p2, @p3, @p4, @p5, @p6)
        `
		_, err := shared.DB.Exec(query, employeeID, firstName, lastName, salary, phoneNumber, role)
		if err != nil {
			http.Error(w, "Error inserting employee: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// redirect to list of employees
		http.Redirect(w, r, "/employees", http.StatusSeeOther)
	}
}
