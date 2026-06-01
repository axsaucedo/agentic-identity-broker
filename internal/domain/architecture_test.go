package domain_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDomainTestsDoNotImportEncryptionNoopAdapter(t *testing.T) {
	assertNoImportInDomainTests(t,
		"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/noop",
		func(path string) bool { return strings.HasSuffix(path, "_test.go") },
	)
}

func TestOAuth2ServerStrategiesTestsDoNotImportStorageMemoryAdapter(t *testing.T) {
	assertNoImportInDomainTests(t,
		"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory",
		func(path string) bool { return path == filepath.Join("oauth2server", "strategies_test.go") },
	)
}

func assertNoImportInDomainTests(t *testing.T, forbiddenImport string, includePath func(string) bool) {
	t.Helper()

	var offenders []string
	fset := token.NewFileSet()

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !includePath(path) {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if importPath == forbiddenImport {
				offenders = append(offenders, path)
				break
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk domain test files: %v", err)
	}

	if len(offenders) > 0 {
		t.Fatalf("domain tests must not import %s:\n%s", forbiddenImport, strings.Join(offenders, "\n"))
	}
}
