package main

import (
	"construction-store/db"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Подключение к базе данных
	connString := "server=localhost;user id=your_user;password=your_password;database=ConstructionStoreDB"
	db.InitDB(connString)

	// Настройка маршрутов
	router := mux.NewRouter()

	router.HandleFunc("/customers", GetCustomers).Methods("GET")
	router.HandleFunc("/customers", AddCustomer).Methods("POST")

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetCustomers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("List of customers"))
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Add a new customer"))
}
