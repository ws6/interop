package interop

import (
	"bufio"
	"encoding/binary"
	"os"
)

type ControlMetrics struct {
	LaneNum        uint16
	TileNum        uint16
	Read           uint16
	Sz_ControlName uint16
	ControlName    string
	Sz_IndexName   uint16
	IndexName      string
	NumClusters    uint32
}

type ControlInfo struct {
	filename string
	Version  uint8
	SSize    uint8
	Metrics  []*ControlMetrics
	err      error
}

func (self *ControlInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	buffer := bufio.NewReader(file)

	self.err = binary.Read(buffer, binary.LittleEndian, &self.Version)
	if self.err != nil {
		return self.err
	}
	for {
		if self.err != nil {
			if self.err.Error() == "EOF" {
				return nil
			}
			break
		}

		m := new(ControlMetrics)
		self.err = binary.Read(buffer, binary.LittleEndian, &m.LaneNum)
		if self.err != nil {
			continue
		}
		self.err = binary.Read(buffer, binary.LittleEndian, &m.TileNum)
		if self.err != nil {
			continue
		}
		self.err = binary.Read(buffer, binary.LittleEndian, &m.Read)
		if self.err != nil {
			continue
		}

		self.err = binary.Read(buffer, binary.LittleEndian, &m.Sz_ControlName)
		if self.err != nil {
			continue
		}
		controlName := make([]byte, m.Sz_ControlName)
		self.err = binary.Read(buffer, binary.LittleEndian, controlName)
		if self.err != nil {
			continue
		}
		m.ControlName = string(controlName)

		self.err = binary.Read(buffer, binary.LittleEndian, &m.Sz_IndexName)
		if self.err != nil {
			continue
		}
		indexName := make([]byte, m.Sz_IndexName)
		self.err = binary.Read(buffer, binary.LittleEndian, indexName)
		if self.err != nil {
			continue
		}
		m.IndexName = string(indexName)

		self.err = binary.Read(buffer, binary.LittleEndian, &m.NumClusters)
		if self.err != nil {
			continue
		}

		//append
		if self.err == nil {
			self.Metrics = append(self.Metrics, m)
		}
	}
	return self.err
}
