package samplesheets

import (
	"fmt"

	"github.com/ws6/interop/samplesheetio"
)

const (
	TST170           = `TST170`
	VERSION_TST170_1 = `v1.0`
)

func init() {
	Register(&TST170V1{Name: TST170, Version: VERSION_TST170_1})
}

type TST170V1 struct {
	Name    string
	Version string
}

func (self *TST170V1) GetNameVersion() (string, string) {
	return self.Name, self.Version
}

func updateTstManifestSection(sec *samplesheetio.Section) *samplesheetio.Section {

	ret := new(samplesheetio.Section)
	ret.Rows = [][]string{
		[]string{`PoolDNA`, `Release2HDNA_manifest_160428.txt`},
		[]string{`PoolDNA`, `Release2HRNA_manifest_160428.txt`},
		[]string{``},
	}

	return ret

}
func (self *TST170V1) MakeWriter(ss *samplesheetio.SampleSheet) *samplesheetio.SampleSheet {
	ss.AddSectionWriter(`Manifests`, updateTstManifestSection)
	return ss
}

func (self *TST170V1) MakeReader() *samplesheetio.Reader {

	return samplesheetio.NewReader(

		[]*samplesheetio.ColumnDef{
			{
				Name: `Sample_ID`, Accepts: []string{`Sample_ID`, `SampleID`}, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
			},
			{
				Name: `Sample_Name`,
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

}

func (self *TST170V1) Read(body string) error {
	return nil
}
