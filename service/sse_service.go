package service

import (
	"fmt"
	"net/http"
)

func SendSSE(w http.ResponseWriter, f http.Flusher, event string, data map[string]any) {

	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)

	f.Flush()
}
