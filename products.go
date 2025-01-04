package main

import (
	"dbs-term-project/shared"
	"html/template"
	"net/http"
)

type Product struct {
	ProductID   int
	Name        string
	Category    string
	Price       float64
	Description string
}

func viewProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := shared.DB.Query("SELECT ProductID, Name, Category, Price, Description FROM Product")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ProductID, &product.Name, &product.Category, &product.Price, &product.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	tmpl, _ := template.ParseFiles("templates/products.html")
	tmpl.Execute(w, products)
}

func viewProductsInWarehouse(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Fetch warehouses to populate the dropdown
		rows, err := shared.DB.Query("SELECT WarehouseID, Location FROM Warehouse")
		if err != nil {
			http.Error(w, "Error fetching warehouses: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var warehouses []struct {
			WarehouseID int
			Location    string
		}
		for rows.Next() {
			var warehouse struct {
				WarehouseID int
				Location    string
			}
			if err := rows.Scan(&warehouse.WarehouseID, &warehouse.Location); err != nil {
				http.Error(w, "Error scanning warehouse data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			warehouses = append(warehouses, warehouse)
		}

		// Parse and render the template with warehouse data
		tmpl, err := template.ParseFiles("templates/warehouse.html")
		if err != nil {
			http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			Warehouses []struct {
				WarehouseID int
				Location    string
			}
			Products []struct {
				ProductID int
				Name      string
				Quantity  int
			}
		}{
			Warehouses: warehouses,
			Products:   nil,
		})
		return
	}

	if r.Method == http.MethodPost {
		// Fetch WarehouseID by its Location
		r.ParseForm()
		location := r.FormValue("location")
		var warehouseID int
		err := shared.DB.QueryRow("SELECT WarehouseID FROM Warehouse WHERE Location = @p1", location).Scan(&warehouseID)
		if err != nil {
			http.Error(w, "Error finding warehouse: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Fetch products in the selected warehouse
		query := `
			SELECT p.ProductID, p.Name, wp.Quantity
			FROM WarehouseProduct wp
			JOIN Product p ON wp.ProductID = p.ProductID
			WHERE wp.WarehouseID = @p1
		`
		rows, err := shared.DB.Query(query, warehouseID)
		if err != nil {
			http.Error(w, "Error fetching products: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var products []struct {
			ProductID int
			Name      string
			Quantity  int
		}
		for rows.Next() {
			var product struct {
				ProductID int
				Name      string
				Quantity  int
			}
			if err := rows.Scan(&product.ProductID, &product.Name, &product.Quantity); err != nil {
				http.Error(w, "Error scanning product data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			products = append(products, product)
		}

		// Parse and render the template with product data
		tmpl, err := template.ParseFiles("templates/warehouse.html")
		if err != nil {
			http.Error(w, "Error loading template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			Warehouses []struct {
				WarehouseID int
				Location    string
			}
			Products []struct {
				ProductID int
				Name      string
				Quantity  int
			}
		}{
			Warehouses: nil, // No need to reload warehouses after selection
			Products:   products,
		})
	}
}

func addProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, _ := template.ParseFiles("templates/addProduct.html")
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		// productID := r.FormValue("productID")
		name := r.FormValue("name")
		category := r.FormValue("category")
		price := r.FormValue("price")
		description := r.FormValue("description")

		// SQL query to insert product
		query := "INSERT INTO Product (Name, Category, Price, Description) VALUES (@p1, @p2, @p3, @p4)"
		_, err := shared.DB.Exec(query, name, category, price, description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/products", http.StatusSeeOther)
	}
}
