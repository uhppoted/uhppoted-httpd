package cardholders

import (
	"log"
)

func warn(err error) {
	log.Printf("%-5s %v", "WARN", err)
}
