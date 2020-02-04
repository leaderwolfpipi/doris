// test cors
package middleware

import (
	"net/http"
	"testing"

	"github.com/leaderwolfpipi/doris"
)

// Test cors in doris
func TestDorisCors(t *testing.T) {
	d := doris.New()
	// d.Debug = true

	handler := func(c *doris.Context) error {
		c.String(http.StatusOK, "test")
		return nil
	}

	// add middleware
	// d.Use(JWT([]byte("secret")))
	d.Use(Cors())
	d.Use(Logger())
	d.Use(Recovery())

	d.GET("/", handler)

	d.Run("localhost:9528")
}
