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
	if strings.TrimSpace(f.sql) == "" && strings.TrimSpace(f.input) == "" {
		return errors.New("sql and input cannot be empty at the same time")
	}
	if strings.TrimSpace(f.sql) != "" && strings.TrimSpace(f.input) != "" {
		return errors.New("sql and input cannot be specified at the same time")
	}
	return nil
}

func main() {
	res := make([]string, 0, 10)
	flag.Parse()
	fmt.Println(f)
	mustNil(checkConfig())
	sqls, err := getSQLs()
	mustNil(err)
	if strings.TrimSpace(sqls) == "" {
		mustNil(errors.New("input sql is nil"))
	}
	stmts, err := parse(sqls)
	mustNil(err)
	var buf bytes.Buffer
	for _, stmt := range stmts {
		buf.Reset()
		restoreCtx := format.NewRestoreCtx(format.DefaultRestoreFlags, &buf)
		err := stmt.Restore(restoreCtx)
		mustNil(err)
		res = append(res, buf.String())
	}
	writeResult(res)
}
