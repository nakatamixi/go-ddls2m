package ddls2m

import (
	"fmt"
	"math/big"
	"strings"

	"cloud.google.com/go/spanner/spansql"
	"github.com/knocknote/vitess-sqlparser/sqlparser"
	"golang.org/x/xerrors"
)

var (
	UnicodeMaxByte      = big.NewInt(4)
	MysqlVarcharMaxByte = big.NewInt(65535)
)

func Convert(sqls string) (string, error) {
	// spansql not allow backquote
	sqls = strings.Replace(sqls, "`", "", -1)
	d, err := spansql.ParseDDL(sqls)
	if err != nil {
		return "", err
	}
	mysqlDDL := ""
	for _, v := range d.List {
		stmt, err := ConvertStmt(d, v)
		if err != nil {
			return "", err
		}
		mysqlDDL = mysqlDDL + stmt + ";\n"
	}
	return mysqlDDL, nil
}

func ConvertStmt(d spansql.DDL, s spansql.DDLStmt) (string, error) {
	switch v := s.(type) {
	case spansql.CreateTable:
		return ConvertTable(d, v)
	case spansql.CreateIndex:
		return ConvertIndex(v), nil
	default:
		return "", xerrors.New(fmt.Sprintf("donot support %T", v))
	}
}

func ConvertTable(d spansql.DDL, t spansql.CreateTable) (string, error) {
	stmt, err := convertFromCreateTableStmt(d, t)
	if err != nil {
		return "", err
	}
	tbuf := sqlparser.NewTrackedBuffer(func(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {})
	stmt.Format(tbuf)
	return string(tbuf.Buffer.String()), nil
}
func ConvertIndex(i spansql.CreateIndex) string {
	return i.SQL()
}

func convertFromCreateTableStmt(d spansql.DDL, t spansql.CreateTable) (sqlparser.Statement, error) {
	columns := []*sqlparser.ColumnDef{}
	for _, col := range t.Columns {
		options := []*sqlparser.ColumnOption{}
		if col.NotNull {
			options = append(options, &sqlparser.ColumnOption{
				Type: sqlparser.ColumnOptionNotNull,
			})
		}
		colType, err := ConvertType(col.Type)
		if err != nil {
			return nil, err
		}
		columns = append(columns, &sqlparser.ColumnDef{
			Name:    col.Name,
			Type:    colType,
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
		parent, err := findTable(d, t.Interleave.Parent)
		if err != nil {
			return nil, err
		}
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
	}, nil
}
func findTable(d spansql.DDL, t string) (*spansql.CreateTable, error) {
	for _, v := range d.List {
		if c, ok := v.(spansql.CreateTable); ok {
			if t == c.Name {
				return &c, nil
			}
		}
	}
	return nil, xerrors.New(fmt.Sprintf("cant find table %s", t))
}

func ConvertType(t spansql.Type) (string, error) {
	switch t.Base {
	case spansql.Bool:
		return "BOOL", nil
	case spansql.Int64:
		return "BIGINT", nil
	case spansql.Float64:
		return "FLOAT", nil
	case spansql.String:
		l := big.NewInt(t.Len)
		bytes := new(big.Int).Mul(l, UnicodeMaxByte)
		if bytes.Cmp(MysqlVarcharMaxByte) > 0 {
			return "TEXT", nil
		}
		return fmt.Sprintf("VARCHAR(%d)", t.Len), nil
	case spansql.Bytes:
		return "BLOB", nil
	case spansql.Date:
		return "DATE", nil
	case spansql.Timestamp:
		return "DATETIME", nil
	}
	return "", xerrors.New(fmt.Sprintf("do not support %T", t.Base))
}
