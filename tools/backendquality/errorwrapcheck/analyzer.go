package errorwrapcheck

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	boundaryReturnDiagnostic = "error wrap: wrap returned error with context before crossing a package boundary"
	fmtErrorfDiagnostic      = "error wrap: use %%w when returning fmt.Errorf with an underlying error"
	ignoredErrorDiagnostic   = "error wrap: do not ignore errors outside cleanup or best-effort paths"
	selfWrapDiagnostic       = "error wrap: do not reassign an error to fmt.Errorf(... %%w, err) on the same variable"
)

// Analyzer reports missing error wrapping and ignored errors in backend packages.
var Analyzer = &analysis.Analyzer{
	Name: "errorwrapcheck",
	Doc:  "reports missing error wrapping and ignored errors in package-boundary flows",
	Run:  run,
}

type walkContext struct {
	boundary bool
	cleanup  bool
	wrapped  map[types.Object]struct{}
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

			walkBlock(pass, fn.Body, walkContext{
				boundary: isBoundaryFunction(fn),
				wrapped:  collectWrappedObjects(pass, fn.Body),
			})
		}
	}

	return nil, nil
}

func shouldSkipFile(filename string) bool {
	clean := filepath.ToSlash(filepath.Clean(filename))
	return strings.HasSuffix(clean, "_test.go")
}

func isBoundaryFunction(fn *ast.FuncDecl) bool {
	return fn.Name != nil && fn.Name.IsExported()
}

func walkBlock(pass *analysis.Pass, block *ast.BlockStmt, ctx walkContext) {
	if block == nil {
		return
	}

	for _, stmt := range block.List {
		walkStmt(pass, stmt, ctx)
	}
}

func walkStmt(pass *analysis.Pass, stmt ast.Stmt, ctx walkContext) {
	switch node := stmt.(type) {
	case *ast.BlockStmt:
		walkBlock(pass, node, ctx)
	case *ast.ReturnStmt:
		if ctx.boundary {
			checkReturn(pass, node, ctx.wrapped)
		}
	case *ast.AssignStmt:
		if !ctx.cleanup {
			checkSelfWrapReassignment(pass, node)
			checkIgnoredAssignments(pass, node)
		}
	case *ast.DeclStmt:
		if !ctx.cleanup {
			checkIgnoredDecl(pass, node)
		}
	case *ast.ExprStmt:
		if !ctx.cleanup {
			checkIgnoredExpr(pass, node)
		}
	case *ast.IfStmt:
		walkStmt(pass, node.Init, ctx)
		walkBlock(pass, node.Body, ctx)
		walkStmt(pass, node.Else, ctx)
	case *ast.ForStmt:
		walkStmt(pass, node.Init, ctx)
		walkStmt(pass, node.Post, ctx)
		walkBlock(pass, node.Body, ctx)
	case *ast.RangeStmt:
		walkBlock(pass, node.Body, ctx)
	case *ast.SwitchStmt:
		walkStmt(pass, node.Init, ctx)
		walkBlock(pass, node.Body, ctx)
	case *ast.TypeSwitchStmt:
		walkStmt(pass, node.Init, ctx)
		walkStmt(pass, node.Assign, ctx)
		walkBlock(pass, node.Body, ctx)
	case *ast.SelectStmt:
		walkBlock(pass, node.Body, ctx)
	case *ast.CaseClause:
		for _, bodyStmt := range node.Body {
			walkStmt(pass, bodyStmt, ctx)
		}
	case *ast.CommClause:
		for _, bodyStmt := range node.Body {
			walkStmt(pass, bodyStmt, ctx)
		}
	case *ast.LabeledStmt:
		walkStmt(pass, node.Stmt, ctx)
	case *ast.DeferStmt:
		walkDeferredCall(pass, node.Call)
	}
}

func walkDeferredCall(pass *analysis.Pass, call *ast.CallExpr) {
	if call == nil {
		return
	}

	lit, ok := call.Fun.(*ast.FuncLit)
	if !ok {
		return
	}

	walkBlock(pass, lit.Body, walkContext{
		boundary: false,
		cleanup:  true,
		wrapped:  nil,
	})
}

func checkReturn(pass *analysis.Pass, stmt *ast.ReturnStmt, wrapped map[types.Object]struct{}) {
	for _, result := range stmt.Results {
		expr := unwrapExpr(result)
		if call, ok := expr.(*ast.CallExpr); ok && isFmtErrorfWithoutWrap(pass, call) {
			pass.Reportf(call.Pos(), fmtErrorfDiagnostic)
			continue
		}

		ident, ok := expr.(*ast.Ident)
		if !ok || ident.Name == "nil" {
			continue
		}
		if !isErrorType(pass.TypesInfo.TypeOf(ident)) {
			continue
		}
		if isPackageLevelObject(pass, ident) {
			continue
		}
		if isWrappedLocalObject(pass, ident, wrapped) {
			continue
		}

		pass.Reportf(ident.Pos(), boundaryReturnDiagnostic)
	}
}

func checkIgnoredAssignments(pass *analysis.Pass, stmt *ast.AssignStmt) {
	if len(stmt.Lhs) == 0 || len(stmt.Rhs) == 0 {
		return
	}

	reported := false
	if len(stmt.Rhs) == 1 {
		if call, ok := unwrapExpr(stmt.Rhs[0]).(*ast.CallExpr); ok {
			if sig, ok := signatureOf(pass, call.Fun); ok {
				results := sig.Results()
				for idx, lhs := range stmt.Lhs {
					if !isBlankIdentifier(lhs) || idx >= results.Len() {
						continue
					}
					if isErrorType(results.At(idx).Type()) {
						pass.Reportf(lhs.Pos(), ignoredErrorDiagnostic)
						reported = true
					}
				}
			}
		}
	}
	if reported {
		return
	}
	if len(stmt.Rhs) == 1 {
		return
	}

	limit := min(len(stmt.Lhs), len(stmt.Rhs))
	for idx := range limit {
		if !isBlankIdentifier(stmt.Lhs[idx]) {
			continue
		}
		if exprReturnsError(pass, stmt.Rhs[idx]) {
			pass.Reportf(stmt.Lhs[idx].Pos(), ignoredErrorDiagnostic)
		}
	}
}

func checkIgnoredDecl(pass *analysis.Pass, stmt *ast.DeclStmt) {
	genDecl, ok := stmt.Decl.(*ast.GenDecl)
	if !ok || genDecl.Tok != token.VAR {
		return
	}

	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok || len(valueSpec.Names) == 0 || len(valueSpec.Values) == 0 {
			continue
		}

		reported := false
		if len(valueSpec.Values) == 1 {
			if call, ok := unwrapExpr(valueSpec.Values[0]).(*ast.CallExpr); ok {
				if sig, ok := signatureOf(pass, call.Fun); ok {
					results := sig.Results()
					for idx, name := range valueSpec.Names {
						if name.Name != "_" || idx >= results.Len() {
							continue
						}
						if isErrorType(results.At(idx).Type()) {
							pass.Reportf(name.Pos(), ignoredErrorDiagnostic)
							reported = true
						}
					}
				}
			}
		}
		if reported {
			continue
		}
		if len(valueSpec.Values) == 1 {
			continue
		}

		limit := min(len(valueSpec.Names), len(valueSpec.Values))
		for idx := range limit {
			if valueSpec.Names[idx].Name != "_" {
				continue
			}
			if exprReturnsError(pass, valueSpec.Values[idx]) {
				pass.Reportf(valueSpec.Names[idx].Pos(), ignoredErrorDiagnostic)
			}
		}
	}
}

func checkIgnoredExpr(pass *analysis.Pass, stmt *ast.ExprStmt) {
	call, ok := unwrapExpr(stmt.X).(*ast.CallExpr)
	if !ok {
		return
	}
	if isBestEffortIgnoredCall(pass, call) {
		return
	}
	if !callReturnsError(pass, call) {
		return
	}

	pass.Reportf(call.Pos(), ignoredErrorDiagnostic)
}

func checkSelfWrapReassignment(pass *analysis.Pass, stmt *ast.AssignStmt) {
	if stmt.Tok != token.ASSIGN || len(stmt.Lhs) != 1 || len(stmt.Rhs) != 1 {
		return
	}

	lhs, ok := unwrapExpr(stmt.Lhs[0]).(*ast.Ident)
	if !ok || lhs.Name == "_" {
		return
	}

	call, ok := unwrapExpr(stmt.Rhs[0]).(*ast.CallExpr)
	if !ok || !isFmtErrorfCall(call) {
		return
	}

	if !containsSelfWrapArg(pass, call.Args[1:], lhs) {
		return
	}

	format, ok := stringLiteralValue(pass, call.Args[0])
	if !ok || !strings.Contains(format, "%w") {
		return
	}

	pass.Reportf(stmt.Pos(), selfWrapDiagnostic)
}

func collectWrappedObjects(pass *analysis.Pass, body *ast.BlockStmt) map[types.Object]struct{} {
	wrapped := make(map[types.Object]struct{})
	if body == nil {
		return wrapped
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			if len(node.Lhs) == 1 && len(node.Rhs) == 1 {
				markWrappedObject(pass, wrapped, node.Lhs[0], node.Rhs[0])
			}
		case *ast.ValueSpec:
			limit := min(len(node.Names), len(node.Values))
			for idx := range limit {
				markWrappedObject(pass, wrapped, node.Names[idx], node.Values[idx])
			}
		}
		return true
	})

	return wrapped
}

func markWrappedObject(pass *analysis.Pass, wrapped map[types.Object]struct{}, lhs ast.Expr, rhs ast.Expr) {
	ident, ok := unwrapExpr(lhs).(*ast.Ident)
	if !ok || ident.Name == "_" {
		return
	}

	call, ok := unwrapExpr(rhs).(*ast.CallExpr)
	if !ok || !isFmtErrorfCall(call) || !containsErrorArg(pass, call.Args[1:]) {
		return
	}

	format, ok := stringLiteralValue(pass, call.Args[0])
	if !ok || !strings.Contains(format, "%w") {
		return
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return
	}

	wrapped[obj] = struct{}{}
}

func isFmtErrorfWithoutWrap(pass *analysis.Pass, call *ast.CallExpr) bool {
	if !isFmtErrorfCall(call) || len(call.Args) < 2 {
		return false
	}

	if !containsErrorArg(pass, call.Args[1:]) {
		return false
	}

	format, ok := stringLiteralValue(pass, call.Args[0])
	if !ok {
		return false
	}

	return !strings.Contains(format, "%w")
}

func isFmtErrorfCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Errorf" {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	return ok && pkgIdent.Name == "fmt"
}

func containsErrorArg(pass *analysis.Pass, args []ast.Expr) bool {
	for _, arg := range args {
		if isErrorType(pass.TypesInfo.TypeOf(arg)) {
			return true
		}
	}
	return false
}

func containsSelfWrapArg(pass *analysis.Pass, args []ast.Expr, lhs *ast.Ident) bool {
	lhsObj := pass.TypesInfo.ObjectOf(lhs)
	if lhsObj == nil {
		return false
	}

	for _, arg := range args {
		ident, ok := unwrapExpr(arg).(*ast.Ident)
		if !ok {
			continue
		}
		if pass.TypesInfo.ObjectOf(ident) == lhsObj {
			return true
		}
	}

	return false
}

func stringLiteralValue(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[unwrapExpr(expr)]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}

	return constant.StringVal(tv.Value), true
}

func exprReturnsError(pass *analysis.Pass, expr ast.Expr) bool {
	switch value := unwrapExpr(expr).(type) {
	case *ast.CallExpr:
		return callReturnsError(pass, value)
	default:
		return isErrorType(pass.TypesInfo.TypeOf(value))
	}
}

func callReturnsError(pass *analysis.Pass, call *ast.CallExpr) bool {
	sig, ok := signatureOf(pass, call.Fun)
	if !ok {
		return false
	}

	return tupleHasError(sig.Results())
}

func signatureOf(pass *analysis.Pass, expr ast.Expr) (*types.Signature, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return nil, false
	}

	sig, ok := t.Underlying().(*types.Signature)
	return sig, ok
}

func tupleHasError(results *types.Tuple) bool {
	if results == nil {
		return false
	}

	for idx := range results.Len() {
		if isErrorType(results.At(idx).Type()) {
			return true
		}
	}

	return false
}

func isBlankIdentifier(expr ast.Expr) bool {
	ident, ok := unwrapExpr(expr).(*ast.Ident)
	return ok && ident.Name == "_"
}

func isPackageLevelObject(pass *analysis.Pass, ident *ast.Ident) bool {
	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	return obj.Parent() == pass.Pkg.Scope()
}

func isWrappedLocalObject(pass *analysis.Pass, ident *ast.Ident, wrapped map[types.Object]struct{}) bool {
	if len(wrapped) == 0 {
		return false
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	_, ok := wrapped[obj]
	return ok
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}

	if named, ok := t.(*types.Named); ok {
		if obj := named.Obj(); obj != nil && obj.Pkg() == nil && obj.Name() == "error" {
			return true
		}
	}

	return types.Implements(t, errorInterface()) || types.Implements(types.NewPointer(t), errorInterface())
}

func errorInterface() *types.Interface {
	return types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
}

func isBestEffortIgnoredCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "fmt" {
		switch sel.Sel.Name {
		case "Print", "Printf", "Println", "Fprint", "Fprintf", "Fprintln":
			return true
		}
	}

	recvType := pass.TypesInfo.TypeOf(sel.X)
	if isNamedType(recvType, "strings", "Builder") || isNamedType(recvType, "bytes", "Buffer") {
		return strings.HasPrefix(sel.Sel.Name, "Write")
	}

	return false
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

func unwrapExpr(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}
