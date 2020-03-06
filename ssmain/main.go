package main

import (
	"fmt"
	"log"

	"encoding/json"

	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ws6/interop/samplesheetio"
	"github.com/ws6/interop/samplesheetio/samplesheets"
)

func main() {

	TestSampleSheets()
}

func TestSampleSheets() {
	io := samplesheets.GetIO(samplesheets.TST170, samplesheets.VERSION_TST170_1)
	if io == nil {
		fmt.Println(`can not find reader io`)
		return
	}
	dir := `../test_data/samplesheets`
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err.Error())
	}

	ss_files := []string{}

	for i, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), `.csv`) {
			continue
		}

		ss_files = append(ss_files, filepath.Join(dir, file.Name()))
		if i > 10 {
			break
		}
	}

	for _, ss_file := range ss_files {
		log.Println(`parsing`, ss_file)
		ssByte, err := ioutil.ReadFile(ss_file)
		if err != nil {
			log.Fatal(err.Error())
		}
		ss, err := samplesheets.Read(io, string(ssByte))
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf(`%+v`, ss)

		ssString := samplesheets.Write(io, ss).PadSections().String()
		log.Println(ssString)
		ioutil.WriteFile(filepath.Base(ss_file), []byte(ssString), 0644)
	}
}

func TestSampleSheetReader() {

	dir := `../test_data/samplesheets`
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err.Error())
	}

	ss_files := []string{}

	for i, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), `.csv`) {
			continue
		}

		ss_files = append(ss_files, filepath.Join(dir, file.Name()))
		if i > 10 {
			break
		}
	}

	reader := samplesheetio.NewReader(

		[]*samplesheetio.ColumnDef{
			{
				Name: `Sample_ID`, Accepts: []string{`Sample_ID`, `SampleID`}, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
			},
			{
				Name: `Sample_Name`,
			},
			//example of stub field and writer
			//if writer is nil, it will fail on read
			{
				Name: `Field_Notexsts`, Writer: func(row *samplesheetio.Row, cell *samplesheetio.Cell) *samplesheetio.Cell {

					c := row.GetCell(&samplesheetio.ColumnDef{Name: `Sample_ID`})
					cell.Value = fmt.Sprintf(`HP-%s`, c.Value)
					return cell
				},
			},
			{
				Name: `Sample_Plate`,
			},
			{
				Name: `Sample_Well`,
			},
			{
				Name: `I7_Index_ID`,
			},
			{
				Name: `index`, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
			},
			{
				Name: `I5_Index_ID`,
			},
			{
				Name: `index2`, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
			},
			{
				Name: `Sample_Project`,
			},
			{
				Name: `Description`,
			},
			{
				Name: `Manifest`, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
				Validators: []samplesheetio.Validator{
					func(row *samplesheetio.Row, cell *samplesheetio.Cell) error {
						if cell.Value == `PoolDNA` {
							return nil
						}
						if cell.Value == `PoolRNA` {
							return nil
						}
						return fmt.Errorf(`either PoolRNA or PoolRNA`)
					},
				},
			},
		},
	)
	for _, ss_file := range ss_files {
		ssByte, err := ioutil.ReadFile(ss_file)
		if err != nil {
			log.Fatal(err.Error())
		}
		ss, err := reader.Read(string(ssByte))
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf(`section len  = %d`, len(ss.Sections))
		log.Printf(`%+v`, ss.Sections)

		log.Printf(`%+v`, ss)

		//trigger the writer  functions
		ss.AddSectionWriter(`Manifests`, func(sec *samplesheetio.Section) *samplesheetio.Section {
			fmt.Println(`writer of Manifests`)
			ret := new(samplesheetio.Section)
			ret.Rows = [][]string{
				[]string{`PoolDNA`, `Release2HDNA_manifest_160428.txt`},
				[]string{`PoolDNA`, `Release2HRNA_manifest_160428.txt`},
				[]string{``},
			}

			return ret

		},
		)
		log.Println(`reader len`, len(ss.Reader.ColumnDefs))
		_ss := ss.Write().PadSections()
		log.Println(`new reader len`, len(ss.Reader.ColumnDefs))
		b, err := json.MarshalIndent(_ss, "", "  ")
		if err != nil {
			log.Fatal(err.Error())
		}
		_ = b
		// log.Printf(`%s`, string(b))
		ssString := _ss.String()
		ioutil.WriteFile(filepath.Base(ss_file), []byte(ssString), 0644)
	}
}
