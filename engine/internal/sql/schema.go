package sql

type SQLSchema struct {
	Tables    map[string]*Table
	Views     map[string]*View
	Enums     map[string]*EnumType
	Sequences map[string]*Sequence
}

type Table struct {
	Name        string
	Columns     map[string]*Column
	Constraints map[string]*Constraint
	Indexes     map[string]*Index
}

type Column struct {
	Name         string
	DataType     string
	Length       *int
	Precision    *int
	Scale        *int
	NotNull      bool
	Default      *string
	IsPrimaryKey bool
}

type Constraint struct {
	Name       string
	Type       string
	Columns    []string
	CheckExpr  *string
	RefTable   *string
	RefColumns []string
}

type Index struct {
	Name    string
	Columns []string
	Unique  bool
	Where   *string
}

type View struct {
	Name         string
	Definition   string
	Columns      []string
	Materialized bool
}

type EnumType struct {
	Name   string
	Values []string
}

type Sequence struct {
	Name      string
	DataType  string
	Start     int64
	Increment int64
}
