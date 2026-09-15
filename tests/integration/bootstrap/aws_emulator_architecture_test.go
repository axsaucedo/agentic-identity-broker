package bootstrap

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootstrapExportsCapabilityOrientedAPI(t *testing.T) {
	exportedNames := bootstrapExportedNames(t)

	assert.Contains(t, exportedNames, "AWSEmulatorContainer")
	assert.Contains(t, exportedNames, "StartAWSEmulator")
	assert.Contains(t, exportedNames, "StartAWSEmulatorForSuite")
	assert.Contains(t, exportedNames, "NewKMSClientForEmulator")

	assert.NotContains(t, exportedNames, "FlociContainer")
	assert.NotContains(t, exportedNames, "StartFloci")
	assert.NotContains(t, exportedNames, "StartFlociForSuite")
	assert.NotContains(t, exportedNames, "NewKMSClientForFloci")
}

func TestBootstrapCentralizesLocalStackCompatibleHealthEndpoint(t *testing.T) {
	constName, ok := bootstrapStringConstValue(t, "/_localstack/health")
	require.True(t, ok, "bootstrap must centralize the LocalStack-compatible health endpoint in a named constant")
	assert.Equal(t, "localStackCompatibleHealthEndpoint", constName)
}

func TestBootstrapDoesNotInjectRyukFlagIntoEmulatorContainerEnv(t *testing.T) {
	assert.False(t, bootstrapFunctionInjectsContainerEnvKey(t, "startEmulator", "TESTCONTAINERS_RYUK_DISABLED"))
}

func TestBootstrapDisablesRyukBeforeStartingTestcontainers(t *testing.T) {
	setenvPos, ok := bootstrapCallPosition(t, "startEmulator", "os", "Setenv", "TESTCONTAINERS_RYUK_DISABLED")
	require.True(t, ok, "startEmulator must set TESTCONTAINERS_RYUK_DISABLED before starting testcontainers")

	containerStartPos, ok := bootstrapCallPosition(t, "startEmulator", "testcontainers", "GenericContainer", "")
	require.True(t, ok, "startEmulator must call testcontainers.GenericContainer")

	assert.Less(t, setenvPos, containerStartPos)
}

func TestBootstrapEnvironmentLifecycleMethodsReturnErrors(t *testing.T) {
	assert.True(t, bootstrapMethodReturnsError(t, "SetupEnvironment"))
	assert.True(t, bootstrapMethodReturnsError(t, "CleanupEnvironment"))
}

func TestBootstrapDoesNotDiscardTerminateErrorsDuringStartupCleanup(t *testing.T) {
	assert.False(t, bootstrapFunctionDiscardsTerminateResult(t, "startEmulator"))
}

func bootstrapExportedNames(t *testing.T) []string {
	t.Helper()

	pkg := parseBootstrapPackage(t)

	var names []string
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Recv == nil && decl.Name.IsExported() {
					names = append(names, decl.Name.Name)
				}
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					continue
				}
				for _, spec := range decl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if ok && typeSpec.Name.IsExported() {
						names = append(names, typeSpec.Name.Name)
					}
				}
			}
		}
	}

	sort.Strings(names)
	return names
}

func bootstrapStringConstValue(t *testing.T, wantValue string) (string, bool) {
	t.Helper()

	pkg := parseBootstrapPackage(t)

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok || len(valueSpec.Names) != 1 || len(valueSpec.Values) != 1 {
					continue
				}
				literal, ok := valueSpec.Values[0].(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					continue
				}
				if literal.Value == `"`+wantValue+`"` {
					return valueSpec.Names[0].Name, true
				}
			}
		}
	}

	return "", false
}

func bootstrapFunctionInjectsContainerEnvKey(t *testing.T, functionName, key string) bool {
	t.Helper()

	fn := bootstrapFunctionDecl(t, functionName)
	var found bool
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}

		selector, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "ContainerRequest" {
			return true
		}

		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			ident, ok := kv.Key.(*ast.Ident)
			if !ok || ident.Name != "Env" {
				continue
			}

			envLit, ok := kv.Value.(*ast.CompositeLit)
			if !ok {
				continue
			}
			for _, envElt := range envLit.Elts {
				envKV, ok := envElt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if stringLiteralValue(envKV.Key) == key {
					found = true
					return false
				}
			}
		}

		return true
	})

	return found
}

func bootstrapCallPosition(t *testing.T, functionName, packageName, callName, firstArg string) (token.Pos, bool) {
	t.Helper()

	fn := bootstrapFunctionDecl(t, functionName)
	var (
		callPos token.Pos
		found   bool
	)

	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkgIdent, ok := selector.X.(*ast.Ident)
		if !ok || pkgIdent.Name != packageName || selector.Sel.Name != callName {
			return true
		}
		if firstArg != "" {
			if len(call.Args) == 0 || stringLiteralValue(call.Args[0]) != firstArg {
				return true
			}
		}
		callPos = call.Pos()
		found = true
		return false
	})

	return callPos, found
}

func bootstrapMethodReturnsError(t *testing.T, methodName string) bool {
	t.Helper()

	fn := bootstrapFunctionDecl(t, methodName)
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}

	result, ok := fn.Type.Results.List[0].Type.(*ast.Ident)
	return ok && result.Name == "error"
}

func bootstrapFunctionDiscardsTerminateResult(t *testing.T, functionName string) bool {
	t.Helper()

	fn := bootstrapFunctionDecl(t, functionName)
	var discarded bool
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			if len(node.Lhs) != 1 || len(node.Rhs) != 1 {
				return true
			}
			ident, ok := node.Lhs[0].(*ast.Ident)
			if !ok || ident.Name != "_" || !isTerminateCall(node.Rhs[0]) {
				return true
			}
			discarded = true
			return false
		case *ast.ExprStmt:
			if !isTerminateCall(node.X) {
				return true
			}
			discarded = true
			return false
		default:
			return true
		}
	})

	return discarded
}

func isTerminateCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return selector.Sel.Name == "Terminate"
}

func bootstrapFunctionDecl(t *testing.T, functionName string) *ast.FuncDecl {
	t.Helper()

	pkg := parseBootstrapPackage(t)

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Name.Name == functionName {
				return fn
			}
		}
	}

	t.Fatalf("function %s not found", functionName)
	return nil
}

func stringLiteralValue(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	return strings.Trim(lit.Value, `"`)
}

type bootstrapPackage struct {
	Files []*ast.File
}

func parseBootstrapPackage(t *testing.T) *bootstrapPackage {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "resolve caller path")

	dir := filepath.Dir(thisFile)
	buildPackage, err := build.Default.ImportDir(dir, 0)
	require.NoError(t, err)
	require.NotEmpty(t, buildPackage.GoFiles, "bootstrap package must include source files")

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(buildPackage.GoFiles))
	for _, name := range buildPackage.GoFiles {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		require.NoErrorf(t, err, "parse %s", name)
		files = append(files, file)
	}

	return &bootstrapPackage{Files: files}
}
