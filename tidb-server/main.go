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
	infile string // inout file
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

func main() {
	res := make([]string, 0, 10)
	flag.Parse()
	fmt.Println(f)
	stmts, err := parse(f.sql)
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
