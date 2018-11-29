package gatewayd

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var routingTable = make(map[string]string)
var rexp, _ = regexp.Compile("\"faas.*=.*\"")

func route(w http.ResponseWriter, r *http.Request) {
	subRoute := strings.Split(r.RequestURI, "/lambda/")[1]
	port, exists := routingTable[strings.Split(subRoute, "?")[0]]

	if (exists) {
		http.Redirect(w, r, "localhost:" + port + "/" + subRoute, http.StatusUseProxy)
	}
}

func updateRoutes() {
	for {
		b, err := ioutil.ReadFile("docker-compose.yml")
		if err != nil {
			continue
		}

		labels := rexp.FindAllString(string(b), 100)

		tempTable := make(map[string]string)
		for i := 0; i < len(labels); i += 2 {
			name := strings.Replace(strings.Split(labels[i], "=")[1], "\"", "", 1)
			port := strings.Replace(strings.Split(labels[i + 1], "=")[1], "\"", "", 1)
			tempTable[name] = port
		}

		routingTable = tempTable

		time.Sleep(1 * time.Second)
	}
}

func main() {
	go updateRoutes()
	http.HandleFunc("/lambda", route)
	http.ListenAndServe(":80", nil)
}