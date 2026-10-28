package main

import (
	"github.com/dirmich/marubot/pkg/adminruntime"
	"log"
)

func main() {
	if err := adminruntime.Run(); err != nil {
		log.Fatal(err)
	}
}
