package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type webRequest struct {
	r      *http.Request
	w      http.ResponseWriter
	doneCh chan struct{}
}

var (
	requestCh    = make(chan *webRequest)
	registerCh   = make(chan string)
	unregisterCh = make(chan string)
	heartbeartCh = time.Tick(5 * time.Second)
)

var (
	transport = http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
)

func init() {
	http.DefaultClient = &http.Client{Transport: &transport}
}

func main() {
	flag.Parse()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		doneCh := make(chan struct{})
		requestCh <- &webRequest{r: r, w: w, doneCh: doneCh}
		<-doneCh
	})

	go processRequests()

	go http.ListenAndServeTLS(":2000", "cert.pem", "key.pem", nil)

	go http.ListenAndServeTLS(":2001", "cert.pem", "key.pem", new(appserverHandler))

	log.Println("Server started, press <ENTER> to exit")
	fmt.Scanln()
}

type appserverHandler struct{}

func (h *appserverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	port := r.URL.Query().Get("port")
	switch r.URL.Path {
	case "/register":
		registerCh <- ip + ":" + port
	case "/unregister":
		unregisterCh <- ip + ":" + port
	}
}
