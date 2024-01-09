package server

import (
	"fmt"
	"log"
	"net/http"
)

func NewServer() {
	println("hello server")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "couldn't parse form", 500)
			return
		}
		fmt.Fprintf(w, "form values: %+v", r.Form)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
