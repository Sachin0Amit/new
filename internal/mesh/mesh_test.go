package mesh

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestKnowledgeMesh(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "mesh-test-*")
	defer os.RemoveAll(tmpDir)

	m, err := NewKnowledgeMesh(tmpDir)
	assert.NoError(t, err)
	defer m.Close()

	ctx := context.Background()

	t.Run("Store and Retrieve", func(t *testing.T) {
		type data struct {
			Val string
		}
		
		tests := []struct {
			name    string
			key     string
			val     data
			wantErr bool
		}{
			{"Basic String", "key1", data{"hello"}, false},
			{"Valid Key 2", "key2", data{"no"}, false}, // Badger allows empty keys
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := m.Store(ctx, tt.key, tt.val)
				assert.NoError(t, err)

				var out data
				err = m.Retrieve(ctx, tt.key, &out)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.val, out)
				}
			})
		}
	})

	t.Run("Query", func(t *testing.T) {
		_ = m.Store(ctx, "prefix:1", "data1")
		_ = m.Store(ctx, "prefix:2", "data2")
		_ = m.Store(ctx, "other:1", "data3")

		results, err := m.Query(ctx, "prefix:", 10)
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
}
