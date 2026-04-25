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
		depth     int
		branching int
	}{
		{4, 10}, // ~11K dirs
		{5, 10}, // ~111K dirs
	}

	for _, tt := range tests {
		b.Run(fmt.Sprintf("depth=%d_branching=%d", tt.depth, tt.branching), func(b *testing.B) {
			root, err := generateTestDirHierarchy(b, tt.depth, tt.branching)
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

func generateTestDirHierarchy(t testing.TB, depth int, branching int) (string, error) {
	rootDir := t.TempDir()
	var build func(dir string, d int) error
	build = func(dir string, d int) error {
		if d == 0 {
			return nil
		}
		for range branching {
			child := path.Join(dir, uuid.NewString()[:8])
			if err := os.Mkdir(child, 0750); err != nil {
				return err
			}
			if err := build(child, d-1); err != nil {
				return err
			}
		}
		return nil
	}
	return rootDir, build(rootDir, depth)
}
