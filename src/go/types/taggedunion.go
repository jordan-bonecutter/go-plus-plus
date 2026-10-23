// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/ast"
	"internal/abi"
	. "internal/types/errors"
	"strings"
)

const taggedUnionPrefix = "union·variant·"

// Keep this representation in sync with cmd/compile/internal/syntax/union.go.
// Zero-length variant fields preserve the payload types in ordinary export data.
func taggedUnion(t Type) *Struct {
	s, _ := t.Underlying().(*Struct)
	if s != nil && len(s.fields) == 3 && s.fields[0].name == "union·data" && s.fields[1].name == "union·anchor" && s.fields[2].name == "union·tag" {
		data, _ := s.fields[0].typ.Underlying().(*Struct)
		return data
	}
	return nil
}

func taggedVariant(t Type, name string) *Var {
	if data := taggedUnion(t); data != nil {
		for _, f := range data.fields[:len(data.fields)-2] {
			if f.name == taggedUnionPrefix+name {
				return f
			}
		}
	}
	return nil
}

func (check *Checker) taggedUnionType(styp *Struct, e *ast.StructType) {
	var fields []*Var
	var names objset
	for _, f := range e.Fields.List {
		if len(f.Names) != 1 || f.Names[0].Name == "_" || f.Tag != nil {
			check.error(f, InvalidSyntaxTree, "union variant must have one name and a payload type")
			continue
		}
		name := f.Names[0]
		var typ Type = Universe.Lookup("unit").Type()
		if f.Type != nil {
			typ = check.varType(f.Type)
		}
		field := NewField(name.Pos(), check.pkg, name.Name, typ, false)
		if !check.declareInSet(&names, name.Pos(), field) {
			continue
		}

		fields = append(fields, NewField(name.Pos(), check.pkg, taggedUnionPrefix+name.Name, NewArray(typ, 0), false))
	}
	if len(fields) == 0 {
		check.error(e, InvalidSyntaxTree, "union must have at least one variant")
	}
	fields = append(fields,
		NewField(e.Pos(), check.pkg, "_", NewArray(Typ[Uintptr], 0), false),
		NewField(e.Pos(), check.pkg, "union·raw", NewArray(Typ[Uint8], 0), false))
	tag := Typ[Uint8]
	switch abi.UnionTagSize(uint64(len(fields) - 2)) {
	case 2:
		tag = Typ[Uint16]
	case 4:
		tag = Typ[Uint32]
	case 8:
		tag = Typ[Uint64]
	}
	header := NewStruct(nil, nil)
	styp.fields = []*Var{
		NewField(e.Pos(), check.pkg, "union·data", NewStruct(fields, nil), false),
		NewField(e.Pos(), check.pkg, "union·anchor", NewPointer(header), false),
		NewField(e.Pos(), check.pkg, "union·tag", tag, false),
	}
	styp.markComplete()
}

type taggedConstructorKey struct {
	union   Type
	variant *Var
}

func (check *Checker) taggedConstructor(typ Type, variant *Var) *Func {
	if check.taggedConstructors == nil {
		check.taggedConstructors = make(map[taggedConstructorKey]*Func)
	}
	key := taggedConstructorKey{Unalias(typ), variant}
	if f := check.taggedConstructors[key]; f != nil {
		return f
	}
	param := NewParam(variant.pos, check.pkg, "value", variant.typ.(*Array).elem)
	result := newVar(ResultVar, variant.pos, check.pkg, "", typ)
	sig := NewSignatureType(nil, nil, nil, NewTuple(param), NewTuple(result), false)
	f := NewFunc(variant.pos, variant.pkg, strings.TrimPrefix(variant.name, taggedUnionPrefix), sig)
	check.taggedConstructors[key] = f
	return f
}

func (check *Checker) taggedUnionConstructors(typ Type, e *ast.StructType) {
	data := taggedUnion(typ)
	for _, f := range data.fields[:len(data.fields)-2] {
		constructor := check.taggedConstructor(typ, f)
		for _, field := range e.Fields.List {
			if len(field.Names) == 1 && field.Names[0].Name == constructor.name {
				check.recordDef(field.Names[0], constructor)
			}
		}
	}
}

func isVariantAssertion(e *ast.TypeAssertExpr) bool {
	if e == nil {
		return false
	}
	name, _ := e.Type.(*ast.Ident)
	return name != nil && name.Name == "variant"
}

func (check *Checker) variantSwitch(inner stmtContext, s *ast.TypeSwitchStmt, guard *ast.TypeAssertExpr, lhs *ast.Ident) {
	var x operand
	check.expr(nil, &x, guard.X)
	if !x.isValid() {
		return
	}
	data := taggedUnion(x.typ())
	if data == nil {
		check.errorf(guard.X, InvalidTypeSwitch, "cannot match variants of %s: not a union", x.typ())
		return
	}
	check.multipleDefaults(s.Body.List)
	seen := make(map[*Var]bool)
	hasDefault := false
	var bindings []*Var
	for _, stmt := range s.Body.List {
		clause := stmt.(*ast.CaseClause)
		typ := x.typ()
		if clause.List == nil {
			hasDefault = true
		} else {
			if len(clause.List) != 1 {
				check.error(clause, InvalidTypeSwitch, "variant case must name exactly one variant")
				continue
			}
			name, ok := clause.List[0].(*ast.Ident)
			if !ok {
				check.error(clause.List[0], InvalidTypeSwitch, "variant case must be a variant name")
				continue
			}
			variant := taggedVariant(x.typ(), name.Name)
			if variant == nil {
				check.errorf(name, InvalidTypeSwitch, "unknown variant %s in %s", name.Name, x.typ())
				continue
			}
			if variant.pkg != check.pkg && !isExported(name.Name) {
				check.errorf(name, UnexportedName, "variant %s is not exported", name.Name)
			}
			if seen[variant] {
				check.errorf(name, DuplicateCase, "duplicate variant %s", name.Name)
			}
			seen[variant] = true
			typ = variant.typ.(*Array).elem
			check.recordTypeAndValue(name, typexpr, typ, nil)
			check.recordUse(name, check.taggedConstructor(x.typ(), variant))
		}
		check.openScope(clause, "case")
		if lhs != nil {
			v := newVar(LocalVar, lhs.Pos(), check.pkg, lhs.Name, typ)
			check.declare(check.scope, nil, v, clause.Colon)
			check.recordImplicit(clause, v)
			bindings = append(bindings, v)
		}
		check.stmtList(inner, clause.Body)
		check.closeScope()
	}
	var missing []string
	for _, f := range data.fields[:len(data.fields)-2] {
		if !seen[f] {
			missing = append(missing, strings.TrimPrefix(f.name, taggedUnionPrefix))
		}
	}
	if !hasDefault && len(missing) != 0 {
		check.errorf(guard, InvalidTypeSwitch, "non-exhaustive variant switch: missing %s", strings.Join(missing, ", "))
	} else {
		if check.exhaustiveVariants == nil {
			check.exhaustiveVariants = make(map[*ast.TypeSwitchStmt]bool)
		}
		check.exhaustiveVariants[s] = true
	}
	if lhs != nil {
		used := false
		for _, v := range bindings {
			used = used || check.usedVars[v]
			check.usedVars[v] = true
		}
		if !used {
			check.softErrorf(lhs, UnusedVar, "%s declared and not used", lhs.Name)
		}
	}
}

func unionData(t *Struct) bool {
	n := len(t.fields)
	return n >= 3 && t.fields[n-1].name == "union·raw" && strings.HasPrefix(t.fields[0].name, "union·variant·")
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
