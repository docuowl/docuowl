package watch

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/radovskyb/watcher"
)

type Watch struct {
	input        string
	output       string
	port         int
	eventHandler func() error
	cond         *sync.Cond
	watcher      *watcher.Watcher
	lock         *sync.Mutex
}

var websocketUpgrade = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (w Watch) watch() {
	w.watcher.SetMaxEvents(1)
	if err := w.watcher.AddRecursive(w.input); err != nil {
		log.Fatalln(err)
	}
	for {
		select {
		case <-w.watcher.Event:
			if err := w.eventHandler(); err != nil {
				log.Printf("Error performing update: %s.", err)
				log.Printf("Will retry on next update")
				continue
			}
			w.lock.Lock()
			w.cond.Broadcast()
			w.lock.Unlock()
		case err := <-w.watcher.Error:
			log.Fatalln(err)
		case <-w.watcher.Closed:
			return
		}
	}
}

func (w Watch) Run() error {
	go w.watch()
	go func() {
		if err := w.watcher.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", w.handleWSRequest)
	mux.HandleFunc("/", w.handleRequest)
	return http.ListenAndServe(fmt.Sprintf(":%d", w.port), mux)
}

func (w Watch) handleRequest(hw http.ResponseWriter, r *http.Request) {
	f, mimeType, err := w.readFile(r.URL.Path)
	if err != nil {
		if os.IsNotExist(err) {
			hw.WriteHeader(404)
		} else {
			hw.WriteHeader(500)
		}
		_, _ = hw.Write([]byte("Error: " + err.Error()))
		return
	}
	if mimeType != "" {
		hw.Header().Add("Content-Type", mimeType)
	}
	_, _ = hw.Write(f)
}

func (w Watch) readFile(name string) ([]byte, string, error) {
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(w.output, name)
	}
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, "", err
	}
	if stat.IsDir() {
		return w.readFile(filepath.Join(w.output, "index.html"))
	}

	f, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	ext := filepath.Ext(name)
	if ext == ".html" {
		rawFile := string(f) +
			fmt.Sprintf("<script type=\"data/owl-reload-ws\">ws://127.0.0.1:%d/ws</script>", w.port) +
			`
	<script type="text/javascript">
		(function() {
    let ws = new WebSocket(document.querySelector("script[type='data/owl-reload-ws']").innerHTML);
    ws.onmessage = function (event) {
        window.location.reload();
    }
})();</script>`
		f = []byte(rawFile)
	}

	mimeType := mime.TypeByExtension(ext)
	return f, mimeType, nil
}

func (w Watch) handleWSRequest(hw http.ResponseWriter, r *http.Request) {
	c, err := websocketUpgrade.Upgrade(hw, r, nil)
	if err != nil {
		log.Println("Error upgrading request:", err)
		return
	}
	defer c.Close()
	for {
		w.lock.Lock()
		w.cond.Wait()
		err = c.WriteMessage(websocket.TextMessage, []byte("{}"))
		w.lock.Unlock()
		if err != nil {
			log.Println("Error writing socket:", err)
			break
		}
	}
}

func New(input, output string, port int, handler func() error) *Watch {
	lock := sync.Mutex{}
	watch := &Watch{
		input:        input,
		output:       output,
		port:         port,
		watcher:      watcher.New(),
		eventHandler: handler,
		lock:         &lock,
		cond:         sync.NewCond(&lock),
	}

	return watch
}
