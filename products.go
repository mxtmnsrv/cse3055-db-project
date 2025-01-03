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
