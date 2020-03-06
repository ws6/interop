package samplesheets

import (
	"github.com/ws6/interop/samplesheetio"
)

const (
	TSO500           = `TSO500`
	VERSION_TSO500_2 = `v2.0`
)

func init() {
	Register(&TSO500IO{
		Name:    TSO500,
		Version: VERSION_TSO500_2,
	})
}

type TSO500IO struct {
	Name    string
	Version string
}

func (self *TSO500IO) GetNameVersion() (string, string) {
	return self.Name, self.Version
}

func updateTSO500SettingsSection(sec *samplesheetio.Section) *samplesheetio.Section {

	ret := new(samplesheetio.Section)
	ret.Rows = [][]string{
		[]string{`AdapterRead1`, `AGATCGGAAGAGCACACGTCTGAACTCCAGTCA`},
		[]string{`AdapterRead2`, `AGATCGGAAGAGCGTCGTGTAGGGAAAGAGTGT`},
		[]string{`AdapterBehavior`, `trim`},
		[]string{`MinimumTrimmedReadLength`, `35`},
		[]string{`MaskShortReads`, `35`},
		[]string{`OverrideCycles`, `U7N1Y93;I8;I8;U7N1Y93`},
		[]string{``},
	}

	return ret

}
func (self *TSO500IO) MakeWriter(ss *samplesheetio.SampleSheet) *samplesheetio.SampleSheet {
	headerSection := ss.GetSection(`Header`)
	if headerSection != nil {
		headerSection.UpdateOrAppend([]string{
			`Workflow`, `GenerateFASTQ`,
		})
	}
	ss.AddSectionWriter(`Settings`, updateTSO500SettingsSection)
	return ss
}

func (self *TSO500IO) MakeReader() *samplesheetio.Reader {

	return samplesheetio.NewReader(

		[]*samplesheetio.ColumnDef{
			{
				Name: `Sample_ID`, Accepts: []string{`Sample_ID`, `SampleID`}, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
			},
			{
				Name: `Pair_ID`,
				Writer: func(row *samplesheetio.Row, cell *samplesheetio.Cell) *samplesheetio.Cell {
					sampleIdCell := row.GetCell(&samplesheetio.ColumnDef{Name: `Sample_ID`})
					cell.Value = sampleIdCell.Value
					return cell
				},
			},

			{
				Name: `Sample_Type`,
			},
			{
				Name: `Sex`,
			},
			{
				Name: `Tumor_Type`,
			},
			{
				Name: `Description`,
			},
			{
				Name: `Index_ID`,
			},
			{
				Name: `index`, StopWhenEmtpy: true, ErrorOnMissingFromHeader: true,
			},
			{
				Name: `I7_Index_ID`,
			},
			{
				Name: `index2`,
			},
			{
				Name: `I5_Index_ID`,
			},
		},
	)

}

func (self *TSO500IO) Read(body string) error {
	return nil
}
