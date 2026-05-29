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
	const forbiddenImport = "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/noop"

	var offenders []string
	fset := token.NewFileSet()

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
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
