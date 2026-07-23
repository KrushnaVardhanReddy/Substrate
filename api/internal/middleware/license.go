package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"os"
	"strings"
	"time"
	"log"

	"github.com/KrushnaVardhanReddy/substrate/engine/licensing"
	"github.com/spf13/viper"
)

type contextKey string

const (
	FeaturesKey contextKey = "features"
)

func LicenseMiddleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	var cachedClaims *licensing.LicenseClaims
	var loaded bool
	var loadErr error

	licensePath := viper.GetString("SUBSTRATE_LICENSE_FILE")
	if licensePath == "" {
		licensePath = os.Getenv("SUBSTRATE_LICENSE_FILE")
	}

	if licensePath != "" {
		validator := licensing.NewValidator(publicKey)
		cachedClaims, loadErr = validator.ValidateLicenseFile(licensePath)
		if loadErr != nil {
			log.Printf("License load error during init, falling back to Free Tier: %v\n", loadErr)
		}
		loaded = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			features := []string{}

			if loaded && loadErr == nil && cachedClaims != nil {
				// Only re-check expiration on each request, not the signature/disk read
				if cachedClaims.ExpiresAt != nil && cachedClaims.ExpiresAt.Time.After(time.Now()) {
					// Copy slice to avoid mutating cached state in downstream handlers
					features = make([]string, len(cachedClaims.Features))
					copy(features, cachedClaims.Features)
				}
			}

			// Add features to context
			ctx = context.WithValue(ctx, FeaturesKey, features)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CheckFeature is a helper to check if a feature is enabled in the current context
func CheckFeature(ctx context.Context, feature string) bool {
	features, ok := ctx.Value(FeaturesKey).([]string)
	if !ok {
		return false
	}

	for _, f := range features {
		if strings.EqualFold(f, feature) {
			return true
		}
	}
	return false
}
