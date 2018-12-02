package main

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Handler struct {
	http.HandlerFunc
	Enabled bool
}

type Gateway struct {
	Client		*client.Client
	Handlers	map[string]*Handler
	*sync.Mutex
}

func NewGateway(cli *client.Client) *Gateway {
	return &Gateway{
		Client:		cli,
		Handlers:	map[string]*Handler{},
		Mutex:		&sync.Mutex{},
	}
}

func (gw *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if handler, ok := gw.Handlers[path]; ok && handler.Enabled {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func Handle(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("request:", r.RemoteAddr, "want", r.RequestURI, "from", r.Host)
		p.ServeHTTP(w, r)
	}
}

func (gw *Gateway) HandleFunc(pattern string, handler http.HandlerFunc) {
	gw.Lock()
	defer gw.Unlock()

	gw.Handlers[pattern] = &Handler{handler, true}
	http.HandleFunc(pattern, handler)
}

func NewReverseProxy(target string) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   target,
	})
}

func (gw *Gateway) reload(cli *client.Client) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
		return
	}

	gw.Lock()
	defer gw.Unlock()

	for _, handler := range gw.Handlers {
		handler.Enabled = false
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.Labels["faas.name"], container.Labels["faas.port"])

		name := container.Labels["faas.name"]
		port := container.Labels["faas.port"]
		if name == "" {
			continue
		}

		targetUrl, err := url.Parse(fmt.Sprintf("localhost:%v", port))
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Printf("%s\n", targetUrl.String())
		//mux := http.NewServeMux()
		//mux.HandleFunc("/lambda/" + name, Handle(NewReverseProxy(targetUrl.String())))
		http.HandleFunc("/lambda/" + name, Handle(NewReverseProxy(targetUrl.String())))
	}
}

func (gw *Gateway) listen(cli *client.Client) {
	filter := filters.NewArgs()
	filter.Add("type", "container")
	filter.Add("event", "start")
	filter.Add("event", "stop")
	filter.Add("event", "destroy")
	filter.Add("event", "kill")
	filter.Add("event", "die")

	msg, errChan := cli.Events(context.Background(), types.EventsOptions {
		Filters: filter,
	})

	for {
		select {
		case err := <-errChan:
			log.Fatal(err)
		case <-msg:
			gw.reload(cli)
		}
	}
}

func main() {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.39"))
	if err != nil {
		panic(err)
	}

	gateway := NewGateway(cli)

	gateway.reload(cli)

	go gateway.listen(cli)

	log.Fatal(http.ListenAndServe(":80", nil))
}