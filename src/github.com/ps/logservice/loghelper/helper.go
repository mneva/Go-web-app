package loghelper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"net/http"

	"github.com/ps/entity"
)

var logserviceURL = flag.String("logservice", "https://172.18.0.14:5000",
	"Address of the logging service")

var tr = http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}
var client = &http.Client{Transport: &tr}

func WriteEntry(entry *entity.LogEntry) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	enc.Encode(entry)
	req, _ := http.NewRequest(http.MethodPost, *logserviceURL, &buf)
	client.Do(req)
}
