package ihttp

import (
    "fmt"
    "net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello")
}

func StartServer(listenOn string,) {
    http.HandleFunc("/", index)
    http.ListenAndServe(listenOn, nil)
}