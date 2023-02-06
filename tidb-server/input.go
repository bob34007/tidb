package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type getPrams interface {
	Open() error
	ReadSqls() error
	Close() error
	GetSqls() []string
}

type ShellInput struct {
	sqls []string
}

func (s *ShellInput) Open() error {
	return nil
}

func (s *ShellInput) ReadSqls() error {
	s.sqls = append(s.sqls, f.sql)
	return nil
}

func (s *ShellInput) Close() error {
	return nil
}

func (s *ShellInput) GetSqls() []string {
	return s.sqls
}

type FileInput struct {
	filename string
	sqls     []string
}

func (fs *FileInput) Open() error {
	return nil
}

func (fs *FileInput) ReadSqls() error {
	//The current version reads the contents of the file in one sitting.
	ss, err := os.ReadFile(f.input)
	if err != nil {
		return err
	}
	fs.sqls = append(fs.sqls, string(ss))
	return nil
}

func (fs *FileInput) Close() error {
	return nil
}

func (fs *FileInput) GetSqls() []string {
	return fs.sqls
}

type DBInput struct {
	dsn              string
	dbName           string
	tables           []string
	db               *sql.DB
	sqls             []string
	lowerCase        string
	needCheckTabName bool
}

func (dbs *DBInput) Open() error {
	var err error
	dbs.db, err = sql.Open("mysql", dbs.dsn)
	return err
}

func (dbs *DBInput) getTablesFromDB() error {
	var tabName string
	var sql string
	// the feature is invalid when new_collations_enabled_on_first_bootstrap is false
	if dbs.lowerCase == "Y" {
		sql = fmt.Sprintf("select TABLE_NAME from information_schema.tables where TABLE_SCHEMA='%v'  collate utf8mb4_general_ci ", dbs.dbName)
	} else {
		sql = fmt.Sprintf("select TABLE_NAME from information_schema.tables where TABLE_SCHEMA='%v'  ", dbs.dbName)
	}
	rows, err := dbs.db.Query(sql)
	if err != nil {
		fmt.Println(sql)
		return err
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	for rows.Next() {
		err = rows.Scan(&tabName)
		if err != nil {
			return err
		}
		dbs.tables = append(dbs.tables, tabName)
	}
	return err
}

func (dbs *DBInput) checkTableName(tabName string) ([]string, error) {
	names := make([]string, 0, 2)
	if !dbs.needCheckTabName {
		names = append(names, tabName)
		return names, nil
	} else {
		var sql string
		var n string
		// the feature is invalid when new_collations_enabled_on_first_bootstrap is false
		if dbs.lowerCase == "Y" {
			sql = fmt.Sprintf("select TABLE_NAME from information_schema.tables where TABLE_SCHEMA='%v'  "+
				" collate utf8mb4_general_ci and table_name  ='%v' collate utf8mb4_general_ci ", dbs.dbName, tabName)
		} else {
			sql = fmt.Sprintf("select TABLE_NAME from information_schema.tables where "+
				"TABLE_SCHEMA='%v' and table_name  ='%v'  ", dbs.dbName, tabName)
		}
		rows, err := dbs.db.Query(sql)
		if err != nil {
			return nil, err
		}
		defer func() {
			if rows != nil {
				rows.Close()
			}
		}()
		for rows.Next() {
			err = rows.Scan(&n)
			if err != nil {
				return names, err
			}
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("%v.%v is not exist in lowerCase %v mode", dbs.dbName, tabName, dbs.lowerCase)
	}
	return names, nil
}

func (dbs *DBInput) getTableDDL(tabName string) error {
	var ddl string
	names, err := dbs.checkTableName(tabName)
	if err != nil {
		return err
	}
	for _, tName := range names {
		sql := fmt.Sprintf("show create table %v", dbs.dbName+"."+tName)
		rows, err := dbs.db.Query(sql)
		if err != nil {
			return err
		}
		defer func() {
			if rows != nil {
				rows.Close()
			}
		}()

		for rows.Next() {
			err = rows.Scan(&tabName, &ddl)
			if err != nil {
				return err
			}
			dbs.sqls = append(dbs.sqls, ddl)
		}
	}
	return nil
}

func (dbs *DBInput) ReadSqls() error {
	if len(dbs.tables) == 0 {
		err := dbs.getTablesFromDB()
		if err != nil {
			return err
		}
	}

	for _, v := range dbs.tables {
		tabName := strings.TrimSpace(v)
		err := dbs.getTableDDL(tabName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dbs *DBInput) Close() error {
	return dbs.db.Close()
}

func (dbs *DBInput) GetSqls() []string {
	return dbs.sqls
}
