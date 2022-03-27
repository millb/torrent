package testutil

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"strings"
)

type Generator struct {
	r *rand.Rand
}

func NewGenerator(seed int64) *Generator {
	return &Generator{
		r: rand.New(rand.NewSource(seed)),
	}
}

func (g *Generator) Generate(w io.Writer, size int, partSize int) ([]Hash, error) {
	var hashes []Hash
	buf := make([]byte, partSize)

	fullParts := size / partSize
	for i := 0; i < fullParts; i++ {
		hash, err := g.generatePart(w, buf)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}

	rem := size % partSize
	if rem > 0 {
		hash, err := g.generatePart(w, buf[:rem])
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func (g *Generator) generatePart(w io.Writer, buf []byte) (Hash, error) {
	g.Read(buf)
	_, err := w.Write(buf)
	if err != nil {
		return Hash{}, err
	}

	return sha1.Sum(buf), nil
}

func (g *Generator) GenerateIPs(count int, subnet string) []string {
	if !strings.HasSuffix(subnet, "/24") {
		panic("subnet must be /24")
	}

	parts := strings.Split(subnet, ".")
	var ips []string
	for i := 2; i < 255; i++ {
		ips = append(ips, fmt.Sprintf("%s.%s.%s.%d", parts[0], parts[1], parts[2], i))
	}

	g.Shuffle(ips)

	return ips[:count]
}

func (g *Generator) Shuffle(arr []string) {
	g.r.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

func (g *Generator) Intn(n int) int {
	return g.r.Intn(n)
}

func (g *Generator) Read(p []byte) {
	_, _ = g.r.Read(p)
}
