package slogcheck

import (
	"go/ast"
	"go/constant"
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	contextDiagnostic = "slog usage: use slog.*Context when calling log/slog directly"
	keyDiagnostic     = "slog usage: structured log keys must be lower_snake_case"
)

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

var directMethods = map[string]struct{}{
	"Debug": {},
	"Error": {},
	"Info":  {},
	"Warn":  {},
}

var contextMethods = map[string]struct{}{
	"DebugContext": {},
	"ErrorContext": {},
	"InfoContext":  {},
	"WarnContext":  {},
}

// Analyzer reports direct slog calls without context and invalid structured key names.
var Analyzer = &analysis.Analyzer{
	Name: "slogcheck",
	Doc:  "reports direct log/slog calls without context and invalid structured log key names",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if shouldSkipFile(filename) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			checkDirectCall(pass, call)
			checkStructuredKeys(pass, call)
			return true
		})
	}

	return nil, nil
}

func shouldSkipFile(filename string) bool {
	clean := filepath.ToSlash(filepath.Clean(filename))
	return strings.HasSuffix(clean, "_test.go")
}

func checkDirectCall(pass *analysis.Pass, call *ast.CallExpr) {
	name, ok := slogSelectorName(pass, call.Fun)
	if !ok {
		return
	}
	if _, ok := directMethods[name]; !ok {
		return
	}

	pass.Reportf(call.Pos(), contextDiagnostic)
}

func checkStructuredKeys(pass *analysis.Pass, call *ast.CallExpr) {
	name, ok := slogSelectorName(pass, call.Fun)
	if !ok {
		return
	}

	switch {
	case isLoggerMethod(name):
		checkLoggerCallKeys(pass, call, name)
	case isAttrBuilder(pass, call):
		checkAttrBuilderKey(pass, call)
	}
}

func checkLoggerCallKeys(pass *analysis.Pass, call *ast.CallExpr, name string) {
	keyStart := 1
	if _, ok := contextMethods[name]; ok {
		keyStart = 2
	}
	if len(call.Args) <= keyStart {
		return
	}

	for idx := keyStart; idx < len(call.Args); idx += 2 {
		checkKeyExpr(pass, call.Args[idx])
	}
}

func checkAttrBuilderKey(pass *analysis.Pass, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}
	checkKeyExpr(pass, call.Args[0])
}

func checkKeyExpr(pass *analysis.Pass, expr ast.Expr) {
	key, ok := stringLiteralValue(pass, expr)
	if !ok {
		return
	}
	if keyPattern.MatchString(key) {
		return
	}

	pass.Reportf(expr.Pos(), keyDiagnostic)
}

func slogSelectorName(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	switch value := expr.(type) {
	case *ast.SelectorExpr:
		if isSlogPackageSelector(pass, value) || isSlogLoggerMethod(pass, value) {
			return value.Sel.Name, true
		}
	case *ast.ParenExpr:
		return slogSelectorName(pass, value.X)
	}

	return "", false
}

func isSlogPackageSelector(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "slog" {
		if pkgName, ok := pass.TypesInfo.ObjectOf(ident).(*types.PkgName); ok {
			return pkgName.Imported().Path() == "log/slog"
		}
	}

	obj := pass.TypesInfo.ObjectOf(sel.Sel)
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Pkg().Path() == "log/slog"
}

func isSlogLoggerMethod(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	return isNamedType(pass.TypesInfo.TypeOf(sel.X), "log/slog", "Logger")
}

func isNamedType(t types.Type, pkgPath string, name string) bool {
	if t == nil {
		return false
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Pkg().Path() == pkgPath && obj.Name() == name
}

func isLoggerMethod(name string) bool {
	_, ok := directMethods[name]
	if ok {
		return true
	}
	_, ok = contextMethods[name]
	return ok
}

func isAttrBuilder(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := unwrapExpr(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if !isSlogPackageSelector(pass, sel) {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}

	return isNamedType(pass.TypesInfo.TypeOf(call), "log/slog", "Attr")
}

func stringLiteralValue(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[unwrapExpr(expr)]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}

	return constant.StringVal(tv.Value), true
}

func unwrapExpr(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}
