package cardholders

import (
	"log"

	"github.com/uhppoted/uhppoted-httpd/types"
)

func warn(err error) {
	switch v := err.(type) {
	case *types.HttpdError:
		log.Printf("%-5s %v", "WARN", v.Detail)

	default:
		log.Printf("%-5s %v", "WARN", v)
	}
}
