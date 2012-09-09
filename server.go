/*
A simple gui toolkit to communicate between html/javascript and go server.
On execution it opens a browser connecting it to the server.
*/
package hgui

import (
	"io"
	"log"
	"net"
	"net/http"
	"sync"
)

var resourceLock sync.RWMutex
var resources = make(map[string][]byte)

func RegisterResource(path string, data []byte) {
	resourceLock.Lock()
	defer resourceLock.Unlock()

	resources[path] = data
}

// Deprecated. Use RegisterResource.
func SetResource(files map[string][]byte) {
	resourceLock.Lock()
	defer resourceLock.Unlock()

	for path, data := range files {
		resources[path] = data
	}
}

var (
	handlers = make(map[string]func())

	TopFrame = &Frame{
		widget:   newWidget(),
		content:  make([]HTMLer, 0, 20),
		topframe: true,
	}

	// Legacy
	Topframe = TopFrame
)

func StartServer(width, height int, title string) {
	// Port 0 is a special case that chooses a random available port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	addr := ln.Addr()

	go func() {
		err := http.Serve(ln, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	startGui(width, height, title, addr.String())
}

func requests(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	switch req.URL.Path {
	case "/events":
		eventPoll(w)

	case "/reply":
		eventReply(reply{query.Get("id"), query.Get("reply")})

	case "/handler":
		if handler, ok := handlers[query.Get("id")]; ok {
			handler()
		}

	case "/":
		io.WriteString(w, head)
		io.WriteString(w, TopFrame.HTML())
		io.WriteString(w, foot)

	case "/js":
		w.Header().Set("Content-Type", "application/javascript")
		w.Write(fileJQuery)
		io.WriteString(w, "\n\n")
		w.Write(filecorejs)

	default:
		resourceLock.RLock()
		if data, ok := resources[req.URL.Path]; ok {
			w.Write(data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "404 - Not Found")
		}
		resourceLock.RUnlock()
	}
}

func init() {
	http.HandleFunc("/", requests)
}

const (
	head = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script src="js"></script>
</head>
`
	foot = `
</html>
`
)
