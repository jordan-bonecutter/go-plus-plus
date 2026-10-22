// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

// The lowered representation is private, including to reflection. Zero-length
// variant fields preserve payload types and alignment in ordinary export data.
const UnionTagField = "union·tag"
const UnionDataField = "union·data"
const UnionFieldPrefix = "union·variant·"
const UnionAnchor = "union·anchor"
const UnionKeepAlive = "union·keepalive"

func (p *parser) unionTypeOrNil() Expr {
	if p.tok != _Name || p.lit != "union" {
		return p.typeOrNil()
	}
	pos := p.pos()
	p.next()
	if p.tok != _Lbrace {
		return p.pexpr(NewName(pos, "union"), false)
	}
	return p.unionType(pos)
}

func (p *parser) unionType(pos Pos) Expr {
	p.hasUnion = true
	typ := new(StructType)
	typ.pos = pos
	typ.FieldList = append(typ.FieldList, unionField(pos, UnionTagField, NewName(pos, "uint")))
	seen := make(map[string]bool)
	p.want(_Lbrace)
	p.list("union type", _Semi, _Rbrace, func() bool {
		name := p.name()
		if name.Value == "_" {
			p.errorAt(name.Pos(), "union variant must have a name")
		}
		if seen[name.Value] {
			p.errorAt(name.Pos(), "duplicate union variant "+name.Value)
		}
		seen[name.Value] = true
		typ.FieldList = append(typ.FieldList, unionField(name.Pos(), UnionFieldPrefix+name.Value, p.type_()))
		return false
	})
	if len(typ.FieldList) == 1 {
		p.errorAt(pos, "union must have at least one variant")
	}
	return typ
}

func unionField(pos Pos, name string, typ Expr) *Field {
	f := &Field{Name: NewName(pos, name), Type: typ}
	f.pos = pos
	return f
}

func unionZeroArray(pos Pos, typ Expr) *ArrayType {
	zero := &BasicLit{Kind: IntLit, Value: "0"}
	zero.pos = pos
	a := &ArrayType{Len: zero, Elem: typ}
	a.pos = pos
	return a
}

// The raw field's size is determined from the zero-length variant fields by
// the sizing code, after substituting type arguments. This lets generic unions
// use precisely their largest instantiated payload, without dependent arrays
// or a new export-data format.
func lowerUnion(d *TypeDecl) {
	st, ok := d.Type.(*StructType)
	if !ok || len(st.FieldList) == 0 || st.FieldList[0].Name == nil || st.FieldList[0].Name.Value != UnionTagField {
		return
	}
	pos := d.Pos()
	data := new(StructType)
	data.pos = pos
	for _, f := range st.FieldList[1:] {
		data.FieldList = append(data.FieldList, unionField(f.Pos(), f.Name.Value, unionZeroArray(f.Pos(), f.Type)))
	}
	// Keep pointer alignment even for scalar-only payloads.
	data.FieldList = append(data.FieldList, unionField(pos, "_", unionZeroArray(pos, NewName(pos, "union·uintptr"))), unionField(pos, "union·raw", unionZeroArray(pos, NewName(pos, "union·byte"))))
	header := new(StructType)
	header.pos = pos
	header.FieldList = []*Field{unionField(pos, "Index", NewName(pos, "union·uintptr"))}
	ptr := &Operation{Op: Mul, X: header}
	ptr.pos = pos
	st.FieldList = []*Field{unionField(pos, UnionDataField, data), unionField(pos, UnionTagField, ptr)}
}

// variant is contextual only in a switch header. Ordinary assertions against a
// user-defined type named variant continue to work outside a matching switch.
func variantGuard(x Expr) Expr {
	if a, ok := x.(*AssertExpr); ok {
		if name, ok := a.Type.(*Name); ok && name.Value == "variant" {
			g := &TypeSwitchGuard{X: a.X, Variant: true}
			g.SetPos(a.Pos())
			return g
		}
	}
	return x
}
