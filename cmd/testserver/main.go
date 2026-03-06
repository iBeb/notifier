package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	status := flag.Int("status", http.StatusNoContent, "response status code")
	delay := flag.Duration("delay", 0, "response delay")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: testserver [flags]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	
	http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		log.Printf("method=%s path=%s body=%q", r.Method, r.URL.Path, string(body))

		if *delay > 0 {
			time.Sleep(*delay)
		}

		w.WriteHeader(*status)
	})
	log.Fatal(http.ListenAndServe(":8081", nil))
}
