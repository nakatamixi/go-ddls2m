package ddls2m

import (
	"fmt"
	"math/big"
	"strings"

	"cloud.google.com/go/spanner/spansql"
	"github.com/knocknote/vitess-sqlparser/sqlparser"
)

var (
	UnicodeMaxByte      = big.NewInt(4)
	MysqlVarcharMaxByte = big.NewInt(65535)
)

func Convert(sqls string, debug bool) {
	// spansql not allow backquote
	sqls = strings.Replace(sqls, "`", "", -1)
	d, err := spansql.ParseDDL(sqls)
	if err != nil {
		panic(err)
	}
	for _, v := range d.List {
		fmt.Println(ConvertStmt(d, v) + ";")
	}
}

func ConvertStmt(d spansql.DDL, s spansql.DDLStmt) string {
	switch v := s.(type) {
	case spansql.CreateTable:
		return ConvertTable(d, v)
	case spansql.CreateIndex:
		return ConvertIndex(v)
	default:
		panic(fmt.Sprintf("donot support %T", v))
	}
	return ""
}

func ConvertTable(d spansql.DDL, t spansql.CreateTable) string {
	stmt := convertFromCreateTableStmt(d, t)
	tbuf := sqlparser.NewTrackedBuffer(func(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {})
	stmt.Format(tbuf)
	return string(tbuf.Buffer.String())
}
func ConvertIndex(i spansql.CreateIndex) string {
	return i.SQL()
}

func convertFromCreateTableStmt(d spansql.DDL, t spansql.CreateTable) sqlparser.Statement {
	columns := []*sqlparser.ColumnDef{}
	for _, col := range t.Columns {
		options := []*sqlparser.ColumnOption{}
		if col.NotNull {
			options = append(options, &sqlparser.ColumnOption{
				Type: sqlparser.ColumnOptionNotNull,
			})
		}
		columns = append(columns, &sqlparser.ColumnDef{
			Name:    col.Name,
			Type:    ConvertType(col.Type),
			Options: options,
		})
	}
	constraints := []*sqlparser.Constraint{}
	if len(t.PrimaryKey) > 0 {
		keys := []sqlparser.ColIdent{}
		for _, pkey := range t.PrimaryKey {
			keys = append(keys, sqlparser.NewColIdent(pkey.Column))
		}
		constraints = append(constraints, &sqlparser.Constraint{
			Type: sqlparser.ConstraintPrimaryKey,
			Keys: keys,
		})
	}
	if t.Interleave != nil {
		parent := findTable(d, t.Interleave.Parent)
		keys := []sqlparser.ColIdent{}
		if len(parent.PrimaryKey) > 0 {
			for _, pkey := range parent.PrimaryKey {
				keys = append(keys, sqlparser.NewColIdent(pkey.Column))
			}
		}
		constraints = append(constraints, &sqlparser.Constraint{
			Type: sqlparser.ConstraintForeignKey,
			Keys: keys,
			Reference: &sqlparser.Reference{
				Name: t.Interleave.Parent,
				Keys: keys,
			},
		})
	}
	return &sqlparser.CreateTable{
		DDL: &sqlparser.DDL{
			Action: "create",
			NewName: sqlparser.TableName{
				Name: sqlparser.NewTableIdent(t.Name),
			},
		},
		Columns:     columns,
		Constraints: constraints,
		Options: []*sqlparser.TableOption{
			&sqlparser.TableOption{
				Type:     sqlparser.TableOptionEngine,
				StrValue: "InnoDB",
			},
			&sqlparser.TableOption{
				Type:     sqlparser.TableOptionCharset,
				StrValue: "utf8mb4",
			},
		},
	}
}
func findTable(d spansql.DDL, t string) *spansql.CreateTable {
	for _, v := range d.List {
		if c, ok := v.(spansql.CreateTable); ok {
			if t == c.Name {
				return &c
			}
		}
	}
	panic(fmt.Sprintf("cant find table %s", t))
	return nil
}

func ConvertType(t spansql.Type) string {
	if t.Base == spansql.Bool {
		return "BOOL"
	}
	if t.Base == spansql.Int64 {
		return "BIGINT"
	}
	if t.Base == spansql.Float64 {
		return "FLOAT"
	}
	if t.Base == spansql.String {
		l := big.NewInt(t.Len)
		bytes := new(big.Int).Mul(l, UnicodeMaxByte)
		if bytes.Cmp(MysqlVarcharMaxByte) > 0 {
			return "TEXT"
		}
		return fmt.Sprintf("VARCHAR(%d)", t.Len)
	}
	if t.Base == spansql.Bytes {
		return "BLOB"
	}
	if t.Base == spansql.Date {
		return "DATE"
	}
	if t.Base == spansql.Timestamp {
		return "DATETIME"
	}
	panic(fmt.Sprintf("do not support %T", t.Base))
	return ""
}
