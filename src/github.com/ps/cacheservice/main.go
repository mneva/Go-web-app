package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type cacheEntry struct {
	data       []byte
	expiration time.Time
}

var (
	cache  = make(map[string]*cacheEntry)
	mutex  = sync.RWMutex{}
	tickCh = time.Tick(5 * time.Second)
)

var maxAgeRexexp = regexp.MustCompile(`maxage=(\d+)`)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getFromCache(w, r)
		} else if r.Method == http.MethodPost {
			saveToCache(w, r)
		}
	})

	http.HandleFunc("/invalidate", invalidateEntry)

	go http.ListenAndServeTLS(":5000", "cert.pem", "key.pem", nil)

	go purgeCache()

	log.Println("Caching service started, press <ENTER> to exit")
	fmt.Scanln()
}

func getFromCache(w http.ResponseWriter, r *http.Request) {
	mutex.RLock()
	defer mutex.RUnlock()

	key := r.URL.Query().Get("key")

	fmt.Printf("Searching cache for %s...", key)
	if entry, ok := cache[key]; ok {
		fmt.Println("found")
		w.Write(entry.data)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Println("not found")
}
func saveToCache(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	key := r.URL.Query().Get("key")
	cacheHeader := r.Header.Get("cache-control")

	fmt.Printf("Saving cache entry with key '%s' for %s seconds\n", key, cacheHeader)

	matches := maxAgeRexexp.FindStringSubmatch(cacheHeader)
	if len(matches) == 2 {
		dur, _ := strconv.Atoi(matches[1])
		data, _ := ioutil.ReadAll(r.Body)
		cache[key] = &cacheEntry{data: data, expiration: time.Now().Add(time.Duration(dur) * time.Second)}
	}
}

func invalidateEntry(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	key := r.URL.Query().Get("key")
	fmt.Printf("purging entry with key '%s'\n", key)
	delete(cache, key)
}

func purgeCache() {
	for range tickCh {
		mutex.Lock()
		now := time.Now()

		fmt.Println("purging cache")

		for k, v := range cache {
			if now.Before(v.expiration) {
				fmt.Printf("purging entry with key '%s'\n", k)
				delete(cache, k)
			}
		}
		mutex.Unlock()
	}
}
