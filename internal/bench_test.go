package internal

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sachin0Amit/new/internal/mesh"
)

// BenchmarkKnowledgeMesh measures BadgerDB read/write throughput.
func BenchmarkKnowledgeMesh(b *testing.B) {
	tmpDir, _ := ioutil.TempDir("", "mesh-bench-*")
	defer os.RemoveAll(tmpDir)

	m, _ := mesh.NewKnowledgeMesh(tmpDir)
	defer m.Close()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := "key-" + string(rune(i))
		_ = m.Store(ctx, key, "some large value content for benchmarking")
		var out string
		_ = m.Retrieve(ctx, key, &out)
	}
}

// BenchmarkCryptoLatency measures ED25519 sign + verify cycle.
func BenchmarkCryptoLatency(b *testing.B) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	msg := []byte("Sovereign Intelligence Core - Secure Handshake")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sig := ed25519.Sign(priv, msg)
		_ = ed25519.Verify(pub, msg, sig)
	}
}
