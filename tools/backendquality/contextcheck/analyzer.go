package contextcheck

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	backgroundDiagnostic = "context propagation: avoid context.Background/TODO inside a function that already receives ctx"
	downstreamDiagnostic = "context propagation: pass ctx or a context derived from ctx to downstream context-aware calls"
)

// Analyzer reports context propagation violations in backend packages.
var Analyzer = &analysis.Analyzer{
	Name: "contextcheck",
	Doc:  "reports context propagation violations in functions that already receive context.Context",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if shouldSkipFile(filename) {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if shouldSkipFunction(fn.Name.Name) {
				continue
			}

			allowed := allowedContextNames(pass, fn.Type.Params)
			if len(allowed) == 0 {
				continue
			}

			inspectFunction(pass, fn.Body, allowed)
		}
	}

	return nil, nil
}

func shouldSkipFile(filename string) bool {
	clean := filepath.ToSlash(filepath.Clean(filename))
	return strings.HasSuffix(clean, "_test.go")
}

func shouldSkipFunction(name string) bool {
	switch name {
	case "SetContext", "context", "contextOrBackground":
		return true
	default:
		return false
	}
}

func allowedContextNames(pass *analysis.Pass, fields *ast.FieldList) map[string]struct{} {
	allowed := make(map[string]struct{})
	if fields == nil {
		return allowed
	}

	for _, field := range fields.List {
		if !isContextType(pass.TypesInfo.TypeOf(field.Type)) {
			continue
		}
		for _, name := range field.Names {
			allowed[name.Name] = struct{}{}
		}
	}

	return allowed
}

func inspectFunction(pass *analysis.Pass, body *ast.BlockStmt, initialAllowed map[string]struct{}) {
	allowed := cloneNames(initialAllowed)
	disallowed := make(map[string]struct{})

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.DeferStmt:
			if call, ok := node.Call.Fun.(*ast.FuncLit); ok && call.Body != nil {
				return false
			}
		case *ast.AssignStmt:
			recordContextAssignments(pass, node.Lhs, node.Rhs, allowed, disallowed)
		case *ast.ValueSpec:
			recordContextAssignments(pass, exprsFromIdents(node.Names), node.Values, allowed, disallowed)
		case *ast.CallExpr:
			reportRootContextCall(pass, node)
			reportDownstreamViolation(pass, node, disallowed)
		}
		return true
	})
}

func exprsFromIdents(names []*ast.Ident) []ast.Expr {
	result := make([]ast.Expr, 0, len(names))
	for _, name := range names {
		result = append(result, name)
	}
	return result
}

func cloneNames(src map[string]struct{}) map[string]struct{} {
	dst := make(map[string]struct{}, len(src))
	for key := range src {
		dst[key] = struct{}{}
	}
	return dst
}

func recordContextAssignments(pass *analysis.Pass, lhs []ast.Expr, rhs []ast.Expr, allowed map[string]struct{}, disallowed map[string]struct{}) {
	if len(lhs) == 0 || len(rhs) == 0 {
		return
	}

	if len(rhs) == 1 {
		if call, ok := rhs[0].(*ast.CallExpr); ok {
			if sig, ok := signatureOf(pass, call.Fun); ok {
				results := sig.Results()
				for idx := 0; idx < results.Len() && idx < len(lhs); idx++ {
					if !isContextType(results.At(idx).Type()) {
						continue
					}
					ident, ok := lhs[idx].(*ast.Ident)
					if !ok {
						continue
					}
					if isAllowedContextExpr(pass, call, allowed) {
						allowed[ident.Name] = struct{}{}
						delete(disallowed, ident.Name)
						continue
					}
					if isDisallowedContextExpr(pass, call, disallowed) {
						disallowed[ident.Name] = struct{}{}
						delete(allowed, ident.Name)
					}
				}
			}
		}
	}

	for idx := 0; idx < len(lhs) && idx < len(rhs); idx++ {
		ident, ok := lhs[idx].(*ast.Ident)
		if !ok {
			continue
		}
		if isAllowedContextExpr(pass, rhs[idx], allowed) {
			allowed[ident.Name] = struct{}{}
			delete(disallowed, ident.Name)
			continue
		}
		if isDisallowedContextExpr(pass, rhs[idx], disallowed) {
			disallowed[ident.Name] = struct{}{}
			delete(allowed, ident.Name)
		}
	}
}

func reportRootContextCall(pass *analysis.Pass, call *ast.CallExpr) {
	if isBackgroundOrTODO(call) {
		pass.Reportf(call.Pos(), backgroundDiagnostic)
	}
}

func reportDownstreamViolation(pass *analysis.Pass, call *ast.CallExpr, disallowed map[string]struct{}) {
	sig, ok := signatureOf(pass, call.Fun)
	if !ok || sig.Params().Len() == 0 || len(call.Args) == 0 {
		return
	}
	if !isContextType(sig.Params().At(0).Type()) {
		return
	}
	if nestedCall, ok := call.Args[0].(*ast.CallExpr); ok && isBackgroundOrTODO(nestedCall) {
		return
	}
	ident, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return
	}
	if _, ok := disallowed[ident.Name]; !ok {
		return
	}

	pass.Reportf(call.Args[0].Pos(), downstreamDiagnostic)
}

func signatureOf(pass *analysis.Pass, expr ast.Expr) (*types.Signature, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return nil, false
	}

	sig, ok := t.Underlying().(*types.Signature)
	return sig, ok
}

func isAllowedContextExpr(pass *analysis.Pass, expr ast.Expr, allowed map[string]struct{}) bool {
	switch value := expr.(type) {
	case *ast.Ident:
		_, ok := allowed[value.Name]
		return ok
	case *ast.CallExpr:
		if isBackgroundOrTODO(value) {
			return false
		}
		return returnsDerivedContext(pass, value, allowed)
	case *ast.ParenExpr:
		return isAllowedContextExpr(pass, value.X, allowed)
	default:
		return false
	}
}

func returnsDerivedContext(pass *analysis.Pass, call *ast.CallExpr, allowed map[string]struct{}) bool {
	sig, ok := signatureOf(pass, call.Fun)
	if !ok || sig.Params().Len() == 0 || len(call.Args) == 0 {
		return false
	}
	if !isContextType(sig.Params().At(0).Type()) || !isContextResult(sig.Results()) {
		return false
	}
	return isAllowedContextExpr(pass, call.Args[0], allowed)
}

func isDisallowedContextExpr(pass *analysis.Pass, expr ast.Expr, disallowed map[string]struct{}) bool {
	switch value := expr.(type) {
	case *ast.Ident:
		_, ok := disallowed[value.Name]
		return ok
	case *ast.CallExpr:
		if isBackgroundOrTODO(value) {
			return true
		}
		if sig, ok := signatureOf(pass, value.Fun); ok && sig.Params().Len() > 0 && len(value.Args) > 0 && isContextType(sig.Params().At(0).Type()) && isContextResult(sig.Results()) {
			return isDisallowedContextExpr(pass, value.Args[0], disallowed)
		}
		return false
	case *ast.ParenExpr:
		return isDisallowedContextExpr(pass, value.X, disallowed)
	default:
		return false
	}
}

func isContextResult(results *types.Tuple) bool {
	if results == nil || results.Len() == 0 {
		return false
	}
	return isContextType(results.At(0).Type())
}

func isContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	named, ok := t.(*types.Named)
	if ok {
		if obj := named.Obj(); obj != nil && obj.Pkg() != nil {
			return obj.Pkg().Path() == "context" && obj.Name() == "Context"
		}
	}
	return types.TypeString(t, nil) == "context.Context"
}

func isBackgroundOrTODO(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "context" {
		return false
	}
	return sel.Sel.Name == "Background" || sel.Sel.Name == "TODO"
}
