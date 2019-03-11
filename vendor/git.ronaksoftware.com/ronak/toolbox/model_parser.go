package ronak

import (
	"go/ast"
	"strings"
)

/*
   Creation Time: 2019 - Feb - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

// ModelParser
type ModelParser struct {
	triggerComment string
	packageName    string
	objs           []ModelObject
	parse          bool
}

func NewModelParser(triggerComment string) ModelParser {
	return ModelParser{
		triggerComment: triggerComment,
		parse:          triggerComment == "",
	}
}

func (v *ModelParser) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	switch x := n.(type) {
	case *ast.File:
		v.packageName = x.Name.Name
		return v
	case *ast.Comment:
		if strings.Trim(x.Text, "/ ") == v.triggerComment {
			v.parse = true
		}
		return v
	case *ast.TypeSpec:
		if !v.parse {
			return v
		}
		obj := ModelObject{}
		obj.Name = x.Name.Name
		obj.Keys = make(map[string]KeyDetails)
		switch y := x.Type.(type) {
		case *ast.StructType:
			for _, field := range y.Fields.List {
				keyDetails := KeyDetails{}
				keyName := strings.Builder{}

				// Detect Name
				for _, name := range field.Names {
					keyName.WriteString(name.Name)
				}

				// Detect Type
				switch x := field.Type.(type) {
				case *ast.ArrayType:
					keyDetails.Array = true
					switch x := x.Elt.(type) {
					case *ast.Ident:
						keyDetails.Type = x.Name
						keyDetails.Object = x.Obj != nil
					case *ast.StarExpr:
						switch x := x.X.(type) {
						case *ast.Ident:
							keyDetails.Type = x.Name
							keyDetails.Object = x.Obj != nil
						}
					default:

					}
				case *ast.StarExpr:
					keyDetails.Pointer = true
				case *ast.Ident:
					keyDetails.Type = x.Name
					keyDetails.Object = x.Obj != nil
				}

				// Detect Tag
				if field.Tag != nil {
					keyDetails.Tags = make(map[string]string)
					for _, field := range strings.Fields(field.Tag.Value) {
						parts := strings.Split(field, ":")
						if len(parts) == 2 {
							part1 := strings.Trim(parts[0], " \"`")
							part2 := strings.Trim(parts[1], " \"`")
							switch part1 {
							case "rsg":
								for _, v := range strings.Split(part2, ",") {
									switch strings.TrimSpace(v) {
									case "index":
										keyDetails.Index = true
									case "unique":
										keyDetails.Unique = true
									}
								}
							case "rsg_model":
								keyDetails.Model = strings.TrimSpace(part2)
							default:
								keyDetails.Tags[part1] = part2
							}
						}
					}
				}
				obj.Keys[keyName.String()] = keyDetails
			}
			v.objs = append(v.objs, obj)
		}
	default:
		return v
	}

	if len(v.triggerComment) > 0 {
		v.parse = false
	}
	return nil
}

func (v ModelParser) GetModel(name string) ModelObject {
	for _, obj := range v.objs {
		if obj.Name == name {
			return obj
		}
	}
	return ModelObject{}
}

func (v ModelParser) GetModels() []ModelObject {
	return v.objs
}

// ModelObject
type ModelObject struct {
	Name    string
	Package string
	Keys    map[string]KeyDetails
}

// KeyDetails
type KeyDetails struct {
	Index   bool
	Unique  bool
	Array   bool
	Pointer bool
	Object  bool
	Type    string
	Model   string
	Tags    map[string]string
}
