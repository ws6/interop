package samplesheetio

//writer.go defines how to output samplesheet
func (ss *SampleSheet) Write() *SampleSheet {
	ret := new(SampleSheet)
	for _, sec := range ss.Sections {
		topush := sec
		if sec.Writer != nil {
			topush = sec.Writer(sec)
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
