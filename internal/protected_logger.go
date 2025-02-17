package internal

import (
	"log"

	"github.com/kvizdos/typesend/pkg/typesend_schemas"
)

func ProtectedLogger(logger typesend_schemas.Logger, format string, args ...any) {
	if logger != nil {
		logger.Errorf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
