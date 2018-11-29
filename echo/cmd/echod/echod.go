package main

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data string `json:"data"`
}

func main() {
	http.HandleFunc("/echo", echo)
	http.ListenAndServe(":8080", nil)
}

func echo(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["msg"]

	if !ok || len(keys[0]) < 1 {
		http.Error(w, "Url Param 'msg' is missing", http.StatusInternalServerError)
		return
	}

	resp := Response{keys[0] }

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

