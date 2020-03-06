package samplesheets

import (
	"sync"

	"github.com/ws6/interop/samplesheetio"
)

//samplesheets.go define interface and driver to use

type SampleSheetIO interface {
	MakeWriter(*samplesheetio.SampleSheet) *samplesheetio.SampleSheet
	MakeReader() *samplesheetio.Reader
	GetNameVersion() (string, string)
}

var Register, GetIO = func() (
	func(io SampleSheetIO),
	func(string, string) SampleSheetIO,

) {
	cache := make(map[string]map[string]SampleSheetIO)
	var lock = &sync.Mutex{}
	return func(io SampleSheetIO) {
			name, version := io.GetNameVersion()
			lock.Lock()
			if _, ok := cache[name]; !ok {
				cache[name] = make(map[string]SampleSheetIO)
			}

			cache[name][version] = io

			lock.Unlock()
		},
		func(name, version string) SampleSheetIO {
			lock.Lock()
			if v, ok := cache[name]; ok {
				if ret, ok2 := v[version]; ok2 {
					return ret
				}
			}
			lock.Unlock()
			return nil
		}
}()

func Read(io SampleSheetIO, body string) (*samplesheetio.SampleSheet, error) {
	reader := io.MakeReader()
	return reader.Read(body)
}

func Write(io SampleSheetIO, ss *samplesheetio.SampleSheet) *samplesheetio.SampleSheet {
	return io.MakeWriter(ss).Write()
}
