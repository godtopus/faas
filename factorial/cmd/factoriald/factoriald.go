package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Response struct {
	Data uint64 `json:"data"`
}

func main() {
	http.HandleFunc("/lambda/factorial", factorial)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func factorial(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["n"]

	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Url Param 'n' is missing", http.StatusInternalServerError)
		return
	}

	n, err := strconv.ParseUint(keys[0], 10, 64)
	if err != nil {
		http.Error(w, "Error converting argument", http.StatusInternalServerError)
	}

	resp := Response{fac(n) }

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func fac(n uint64) uint64 {
	if n == 0 {
		return 1
	}

	return n * fac(n-1)
}