// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
	"fmt"
	. "internal/types/errors"
	"strconv"
	"strings"
)

func unionStruct(t Type) *Struct {
	s, _ := t.Underlying().(*Struct)
	if s != nil && len(s.fields) == 3 && s.fields[0].name == syntax.UnionDataField && s.fields[1].name == syntax.UnionAnchor && s.fields[2].name == syntax.UnionTagField {
		data, _ := s.fields[0].typ.Underlying().(*Struct)
		return data
	}
	return nil
}

func unionVariant(t Type, name string) int {
	if s := unionStruct(t); s != nil {
		for i, f := range s.fields[:len(s.fields)-2] {
			if f.name == syntax.UnionFieldPrefix+name {
				return i
			}
		}
	}
	return -1
}

// unionFunction builds operations using checked type expressions rather than
// user-spellable type names. This also supports local, imported, and derived
// union types without adding methods or changing unified IR.
func unionFunction(pos syntax.Pos, typ Type, variant int, operation string, boolTypes ...Type) *syntax.FuncLit {
	var body string
	switch operation {
	case "construct":
		body = fmt.Sprintf(`func f(V T) U { var R U; R.D = %d; A(&R, &V); *(*T)(P(&R.B)) = V; return R }`, variant)
	case "assert":
		body = fmt.Sprintf(`func f(V U) T { if V.D != %d { panic("union variant assertion failed") }; R := *(*T)(P(&V.B)); K(V); return R }`, variant)
	case "assertok":
		body = fmt.Sprintf(`func f(V U) (T, bool) { if V.D != %d { var R T; return R, false }; R := *(*T)(P(&V.B)); K(V); return R, true }`, variant)
	case "extract":
		body = `func f(V U) T { R := *(*T)(P(&V.B)); K(V); return R }`
	case "index":
		body = `func f(V U) W { return W(V.D) }`
	}
	file, err := syntax.Parse(syntax.NewFileBase(pos.RelFilename()), strings.NewReader("package p;"+body), nil, nil, 0)
	if err != nil {
		panic(err)
	}
	decl := file.DeclList[0].(*syntax.FuncDecl)
	lit := &syntax.FuncLit{Type: decl.Type, Body: decl.Body}
	types := map[string]Type{"U": typ, "P": Typ[UnsafePointer], "W": Typ[Uintptr], "bool": Typ[Bool], "H": typ.Underlying().(*Struct).fields[1].typ}
	if len(boolTypes) != 0 {
		types["bool"] = boolTypes[0]
	}
	if operation != "index" {
		types["T"] = unionStruct(typ).fields[variant].typ.(*Array).elem
	}
	names := map[string]string{"panic": "union·panic", "true": "union·true", "false": "union·false", "V": "union·value", "R": "union·result", "D": syntax.UnionTagField, "C": syntax.UnionAnchor, "B": syntax.UnionDataField, "A": syntax.UnionAnchor, "K": syntax.UnionKeepAlive, "N": "union·nil"}
	syntax.Inspect(lit, func(n syntax.Node) bool {
		if n != nil {
			n.SetPos(pos)
		}
		switch n := n.(type) {
		case *syntax.Name:
			if t := types[n.Value]; t != nil {
				n.Value = "union·type"
				tv := syntax.TypeAndValue{Type: t}
				tv.SetIsType()
				n.SetTypeInfo(tv)
			} else if name, ok := names[n.Value]; ok {
				n.Value = name
			}
		case *syntax.BlockStmt:
			n.Rbrace = pos
		}
		return true
	})
	return lit
}

func variantCall(pos syntax.Pos, x syntax.Expr, typ Type, variant int, operation string, boolTypes ...Type) *syntax.CallExpr {
	c := &syntax.CallExpr{Fun: unionFunction(pos, typ, variant, operation, boolTypes...), ArgList: []syntax.Expr{x}}
	c.SetPos(pos)
	return c
}

// Check only the helper, since the source operand was already checked and
// may contain function bodies whose variant switches have been lowered.
func (check *Checker) checkVariantCall(call *syntax.CallExpr) {
	var fun operand
	check.expr(nil, &fun, call.Fun)
	results := fun.typ().(*Signature).results
	var result Type = results
	if results.Len() == 1 {
		result = results.At(0).typ
	}
	check.recordTypeAndValue(call, value, result, nil)
}

func variantInt(pos syntax.Pos, n int) *syntax.BasicLit {
	lit := &syntax.BasicLit{Kind: syntax.IntLit, Value: strconv.Itoa(n)}
	lit.SetPos(pos)
	return lit
}

// Lower to ordinary switches before writing unified IR. The outer default
// retains the original init and labeled break targets; the inner init evaluates
// the union once. The representation travels through existing export data.
func (check *Checker) variantSwitchStmt(inner stmtContext, s *syntax.SwitchStmt, guard *syntax.TypeSwitchGuard) {
	pos := guard.X.Pos()
	temp := syntax.NewName(pos, "union·value")
	init := &syntax.AssignStmt{Op: syntax.Def, Lhs: temp, Rhs: guard.X}
	init.SetPos(pos)
	match := &syntax.SwitchStmt{Init: init, Body: s.Body, Rbrace: s.Rbrace}
	match.SetPos(pos)
	outer := &syntax.CaseClause{Body: []syntax.Stmt{match}, Colon: pos}
	outer.SetPos(pos)
	s.Tag, s.Body = nil, []*syntax.CaseClause{outer}
	check.openScope(outer, "variant switch")
	defer check.closeScope()
	check.openScope(match, "switch")
	defer check.closeScope()
	check.simpleStmt(init)
	v, _ := check.lookup(temp.Value).(*Var)
	if v != nil {
		check.usedVars[v] = true
	}
	if v == nil || !isValid(v.typ) {
		return
	}
	st := unionStruct(v.typ)
	if st == nil {
		check.errorf(guard.X, InvalidTypeSwitch, "cannot match variants of %s: not a union", v.typ)
		return
	}
	match.Tag = variantCall(pos, syntax.NewName(pos, temp.Value), v.typ, 0, "index")
	var tag operand
	check.expr(nil, &tag, match.Tag)
	check.multipleSwitchDefaults(match.Body)
	seen := make(map[int]bool)
	hasDefault := false
	var bindings []*Var
	lhs := guard.Lhs
	if lhs != nil && lhs.Value == "_" {
		check.softErrorf(lhs, NoNewVar, "no new variable on left side of :=")
		lhs = nil
	}
	for _, clause := range match.Body {
		payload := syntax.Expr(syntax.NewName(clause.Pos(), temp.Value))
		if clause.Cases == nil {
			hasDefault = true
		} else {
			cases := syntax.UnpackListExpr(clause.Cases)
			if len(cases) != 1 {
				check.error(clause.Cases, InvalidTypeSwitch, "variant case must name exactly one variant")
				continue
			}
			name, ok := cases[0].(*syntax.Name)
			if !ok {
				check.error(cases[0], InvalidTypeSwitch, "variant case must be a variant name")
				continue
			}
			i := unionVariant(v.typ, name.Value)
			if i < 0 {
				check.errorf(name, InvalidTypeSwitch, "unknown variant %s in %s", name.Value, v.typ)
				continue
			}
			field := st.fields[i]
			variant := strings.TrimPrefix(field.name, syntax.UnionFieldPrefix)
			if field.pkg != check.pkg && !isExported(variant) {
				check.errorf(name, UnexportedName, "variant %s is not exported", name.Value)
			}
			if seen[i] {
				check.errorf(name, DuplicateCase, "duplicate variant %s", name.Value)
			}
			seen[i] = true
			clause.Cases = variantInt(name.Pos(), i)
			check.caseValues(&tag, []syntax.Expr{clause.Cases}, make(valueMap))
			payload = variantCall(name.Pos(), payload, v.typ, i, "extract")
		}
		check.openScope(clause, "case")
		body := clause.Body
		if lhs != nil {
			name := syntax.NewName(lhs.Pos(), lhs.Value)
			bind := &syntax.AssignStmt{Op: syntax.Def, Lhs: name, Rhs: payload}
			bind.SetPos(clause.Pos())
			check.simpleStmt(bind)
			clause.Body = append([]syntax.Stmt{bind}, body...)
			if v, _ := check.lookup(name.Value).(*Var); v != nil {
				bindings = append(bindings, v)
			}
		}
		check.stmtList(inner|inTypeSwitch, body)
		check.closeScope()
	}
	if !hasDefault {
		var missing []string
		for i, f := range st.fields[:len(st.fields)-2] {
			if !seen[i] {
				missing = append(missing, strings.TrimPrefix(f.name, syntax.UnionFieldPrefix))
			}
		}
		if len(missing) != 0 {
			check.errorf(guard, InvalidTypeSwitch, "non-exhaustive variant switch: missing %s", strings.Join(missing, ", "))
		} else if len(match.Body) != 0 {
			// Let the ordinary return analysis see that all cases are accounted for.
			match.Body[len(match.Body)-1].Cases = nil
		}
	}
	if lhs != nil {
		used := false
		for _, v := range bindings {
			used = used || check.usedVars[v]
			check.usedVars[v] = true
		}
		if !used {
			check.softErrorf(lhs, UnusedVar, "%s declared and not used", lhs.Value)
		}
	}
}

func init() {
	Universe.Insert(NewFunc(nopos, nil, syntax.UnionAnchor, makeSig(nil, &emptyInterface, &emptyInterface)))
	Universe.Insert(NewFunc(nopos, nil, syntax.UnionKeepAlive, makeSig(nil, &emptyInterface)))
}

func unionBuiltin(name string) string {
	switch name {
	case "union·panic", "union·true", "union·false", "union·byte", "union·uintptr", "union·uint8", "union·uint16", "union·uint32", "union·uint64", "union·nil":
		return strings.TrimPrefix(name, "union·")
	}
	return ""
}

func unionData(t *Struct) bool {
	n := len(t.fields)
	return n >= 3 && t.fields[n-1].name == "union·raw" && strings.HasPrefix(t.fields[0].name, "union·variant·")
}

func (check *Checker) recordVariantCommaOkTypes(x syntax.Expr, t1 Type) {
	if assertion, ok := syntax.Unparen(x).(*syntax.AssertExpr); ok && assertion.LoweredOK != nil && t1 != Typ[Bool] {
		// Unlike an ordinary function result, comma-ok's boolean is untyped
		// and can be assigned directly to a defined boolean type.
		previous := assertion.LoweredOK.(*syntax.CallExpr)
		// unionFunction embeds the checked union type in its parameter,
		// including when neither Info.Types nor syntax recording is enabled.
		typ := previous.Fun.(*syntax.FuncLit).Type.ParamList[0].Type.GetTypeInfo().Type
		i := unionVariant(typ, assertion.Type.(*syntax.Name).Value)
		call := variantCall(assertion.Pos(), assertion.X, typ, i, "assertok", t1)
		check.checkVariantCall(call)
		assertion.LoweredOK = call
	}
}

// unionWithoutAnchor reports whether fields describe a pointer-free tagged union.
// Match the physical layout decision in cmd/compile/internal/types/size.go.
func unionWithoutAnchor(fields []*Var) bool {
	if len(fields) != 3 || fields[0].name != "union·data" || fields[1].name != "union·anchor" || fields[2].name != "union·tag" {
		return false
	}
	data, ok := fields[0].typ.Underlying().(*Struct)
	if !ok {
		return false
	}
	for _, f := range data.fields[:len(data.fields)-2] {
		if unionPayloadPointers(f.typ.(*Array).elem, make(map[Type]bool)) {
			return false
		}
	}
	return true
}

func unionPayloadPointers(t Type, active map[Type]bool) bool {
	if t == nil || active[t] {
		return true
	}
	active[t] = true
	defer delete(active, t)
	switch u := t.Underlying().(type) {
	case *Basic:
		return u.kind == String || u.kind == UnsafePointer || u.kind == Invalid
	case *Array:
		return u.len != 0 && unionPayloadPointers(u.elem, active)
	case *Struct:
		if len(u.fields) == 3 && u.fields[0].name == "union·data" && u.fields[1].name == "union·anchor" && u.fields[2].name == "union·tag" {
			data := u.fields[0].typ.Underlying().(*Struct)
			for _, f := range data.fields[:len(data.fields)-2] {
				if unionPayloadPointers(f.typ.(*Array).elem, active) {
					return true
				}
			}
			return false
		}
		for _, f := range u.fields {
			if unionPayloadPointers(f.typ, active) {
				return true
			}
		}
		return false
	default:
		return true // pointers, strings, interfaces, slices, maps, funcs, channels, and type parameters
	}
}
