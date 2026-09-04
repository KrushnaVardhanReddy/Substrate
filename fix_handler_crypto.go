package main

import (
	"io/ioutil"
	"strings"
)

func main() {
	content, _ := ioutil.ReadFile("api/internal/handlers/airbyte_handler.go")
	newContent := strings.ReplaceAll(string(content), "crypto.NewMockKMSClient(keyARN)", "crypto.NewKMSClient(keyARN)")
	ioutil.WriteFile("api/internal/handlers/airbyte_handler.go", []byte(newContent), 0644)

	testContent, _ := ioutil.ReadFile("api/internal/handlers/airbyte_handler_test.go")
	newTestContent := strings.ReplaceAll(string(testContent), "crypto.NewMockKMSClient(validKeyARN)", "crypto.NewKMSClient(validKeyARN)")
	ioutil.WriteFile("api/internal/handlers/airbyte_handler_test.go", []byte(newTestContent), 0644)
}
