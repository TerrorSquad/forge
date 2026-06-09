package runner

import (
	"testing"
)

// Fix 3: pathHasPrefix must use forward-slash separator because StagedFiles
// normalises all paths with filepath.ToSlash on every platform.

func TestPathHasPrefix_ForwardSlash(t *testing.T) {
	cases := []struct {
		file, prefix string
		want         bool
	}{
		{"packages/foo/src/Bar.php", "packages/foo", true},
		{"packages/foobar/src/Baz.php", "packages/foo", false}, // prefix match must stop at separator
		{"packages/foo/deep/nested/File.php", "packages/foo", true},
		{"other/file.php", "packages/foo", false},
		{"packages/foo/file.php", "packages/foo/", true}, // trailing slash in prefix is OK
		{"anything", "", true},                            // empty prefix always matches
	}
	for _, c := range cases {
		got := pathHasPrefix(c.file, c.prefix)
		if got != c.want {
			t.Errorf("pathHasPrefix(%q, %q) = %v, want %v", c.file, c.prefix, got, c.want)
		}
	}
}

func TestPathHasPrefix_NoPrefixCollision(t *testing.T) {
	// "packages/foo" must not match files under "packages/foobar"
	if pathHasPrefix("packages/foobar/main.go", "packages/foo") {
		t.Error("packages/foo should not be a prefix of packages/foobar/main.go")
	}
}
