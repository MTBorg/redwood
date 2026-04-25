package walker

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func BenchmarkWalker(b *testing.B) {
	tests := []struct {
		levels int
		width  int
	}{
		{10, 1000},
		{20, 10000},
		// {30, 20000},
	}

	for _, tt := range tests {
		b.Run(fmt.Sprintf("levels=%d_width=%d", tt.levels, tt.width), func(b *testing.B) {
			root, err := generateTestDirHierarchy(b, tt.levels, tt.width)
			assert.NoError(b, err)
			rootFS := os.DirFS(root)

			for b.Loop() {
				fs.WalkDir(rootFS, ".", func(path string, d fs.DirEntry, err error) error {
					return nil
				})
			}
			b.ReportMetric(b.Elapsed().Seconds()/float64(b.N), "seconds/op")
		})
	}

}

// func TestWalker(t *testing.T) {
// 	fmt.Println("Root directory:", rootDir)

// 	for range 1000 {
// 		p := generatePath(5)
// 		p = path.Join(rootDir, p)
// 		err := os.MkdirAll(p, 0750)
// 		require.NoError(t, err)
// 	}
// 	t.Log("Directories created. Starting walker...")

// 	time.Sleep(1000 * time.Second)
// }

func generateTestDirHierarchy(t testing.TB, levels int, width int) (string, error) {
	rootDir := t.TempDir()
	for range width {
		p := generatePath(levels)
		p = path.Join(rootDir, p)
		err := os.MkdirAll(p, 0750)
		if err != nil {
			return "", fmt.Errorf("os.MkdirAll: %w", err)
		}
	}
	return rootDir, nil
}

func generatePath(levels int) string {
	components := []string{}
	for range levels {
		component := uuid.NewString()[:8]
		components = append(components, component)
	}
	return path.Join(components...)
}
