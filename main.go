package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/lesismal/nbio/nbhttp"
)

func getPort() string {
	port, ok := os.LookupEnv("PORT")
	if ok {
		return port
	}

	return "8080"
}

var (
	qps   uint64 = 0
	total uint64 = 0
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func middlewareContext(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("middlewareContext")

		atomic.AddUint64(&qps, 1)
		next.ServeHTTP(w, r)
	})
}

func middlewareLogger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Println(r.URL.Path, fmt.Sprintf("%d ms", time.Since(start).Milliseconds()))
	})
}

func main() {
	flag.Parse()

	mux := &http.ServeMux{}

	// mux := mux.NewRouter()
	// apiRouter(mux.PathPrefix("/api").Subrouter())
	mux.HandleFunc("/", middlewareLogger(http.HandlerFunc(serveHome)))
	mux.HandleFunc("/ws", middlewareLogger(onWebsocket))

	// Run our server in a goroutine so that it doesn't block.
	svr := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{"localhost:" + getPort()},
		Handler: mux,
	}) // pool.Go)

	err := svr.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	delta := 5
	ticker := time.NewTicker(time.Second * time.Duration(delta))
	for i := 1; true; i++ {
		<-ticker.C
		n := atomic.SwapUint64(&qps, 0)
		total += n
		log.Printf("running for %v seconds, NumGoroutine: %v, qps: %v, total: %v\n", i*delta, runtime.NumGoroutine(), n/uint64(delta), total)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	svr.Shutdown(ctx)
}
