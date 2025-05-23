package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("OpenAPI 3.0 specification is manually maintained.")
	
	// Check if the openapi.yml file exists
	if _, err := os.Stat("../server/openapi.yml"); os.IsNotExist(err) {
		log.Fatal("Error: openapi.yml not found. The OpenAPI 3.0 spec should be manually maintained.")
	}
	
	fmt.Println("✅ OpenAPI 3.0 specification is ready:")
	fmt.Println("  - OpenAPI 3.0: internal/server/openapi.yml")
}
