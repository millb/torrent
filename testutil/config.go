package testutil

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
)

type Config struct {
	Peers    []string
	FileInfo *FileInfo
}

func (c Config) Encode() []byte {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Panic("failed to encode json", err)
	}
	return b
}

type FileInfo struct {
	Size     int
	PartSize int
	Parts    []Hash
}

// Default hash is SHA-1.
type Hash [sha1.Size]byte

// Convert bytes to hex.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h *Hash) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", h.String())), nil
}
