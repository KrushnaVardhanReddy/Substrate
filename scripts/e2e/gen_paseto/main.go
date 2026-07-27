package main

import (
	"crypto/sha256"
	"fmt"
	"time"
	"aidanwoods.dev/go-paseto"
)

func main() {
	secret := "local-jwt-secret"
	hash := sha256.Sum256([]byte(secret))
	key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
	if err != nil {
		panic(err)
	}

	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(24 * time.Hour))
	err = token.Set("orgs", map[string]string{
		"admin":       "admin",
		"adminuser":   "admin",
		"mcp-org":     "admin",
		"testorg":     "admin",
		"p3-org":      "admin",
		"stress-test": "admin",
	})
	if err != nil {
		panic(err)
	}

	fmt.Print(token.V4Encrypt(key, nil))
}
