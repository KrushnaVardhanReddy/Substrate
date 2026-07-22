//go:build !wasip1

package sql

import (
	"fmt"
	"os"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
	wasipg "github.com/wasilibs/go-pgquery"
)

func ParseSchema(path string) (*SQLSchema, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sql parse error: failed to read file: %w", err)
	}
	return ParseSchemaFromBytes(bytes)
}

func ParseSchemaFromBytes(src []byte) (*SQLSchema, error) {
	schema := &SQLSchema{
		Tables:    make(map[string]*Table),
		Views:     make(map[string]*View),
		Enums:     make(map[string]*EnumType),
		Sequences: make(map[string]*Sequence),
	}

	result, err := wasipg.Parse(string(src))
	if err != nil {
		return nil, fmt.Errorf("sql parse error: %v", err)
	}

	for _, stmtNode := range result.Stmts {
		stmt := stmtNode.Stmt
		if stmt == nil {
			continue
		}

		if stmt.GetCreateStmt() != nil {
			createStmt := stmt.GetCreateStmt()
			table := parseCreateStmt(createStmt)
			if table != nil {
				schema.Tables[table.Name] = table
			}
		} else if stmt.GetAlterTableStmt() != nil {
			alterTableStmt := stmt.GetAlterTableStmt()
			handleAlterTableStmt(schema, alterTableStmt)
		} else if stmt.GetIndexStmt() != nil {
			indexStmt := stmt.GetIndexStmt()
			handleIndexStmt(schema, indexStmt)
		} else if stmt.GetViewStmt() != nil {
			viewStmt := stmt.GetViewStmt()
			view := parseViewStmt(viewStmt)
			if view != nil {
				schema.Views[view.Name] = view
			}
		} else if stmt.GetCreateTableAsStmt() != nil {
			createTableAsStmt := stmt.GetCreateTableAsStmt()
			view := parseCreateTableAsStmt(createTableAsStmt)
			if view != nil {
				schema.Views[view.Name] = view
			}
		} else if stmt.GetCompositeTypeStmt() != nil {
			_ = stmt.GetCompositeTypeStmt()
			// only ENUM is handled? composite types and enums are different in postgres AST.
			// pg_dump usually dumps ENUM as CreateEnumStmt
			// Need to verify
		} else if stmt.GetCreateEnumStmt() != nil {
			createEnumStmt := stmt.GetCreateEnumStmt()
			enum := parseCreateEnumStmt(createEnumStmt)
			if enum != nil {
				schema.Enums[enum.Name] = enum
			}
		} else if stmt.GetCreateSeqStmt() != nil {
			createSeqStmt := stmt.GetCreateSeqStmt()
			seq := parseCreateSeqStmt(createSeqStmt)
			if seq != nil {
				schema.Sequences[seq.Name] = seq
			}
		}
	}

	return schema, nil
}

func parseCreateStmt(stmt *pg_query.CreateStmt) *Table {
	rel := stmt.GetRelation()
	if rel == nil || rel.GetRelname() == "" {
		return nil
	}

	table := &Table{
		Name:        rel.GetRelname(),
		Columns:     make(map[string]*Column),
		Constraints: make(map[string]*Constraint),
		Indexes:     make(map[string]*Index),
	}

	for _, eltNode := range stmt.GetTableElts() {
		if eltNode.GetColumnDef() != nil {
			colDef := eltNode.GetColumnDef()
			col := parseColumnDef(colDef)
			if col != nil {
				table.Columns[col.Name] = col

				// Extract constraints from column def to table level
				for _, conNode := range colDef.GetConstraints() {
					con := conNode.GetConstraint()
					if con != nil {
						constraint := parseConstraint(con)
						if constraint != nil {
							// If it's a PK or UNIQUE, we need to add the column name
							if constraint.Type == "PRIMARY_KEY" || constraint.Type == "UNIQUE" {
								constraint.Columns = append(constraint.Columns, col.Name)
							}

							if constraint.Name == "" {
								constraint.Name = fmt.Sprintf("%s_%s_%s", table.Name, col.Name, strings.ToLower(constraint.Type))
							}
							table.Constraints[constraint.Name] = constraint
						}
					}
				}
			}
		} else if eltNode.GetConstraint() != nil {
			con := eltNode.GetConstraint()
			constraint := parseConstraint(con)
			if constraint != nil {
				if constraint.Name == "" {
					constraint.Name = fmt.Sprintf("%s_%s", table.Name, strings.ToLower(constraint.Type))
				}
				table.Constraints[constraint.Name] = constraint
			}
		}
	}

	return table
}

func parseColumnDef(colDef *pg_query.ColumnDef) *Column {
	if colDef.GetColname() == "" {
		return nil
	}

	typeName, length, precision, scale := parseTypeName(colDef.GetTypeName())

	col := &Column{
		Name:      colDef.GetColname(),
		DataType:  typeName,
		Length:    length,
		Precision: precision,
		Scale:     scale,
	}

	for _, conNode := range colDef.GetConstraints() {
		con := conNode.GetConstraint()
		if con == nil {
			continue
		}

		switch con.GetContype() {
		case pg_query.ConstrType_CONSTR_NOTNULL:
			col.NotNull = true
		case pg_query.ConstrType_CONSTR_PRIMARY:
			col.IsPrimaryKey = true
			col.NotNull = true
		case pg_query.ConstrType_CONSTR_DEFAULT:
			defExpr := formatExpr(con.GetRawExpr())
			if defExpr != "" {
				col.Default = &defExpr
			}
		}
	}

	return col
}

func parseTypeName(typeNameNode *pg_query.TypeName) (string, *int, *int, *int) {
	if typeNameNode == nil {
		return "", nil, nil, nil
	}

	var names []string
	for _, n := range typeNameNode.GetNames() {
		if str := n.GetString_(); str != nil {
			names = append(names, str.GetSval())
		}
	}

	rawName := strings.Join(names, ".")
	if len(names) > 0 {
		rawName = names[len(names)-1]
	}

	// Normalize data type
	normalized := strings.ToLower(rawName)
	switch normalized {
	case "int", "int4", "integer":
		normalized = "integer"
	case "int2", "smallint":
		normalized = "smallint"
	case "int8", "bigint":
		normalized = "bigint"
	case "bool", "boolean":
		normalized = "boolean"
	case "float4", "real":
		normalized = "real"
	case "float8", "double precision":
		normalized = "double precision"
	case "character varying", "varchar":
		normalized = "varchar"
	case "character", "char":
		normalized = "char"
	case "timestamp with time zone", "timestamptz":
		normalized = "timestamptz"
	case "timestamp", "timestamp without time zone":
		normalized = "timestamp"
	}

	var length, precision, scale *int

	typmods := typeNameNode.GetTypmods()
	if len(typmods) > 0 {
		if len(typmods) == 1 {
			if a := typmods[0].GetAConst(); a != nil {
				if val := a.GetIval(); val != nil {
					l := int(val.GetIval())
					// some types apply length, some apply precision
					if normalized == "numeric" || normalized == "decimal" {
						precision = &l
					} else {
						length = &l
					}
				}
			}
		} else if len(typmods) == 2 {
			if a1 := typmods[0].GetAConst(); a1 != nil {
				if val1 := a1.GetIval(); val1 != nil {
					p := int(val1.GetIval())
					precision = &p
				}
			}
			if a2 := typmods[1].GetAConst(); a2 != nil {
				if val2 := a2.GetIval(); val2 != nil {
					s := int(val2.GetIval())
					scale = &s
				}
			}
		}
	}

	return normalized, length, precision, scale
}

func parseConstraint(con *pg_query.Constraint) *Constraint {
	if con == nil {
		return nil
	}

	constraint := &Constraint{
		Name: con.GetConname(),
	}

	switch con.GetContype() {
	case pg_query.ConstrType_CONSTR_PRIMARY:
		constraint.Type = "PRIMARY_KEY"
		for _, key := range con.GetKeys() {
			if str := key.GetString_(); str != nil {
				constraint.Columns = append(constraint.Columns, str.GetSval())
			}
		}
	case pg_query.ConstrType_CONSTR_UNIQUE:
		constraint.Type = "UNIQUE"
		for _, key := range con.GetKeys() {
			if str := key.GetString_(); str != nil {
				constraint.Columns = append(constraint.Columns, str.GetSval())
			}
		}
	case pg_query.ConstrType_CONSTR_CHECK:
		constraint.Type = "CHECK"
		exprStr, _ := deparseExpr(con.GetRawExpr())
		constraint.CheckExpr = &exprStr
	case pg_query.ConstrType_CONSTR_FOREIGN:
		constraint.Type = "FOREIGN_KEY"
		if pktable := con.GetPktable(); pktable != nil {
			refTable := pktable.GetRelname()
			constraint.RefTable = &refTable
		}
		for _, fk := range con.GetFkAttrs() {
			if str := fk.GetString_(); str != nil {
				constraint.Columns = append(constraint.Columns, str.GetSval())
			}
		}
		for _, pk := range con.GetPkAttrs() {
			if str := pk.GetString_(); str != nil {
				constraint.RefColumns = append(constraint.RefColumns, str.GetSval())
			}
		}
	default:
		return nil
	}

	return constraint
}

func formatExpr(node *pg_query.Node) string {
	if node == nil {
		return ""
	}

	stmt := &pg_query.SelectStmt{
		TargetList: []*pg_query.Node{
			{
				Node: &pg_query.Node_ResTarget{
					ResTarget: &pg_query.ResTarget{
						Val: node,
					},
				},
			},
		},
	}

	tree := &pg_query.ParseResult{
		Stmts: []*pg_query.RawStmt{
			{
				Stmt: &pg_query.Node{
					Node: &pg_query.Node_SelectStmt{
						SelectStmt: stmt,
					},
				},
			},
		},
	}

	out, err := wasipg.Deparse(tree)
	if err == nil && out != "" {
		if len(out) > 7 && out[:7] == "SELECT " {
			return out[7:]
		}
		return out
	}
	return ""
}

func deparseExpr(node *pg_query.Node) (string, error) {
	return formatExpr(node), nil
}

func formatStmt(node *pg_query.Node) string {
	if node == nil {
		return ""
	}
	tree := &pg_query.ParseResult{
		Stmts: []*pg_query.RawStmt{
			{
				Stmt: node,
			},
		},
	}
	out, err := wasipg.Deparse(tree)
	if err == nil {
		return out
	}
	return ""
}
func deparseStmt(node *pg_query.Node) (string, error) {
	return formatStmt(node), nil
}

// Dummy to replace previous declarations

func handleAlterTableStmt(schema *SQLSchema, stmt *pg_query.AlterTableStmt) {
	rel := stmt.GetRelation()
	if rel == nil || rel.GetRelname() == "" {
		return
	}
	tableName := rel.GetRelname()
	table, ok := schema.Tables[tableName]
	if !ok {
		return
	}

	for _, cmdNode := range stmt.GetCmds() {
		cmd := cmdNode.GetAlterTableCmd()
		if cmd == nil {
			continue
		}

		switch cmd.GetSubtype() {
		case pg_query.AlterTableType_AT_AddColumn:
			if colDef := cmd.GetDef().GetColumnDef(); colDef != nil {
				col := parseColumnDef(colDef)
				if col != nil {
					table.Columns[col.Name] = col
				}
			}
		case pg_query.AlterTableType_AT_AddConstraint:
			if con := cmd.GetDef().GetConstraint(); con != nil {
				constraint := parseConstraint(con)
				if constraint != nil {
					if constraint.Name == "" {
						constraint.Name = fmt.Sprintf("%s_%s", table.Name, strings.ToLower(constraint.Type))
					}
					table.Constraints[constraint.Name] = constraint
				}
			}
		case pg_query.AlterTableType_AT_DropColumn:
			delete(table.Columns, cmd.GetName())
		case pg_query.AlterTableType_AT_DropConstraint:
			delete(table.Constraints, cmd.GetName())
		case pg_query.AlterTableType_AT_AlterColumnType:
			if colDef := cmd.GetDef().GetColumnDef(); colDef != nil {
				colName := cmd.GetName()
				if col, exists := table.Columns[colName]; exists {
					typeName, length, precision, scale := parseTypeName(colDef.GetTypeName())
					col.DataType = typeName
					col.Length = length
					col.Precision = precision
					col.Scale = scale
				}
			}
		case pg_query.AlterTableType_AT_SetNotNull:
			if col, exists := table.Columns[cmd.GetName()]; exists {
				col.NotNull = true
			}
		case pg_query.AlterTableType_AT_DropNotNull:
			if col, exists := table.Columns[cmd.GetName()]; exists {
				col.NotNull = false
			}
		case pg_query.AlterTableType_AT_ColumnDefault:
			colName := cmd.GetName()
			if col, exists := table.Columns[colName]; exists {
				defExpr := formatExpr(cmd.GetDef())
				if defExpr != "" {
					col.Default = &defExpr
				} else {
					col.Default = nil // Drop default
				}
			}
		}
	}
}

func handleIndexStmt(schema *SQLSchema, stmt *pg_query.IndexStmt) {
	rel := stmt.GetRelation()
	if rel == nil || rel.GetRelname() == "" {
		return
	}
	tableName := rel.GetRelname()
	table, ok := schema.Tables[tableName]
	if !ok {
		return
	}

	idxName := stmt.GetIdxname()
	if idxName == "" {
		// PostgreSQL generates a name, but we might not have it in the AST. Skip if empty.
		return
	}

	idx := &Index{
		Name:   idxName,
		Unique: stmt.GetUnique(),
	}

	for _, paramNode := range stmt.GetIndexParams() {
		if param := paramNode.GetIndexElem(); param != nil {
			if name := param.GetName(); name != "" {
				idx.Columns = append(idx.Columns, name)
			}
		}
	}

	if stmt.GetWhereClause() != nil {
		whereStr, _ := deparseExpr(stmt.GetWhereClause())
		if whereStr != "" {
			idx.Where = &whereStr
		}
	}

	table.Indexes[idx.Name] = idx
}

func parseViewStmt(stmt *pg_query.ViewStmt) *View {
	viewName := stmt.GetView().GetRelname()
	if viewName == "" {
		return nil
	}

	view := &View{
		Name:         viewName,
		Materialized: false,
	}

	// Definition parsing can be tricky, might need wasipg.Deparse
	// For views, the definition is a query.
	if stmt.GetQuery() != nil {
		def, _ := deparseStmt(stmt.GetQuery())
		view.Definition = def
	}

	// Columns are often implicit in the query, but sometimes explicit in ViewStmt aliases
	if len(stmt.GetAliases()) > 0 {
		for _, aliasNode := range stmt.GetAliases() {
			if str := aliasNode.GetString_(); str != nil {
				view.Columns = append(view.Columns, str.GetSval())
			}
		}
	} else if stmt.GetQuery() != nil {
		// basic extraction of column names from target list if not aliased at the view level
		if selectStmt := stmt.GetQuery().GetSelectStmt(); selectStmt != nil {
			for _, targetNode := range selectStmt.GetTargetList() {
				target := targetNode.GetResTarget()
				if target != nil {
					if name := target.GetName(); name != "" {
						view.Columns = append(view.Columns, name)
					} else if colRef := target.GetVal().GetColumnRef(); colRef != nil {
						fields := colRef.GetFields()
						if len(fields) > 0 {
							if str := fields[len(fields)-1].GetString_(); str != nil {
								view.Columns = append(view.Columns, str.GetSval())
							}
						}
					}
				}
			}
		}
	}

	return view
}

func parseCreateTableAsStmt(stmt *pg_query.CreateTableAsStmt) *View {
	if stmt.GetObjtype() != pg_query.ObjectType_OBJECT_MATVIEW {
		return nil // We only care if it's creating a materialized view in this context, or maybe a regular table AS? The prompt says CreateTableAsStmt -> View (materialized if indicated). Usually CreateTableAsStmt is used for MATERIALIZED VIEWs in pg_dump.
	}

	into := stmt.GetInto()
	if into == nil || into.GetRel() == nil {
		return nil
	}

	viewName := into.GetRel().GetRelname()
	if viewName == "" {
		return nil
	}

	view := &View{
		Name:         viewName,
		Materialized: stmt.GetObjtype() == pg_query.ObjectType_OBJECT_MATVIEW,
	}

	if stmt.GetQuery() != nil {
		def, _ := deparseStmt(stmt.GetQuery())
		view.Definition = def
	}

	return view
}

func parseCreateEnumStmt(stmt *pg_query.CreateEnumStmt) *EnumType {
	var typeName string
	for _, n := range stmt.GetTypeName() {
		if str := n.GetString_(); str != nil {
			typeName = str.GetSval()
		}
	}

	if typeName == "" {
		return nil
	}

	enum := &EnumType{
		Name: typeName,
	}

	for _, valNode := range stmt.GetVals() {
		if str := valNode.GetString_(); str != nil {
			enum.Values = append(enum.Values, str.GetSval())
		}
	}

	return enum
}

func parseCreateSeqStmt(stmt *pg_query.CreateSeqStmt) *Sequence {
	seqName := stmt.GetSequence().GetRelname()
	if seqName == "" {
		return nil
	}

	seq := &Sequence{
		Name:      seqName,
		DataType:  "bigint", // Default in PG
		Start:     1,
		Increment: 1,
	}

	for _, optNode := range stmt.GetOptions() {
		opt := optNode.GetDefElem()
		if opt == nil {
			continue
		}

		switch opt.GetDefname() {
		case "as":
			if typeNameNode := opt.GetArg().GetTypeName(); typeNameNode != nil {
				dt, _, _, _ := parseTypeName(typeNameNode)
				seq.DataType = dt
			}
		case "start":
			if intVal := opt.GetArg().GetInteger(); intVal != nil {
				seq.Start = int64(intVal.GetIval())
			}
		case "increment":
			if intVal := opt.GetArg().GetInteger(); intVal != nil {
				seq.Increment = int64(intVal.GetIval())
			}
		}
	}

	return seq
}
