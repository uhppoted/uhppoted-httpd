package cards

import (
	"log"

	"github.com/uhppoted/uhppoted-httpd/types"
)

const GZIP_MINIMUM = 16384

func warn(err error) {
	switch v := err.(type) {
	case *types.HttpdError:
		log.Printf("%-5s %v", "WARN", v.Detail)

	default:
		log.Printf("%-5s %v", "WARN", v)
	}
}
