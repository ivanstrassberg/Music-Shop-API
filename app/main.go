package main

import (
	"fmt"
	api "musicShopBackend/api"
	database "musicShopBackend/database"
)

func main() {
	storage, err := database.NewPostgresStorage()
	if err != nil {
		fmt.Println(err)
		return
	}
	server := api.NewAPIServer(":8080", storage)
	server.Run()
}
