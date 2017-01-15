package main

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ps/entity"
	"github.com/ps/logservice/loghelper"
)

var (
	appservers   = []string{}
	currentIndex = 0
	client       = http.Client{Transport: &transport}
)

func processRequests() {
	for {
		select {
		case request := <-requestCh:
			println("request")
			if len(appservers) == 0 {
				request.w.WriteHeader(http.StatusInternalServerError)
				request.w.Write([]byte("No app servers found"))
				request.doneCh <- struct{}{}
				continue
			}
			currentIndex++
			if currentIndex == len(appservers) {
				currentIndex = 0
			}
			host := appservers[currentIndex]
			go processRequest(host, request)
		case host := <-registerCh:
			println("register " + host)
			go loghelper.WriteEntry(&entity.LogEntry{
				Level:     entity.LogLevelInfo,
				Timestamp: time.Now(),
				Source:    "load balancer",
				Message:   "Registering application server with address: " + host,
			})
			isFound := false
			for _, h := range appservers {
				if host == h {
					isFound = true
					break
				}
			}

			if !isFound {
				appservers = append(appservers, host)
			}
		case host := <-unregisterCh:
			println("unregister " + host)
			go loghelper.WriteEntry(&entity.LogEntry{
				Level:     entity.LogLevelInfo,
				Timestamp: time.Now(),
				Source:    "load balancer",
				Message:   "Unregistering application server with address: " + host,
			})
			for i := len(appservers) - 1; i >= 0; i-- {
				if appservers[i] == host {
					appservers = append(appservers[:i], appservers[i+1:]...)
				}
			}
		case <-heartbeartCh:
			println("heartbeat")
			servers := appservers[:]
			go func(servers []string) {
				for _, server := range servers {
					resp, err := http.Get("https://" + server + "/ping")
					if err != nil || resp.StatusCode != 200 {
						unregisterCh <- server
					}
				}
			}(servers)
		}
	}
}

func processRequest(host string, request *webRequest) {
	hostURL, _ := url.Parse(request.r.URL.String())
	hostURL.Scheme = "https"
	hostURL.Host = host
	println(host)
	println(hostURL.String())
	req, _ := http.NewRequest(request.r.Method, hostURL.String(), request.r.Body)
	for k, v := range request.r.Header {
		values := ""
		for _, headerValue := range v {
			values += headerValue + " "
		}
		req.Header.Add(k, values)
	}

	resp, err := client.Do(req)

	if err != nil {
		request.w.WriteHeader(http.StatusInternalServerError)
		request.doneCh <- struct{}{}
		return
	}

	for k, v := range resp.Header {
		values := ""
		for _, headerValue := range v {
			values += headerValue + " "
		}
		request.w.Header().Add(k, values)
	}
	io.Copy(request.w, resp.Body)

	request.doneCh <- struct{}{}
}
