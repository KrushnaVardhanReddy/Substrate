package main

import (
	"os"
	"io/ioutil"
	"strings"
)

func main() {
	b, _ := ioutil.ReadFile("scripts/e2e/helpers/fix.go")
	s := string(b)
	s = strings.ReplaceAll(s, `fmt.Sprintf("http://localhost:%s/api/v1", cmp.Or(os.Getenv("FORGEJO_PORT"), "3000"))`, `func() string { if p := os.Getenv("FORGEJO_PORT"); p != "" { return "http://localhost:" + p + "/api/v1" } else { return "http://localhost:3000/api/v1" } }()`)
	ioutil.WriteFile("scripts/e2e/helpers/fix.go", []byte(s), 0644)
}
