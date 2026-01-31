package cli

import (
	"github.com/google/uuid"
)

// parseUUID memparse string menjadi UUID.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
