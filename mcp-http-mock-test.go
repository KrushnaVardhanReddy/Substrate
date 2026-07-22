//go:build ignore

package main

import (
    "fmt"
    "net/http"
)

func main() {
    client := &http.Client{}
    resp, err := client.Get("http://example.com")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(resp.StatusCode)
}
