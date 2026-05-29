package app

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"
)

func TestBuildLocalProviderReusesSharedLocalServices(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "builder.go", nil, 0)
	if err != nil {
		t.Fatalf("parse builder.go: %v", err)
	}

	funcLit := findAssignedFuncLiteral(t, file, "buildLocalProvider")
	forbidden := map[string]struct{}{
		"oauth2server.NewSigningKeyService": {},
		"oauth2server.NewClientAuthService": {},
	}

	found := make(map[string]struct{})
	ast.Inspect(funcLit.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		qualified := pkg.Name + "." + sel.Sel.Name
		if _, isForbidden := forbidden[qualified]; isForbidden {
			found[qualified] = struct{}{}
		}
		return true
	})

	if len(found) > 0 {
		offenders := make([]string, 0, len(found))
		for qualified := range found {
			offenders = append(offenders, qualified)
		}
		sort.Strings(offenders)
		t.Fatalf("buildLocalProvider must reuse shared local services; found constructors inside the provider factory: %s", strings.Join(offenders, ", "))
	}
}

func findAssignedFuncLiteral(t *testing.T, file *ast.File, name string) *ast.FuncLit {
	t.Helper()

	var found *ast.FuncLit
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}

		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhs.Name != name {
			return true
		}

		funcLit, ok := assign.Rhs[0].(*ast.FuncLit)
		if !ok {
			t.Fatalf("%s must be assigned a function literal", name)
		}
		found = funcLit
		return false
	})

	if found == nil {
		t.Fatalf("did not find function literal assigned to %s", name)
	}

	return found
}
