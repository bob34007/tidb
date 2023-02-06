package main

import (
	"fmt"
	"os"
)

type WriteRes interface {
	Open() error
	Write(data string) error
	Close() error
}

type Windows struct {
}

func (w *Windows) Open() error {
	return nil
}

func (w *Windows) Write(data string) error {
	fmt.Println(data, ";")
	return nil
}

func (w *Windows) Close() error {
	return nil
}

type Files struct {
	filename string //filepath + filename
	fp       *os.File
}

func (f *Files) Open() error {
	var err error
	f.fp, err = os.OpenFile(f.filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (f *Files) Write(data string) error {
	_, err := f.fp.WriteString(data + ";\n")
	if err != nil {
		fmt.Println("write data failed,", err, data)
		return err
	}
	//f.fp.Sync()
	return nil
}

func (f *Files) Close() error {
	return f.fp.Close()
}
