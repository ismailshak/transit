// Package fixtures resolves paths under the module's top-level testdata
// directory.
//
// Should not be used by any non _test packages and should not import any
// other transit module.
package fixtures

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// root comes from this file's own path rather than the working directory
// for a particular test.
var root = func() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("fixtures: cannot resolve the module root")
	}

	return filepath.Join(filepath.Dir(file), "..", "..")
}()

// Path returns the absolute path to a file or directory under testdata. The
// parts are joined so Path("sample-feed", "agency.txt") reaches into a
// fixture directory.
func Path(parts ...string) string {
	return filepath.Join(append([]string{root, "testdata"}, parts...)...)
}

// Read returns the contents of a file under testdata. The test will fail if
// the file cannot be read.
func Read(t *testing.T, parts ...string) []byte {
	t.Helper()

	path := Path(parts...)

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %s", path, err)
	}

	return content
}
