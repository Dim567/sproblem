package main

import (
	"fmt"
	"log"

	"github.com/Dim567/sproblem/server"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting the server...")
	err := server.Start("mock config")
	if err != nil {
		log.Println(fmt.Errorf("server error"), err)
	}
}
