// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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
