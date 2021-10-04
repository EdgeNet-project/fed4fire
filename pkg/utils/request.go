package utils

import (
	"fmt"
	"net/http"
)

func RequestId(r *http.Request) string {
	return fmt.Sprintf("%p", r)
}
