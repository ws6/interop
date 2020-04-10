package samplesheetio

//reader.go aims reading varidac samplesheet schema

import (
	"fmt"
	"regexp"
	"strings"
)

var ERR_EMPTY = fmt.Errorf(`Sample Sheet is empty`)
var ERR_MISSING_HEADER = fmt.Errorf(`missing required header`)
var ERR_STOP = fmt.Errorf(`stop on cell value is empty`)

type Reader struct {
	// SectionNames []string
	ColumnDefs []*ColumnDef
}

type SectionWriter func(*Section) *Section

func trimOrPad(in []string, n int, placeholder string) []string {
	ret := []string{}
	for i := 0; i < n; i++ {
		if i >= len(in) {
			ret = append(ret, placeholder)
			continue
		}

		ret = append(ret, in[i])
	}
	return ret
}

type Section struct {
	Name   string
	Rows   [][]string
	Writer SectionWriter `json:"-"`
}

func (self *Section) TrimOrPad(n int, placeholder string) {
	for i, row := range self.Rows {
		self.Rows[i] = trimOrPad(row, n, placeholder)
	}
}

type Cell struct {
	*ColumnDef `json:"-"`
	Name       string
	Value      string
}

type Row struct {
	Cells []*Cell
}

func (self *Row) GetCellByName(n string) *Cell {
	h := &Cell{
		Name: n,
	}
	for _, c := range self.Cells {

		if c.ColumnDef.Name == h.Name {
			return c
		}
	}
	return nil
}

func (self *Row) GetCell(h *ColumnDef) *Cell {
	for _, c := range self.Cells {
		if c.ColumnDef == h {
			return c
		}
		if c.ColumnDef.Name == h.Name {
			return c
		}
	}
	return nil
}

type SampleSheet struct {
	Sections []*Section
	Data     []*Row

	*Reader `json:"-"`
}

type Validator func(*Row, *Cell) error
type WriterFn func(*Row, *Cell) *Cell
type ColumnDef struct {
	Name                     string   //consolidated name
	Position                 int      //1-offset to avoid default initialization
	Accepts                  []string //any strings can accept
	StopWhenEmtpy            bool
	ErrorOnMissingFromHeader bool //return error when can not find such header

	//TODO add validator func and writer func
	Validators []Validator `json:"-"`
	Writer     WriterFn    `json:"-"`
}

func NewReader(columns []*ColumnDef) *Reader {
	ret := new(Reader)

	for _, c := range columns {
		ret.AddColumnDef(c)
	}

	return ret
}

func (self *Reader) AddColumnDef(col *ColumnDef) {
	if len(col.Accepts) == 0 {
		col.Accepts = append(col.Accepts, col.Name)
	}

	self.ColumnDefs = append(self.ColumnDefs, col)
}

func setPosition(hd *ColumnDef, fields []string) int {
	for pos, f := range fields {
		for _, ac := range hd.Accepts {
			if strings.ToLower(ac) == strings.ToLower(f) {
				return pos + 1
			}
		}
	}
	return -1
}

func (self *Reader) ParseDataHeader(ln []string) error {

	for _, h := range self.ColumnDefs {
		pos := setPosition(h, ln)
		if pos < 0 && h.ErrorOnMissingFromHeader {
			return fmt.Errorf(`Missing [Data] header [%s]: it can use any of [%s]`, h.Name, strings.Join(h.Accepts, ", "))
		}
		h.Position = pos
	}

	return nil
}

func (self *Reader) ParseRow(ln []string) (*Row, error) {
	ret := new(Row)
	sz := len(ln)
	for _, h := range self.ColumnDefs {
		topush := new(Cell)
		topush.Name = h.Name
		topush.ColumnDef = h
		ret.Cells = append(ret.Cells, topush)

		if (h.Position - 1) < 0 {
			// if h.Writer == nil {

			// 	return nil, fmt.Errorf(`position is too low`)
			// }
			continue
		}
		if h.Position > sz {
			return nil, fmt.Errorf(`position is too high`)
		}

		value := ln[h.Position-1]
		value = strings.TrimSpace(value)
		if len(value) == 0 {
			if h.StopWhenEmtpy {
				return nil, ERR_STOP
			}
		}
		topush.Value = value

	}

	for _, cell := range ret.Cells {
		for _, val := range cell.Validators {
			if err := val(ret, cell); err != nil {
				return nil, fmt.Errorf(`Validator err on cell[%+v]: %s`, cell, err.Error())
			}
		}
	}
	return ret, nil
}

func (self *Reader) ReadDataSection(ret *SampleSheet, rows []string) error {
	ret.Data = []*Row{}
	dataStarted := false
	cnt := 0

	// parse data section

	for _, ln := range rows {
		if len(ln) == 0 {
			continue
		}
		if strings.HasPrefix(ln, "[Data]") {
			dataStarted = true
			//parse data

			continue
		}
		if !dataStarted {
			continue
		}
		if strings.HasPrefix(ln, `#`) { //allow comments
			continue
		}
		cnt++
		//Start processing
		dataLn := strings.Split(ln, ",")

		if cnt == 1 {
			if err := self.ParseDataHeader(dataLn); err != nil {
				return err
			}
			continue

		}
		row, err := self.ParseRow(dataLn)
		if err != nil {
			if err == ERR_STOP {
				break
			}
			return err
		}
		ret.Data = append(ret.Data, row)

	}
	return nil
}

func ReadSection(ss []string, sectionName string) (*Section, error) {
	dataStarted := false
	ret := new(Section)
	ret.Name = sectionName
	ret.Rows = [][]string{}
	for _, ln := range ss {

		if len(ln) == 0 {

			continue
		}
		if strings.Contains(ln, fmt.Sprintf(`[%s]`, sectionName)) {
			dataStarted = true
			continue
		}
		// //for BOM at the first byte of file
		// if sectionName == `Header` && strings.Contains(ln, sectionName) {
		// 	dataStarted = true
		// 	continue
		// }

		if dataStarted && strings.Contains(ln, "[") {

			break
		}

		if !dataStarted {
			continue
		}

		dataLn := strings.Split(ln, ",")

		ret.Rows = append(ret.Rows, dataLn)

	}

	return ret, nil
}

var FindSectionName = func() func(string) string {

	r, err := regexp.Compile(`\[([a-z|A-Z]+)\]`)
	if err != nil {
		fmt.Println(`err compile regexp`, err.Error())

	}

	return func(line string) string {
		founds := r.FindStringSubmatch(line)
		if founds == nil {
			return ""
		}
		if len(founds) < 2 {
			return ""
		}

		return founds[1]
	}

}()

func SearchSectionNames(rows []string) []string {
	ret := []string{}
	for _, line := range rows {
		secName := FindSectionName(line)
		if secName != "" {
			ret = append(ret, secName)
		}
	}

	return ret
}

func (self *Reader) ReadAllSections(sectionNames []string, ss *SampleSheet, rows []string, noData bool) error {
	ss.Sections = []*Section{}
	for _, sectionName := range sectionNames {
		if noData && sectionName == `Data` {
			continue
		}
		section, err := ReadSection(rows, sectionName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		ss.Sections = append(ss.Sections, section)
	}

	return nil
}

func (self *Reader) Read(sampleSheetBody string) (*SampleSheet, error) {

	ssSplit := strings.Split(string(sampleSheetBody), "\n")
	trimedSplit := []string{}
	for _, s := range ssSplit {
		if accepted, ts := trimSamplSheetLine(s); accepted {
			trimedSplit = append(trimedSplit, ts)
		}

	}

	if len(trimedSplit) == 0 {

		return nil, ERR_EMPTY
	}
	sectionNames := SearchSectionNames(trimedSplit)
	if len(sectionNames) == 0 {
		return nil, fmt.Errorf(`Can not find Sections names`)
	}
	ret := new(SampleSheet)
	skipData := true
	if err := self.ReadAllSections(sectionNames, ret, trimedSplit, skipData); err != nil {
		return nil, err
	}

	if err := self.ReadDataSection(ret, trimedSplit); err != nil {
		return nil, err
	}
	ret.Reader = self
	return ret, nil

}

func trimSamplSheetLine(s string) (bool, string) {
	r := strings.TrimRight(s, "\r ")
	//skip empty
	if len(r) == 0 {
		return false, r
	}
	//skip # comments
	if strings.HasPrefix(r, "#") {
		return false, r
	}
	r = strings.Trim(r, " ")
	//skip trimmed empty
	if len(r) == 0 {
		return false, r
	}
	return true, r
}

func (self *SampleSheet) PadSections() *SampleSheet {
	numOfDataFields := len(self.Reader.ColumnDefs)
	for _, sec := range self.Sections {
		sec.TrimOrPad(numOfDataFields, "")
	}
	return self
}
