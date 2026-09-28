// Command base is the size baseline probe: a bare net/http server with no
// third-party dependencies. Use it to measure the floor of a static Go binary.
package main

import (
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe("127.0.0.1:18080", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
}
