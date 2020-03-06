package samplesheetio

import (
	"fmt"
	"strings"
)

//writer.go defines how to output samplesheet
func (ss *SampleSheet) Write() *SampleSheet {
	ret := new(SampleSheet)
	ret.Reader = ss.Reader
	for _, sec := range ss.Sections {
		topush := sec
		if sec.Writer != nil {
			topush = sec.Writer(sec)
			topush.Name = sec.Name
		}
		ret.Sections = append(ret.Sections, topush)
	}
	for _, row := range ss.Data {
		topushRow := new(Row)
		for _, c := range row.Cells {
			topush := c
			if c.Writer != nil {
				topush = c.Writer(row, c)
			}
			topushRow.Cells = append(topushRow.Cells, topush)
		}
		ret.Data = append(ret.Data, topushRow)
	}
	return ret
}

func (ss *SampleSheet) AddSectionWriter(sectionName string, fn SectionWriter) {
	for _, sec := range ss.Sections {
		if sec.Name == sectionName {

			sec.Writer = fn
			return
		}
	}

	newSection := new(Section)
	newSection.Name = sectionName
	newSection.Writer = fn
	ss.Sections = append(ss.Sections, newSection)

	return
}

func (ss *SampleSheet) trimOrPad(row []string) []string {
	n := len(ss.Reader.ColumnDefs)
	return trimOrPad(row, n, "")
}

func (ss *SampleSheet) _string(endOfLine string) string {
	lines := []string{}
	for _, sec := range ss.Sections {
		//header

		_line := ss.trimOrPad([]string{fmt.Sprintf(`[%s]`, sec.Name)})
		lines = append(lines, strings.Join(_line, ","))

		for _, row := range sec.Rows {
			lines = append(lines, strings.Join(row, ","))
		}
	}
	//Add data
	_dataLine := ss.trimOrPad([]string{`[Data]`})
	lines = append(lines, strings.Join(_dataLine, ","))
	dataHeader := []string{}
	for _, h := range ss.Reader.ColumnDefs {
		dataHeader = append(dataHeader, h.Name)
	}
	lines = append(lines, strings.Join(dataHeader, ","))

	for _, row := range ss.Data {
		rowLine := []string{}
		for _, c := range row.Cells {
			rowLine = append(rowLine, c.Value)
		}
		lines = append(lines, strings.Join(rowLine, ","))
	}

	return strings.Join(lines, endOfLine)
}

func (ss *SampleSheet) String() string {
	return ss._string("\n")
}

func (ss *SampleSheet) WinString() string {
	return ss._string("\r\n")
}

func (self *Section) UpdateOrAppend(row []string) {
	if len(row) == 0 {
		return
	}

	if row[0] == "" {
		self.Rows = append(self.Rows, row)
		return
	}
	k := row[0]

	for i, _row := range self.Rows {
		if len(_row) == 0 {
			continue
		}
		if _row[0] == k {
			self.Rows[i] = row
			return
		}

		if i == len(self.Rows) {
			self.Rows[i] = row
			self.Rows = append(self.Rows, _row)
			return
		}
	}

}

func (ss *SampleSheet) GetSection(sectionName string) *Section {
	for _, sec := range ss.Sections {
		if sec.Name == sectionName {
			return sec
		}
	}
	return nil
}
