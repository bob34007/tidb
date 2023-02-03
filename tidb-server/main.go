package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/format"
)

var f flags

type flags struct {
	sql    string // sql to be desensitized
	output string // output file
	input  string // in file
	dsn    string // MySQL DSN
	db     string //database name
	tables string //table name list
}

func mustNil(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func init() {
	flag.StringVar(&f.sql, "sql", " ", "sql to be desensitized")
	flag.StringVar(&f.output, "output", " ", "output file")
	flag.StringVar(&f.input, "input", "", "input file")
	flag.StringVar(&f.dsn, "dsn", "", "mysql dsn")
	flag.StringVar(&f.db, "db", "", "database name")
	flag.StringVar(&f.tables, "tables", "", "table  list")
}

func parse(sql string) ([]ast.StmtNode, error) {
	p := parser.New()

	stmtNodes, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	if len(stmtNodes) == 0 {
		return stmtNodes, errors.New("parse sql result is nil , " + sql)
	}

	return stmtNodes, nil
}

func newWriteRes() WriteRes {
	switch strings.TrimSpace(f.output) {
	case "":
		return &Windows{}
	default:
		return &Files{
			filename: f.output,
		}
	}
}

func writeResult(res []string) {
	wrs := newWriteRes()
	mustNil(wrs.Open())
	for _, d := range res {
		mustNil(wrs.Write(d))
	}
	mustNil(wrs.Close())
}

func checkConfig() error {
	f.sql = strings.TrimSpace(f.sql)
	f.input = strings.TrimSpace(f.input)
	f.dsn = strings.TrimSpace(f.dsn)
	if f.sql == "" && f.input == "" && f.dsn == "" {
		return errors.New("sql and input and dsn cannot be empty at the same time")
	}
	if f.sql != "" && f.input != "" && f.dsn != "" {
		return errors.New("sql and input and dsn cannot be specified at the same time")
	}
	return nil
}

func newgetPrams() getPrams {
	if strings.TrimSpace(f.sql) != "" {
		fmt.Println("sql in shell")
		return &ShellInput{
			sqls: make([]string, 0, 10),
		}
	} else {
		if strings.TrimSpace(f.input) != "" {
			fmt.Println("sql in file")
			return &FileInput{
				sqls:     make([]string, 0, 10),
				filename: f.input,
			}
		} else {
			fmt.Println("sql in db")
			var tabs []string
			if strings.TrimSpace(f.tables) != "" {
				tabs = strings.Split(strings.TrimSpace(f.tables), ",")
			} else {
				tabs = make([]string, 0, 10)
			}
			return &DBInput{
				sqls:   make([]string, 0, 10),
				dsn:    strings.TrimSpace(f.dsn),
				dbName: strings.TrimSpace(f.db),
				tables: tabs,
			}
		}
	}
}
func getInputSqls() ([]string, error) {
	inparm := newgetPrams()
	err := inparm.Open()
	if err != nil {
		return nil, err
	}
	err = inparm.ReadSqls()
	if err != nil {
		return nil, err
	}
	err = inparm.Close()
	if err != nil {
		return nil, err
	}
	return inparm.GetSqls(), nil
}

func main() {
	res := make([]string, 0, 10)
	flag.Parse()
	fmt.Println(f)
	mustNil(checkConfig())
	sqls, err := getInputSqls()
	mustNil(err)
	var buf bytes.Buffer
	for _, sql := range sqls {
		if strings.TrimSpace(sql) == "" {
			mustNil(errors.New("input sql is nil"))
		}
		stmts, err := parse(sql)
		mustNil(err)
		for _, stmt := range stmts {
			buf.Reset()
			restoreCtx := format.NewRestoreCtx(format.DefaultRestoreFlags, &buf)
			err := stmt.Restore(restoreCtx)
			mustNil(err)
			res = append(res, buf.String())
		}
	}
	writeResult(res)
}
