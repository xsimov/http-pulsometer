package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	go StartPulsometer()

	ch := make(chan string)

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		go HeartbeatReceived()
		for {
			select {
			case e := <-ch:
				fmt.Fprintf(w, "%s\n", e)
			case <-time.After(50 * time.Millisecond):
				w.WriteHeader(204)
				return
			}
		}
	})

	http.HandleFunc("/lights", func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query()["action"][0]
		go func() {
			ch <- fmt.Sprintf(`{"type": "lights", "action": "%v"}`, action)
		}()
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
