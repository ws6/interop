package interop

import (
	"bufio"
	"encoding/binary"
	"os"
)

type IndexMetrics struct {
	LaneNum        uint16
	TileNum        uint16
	Read           uint16
	Sz_IndexName   uint16
	IndexName      string
	Clusters_PF    uint32 //number of clusters passing filter
	Sz_SampleName  uint16
	SampleName     string
	Sz_ProjectName uint16
	ProjectName    string
}

type IndexInfo struct {
	filename string
	Version  uint8
	SSize    uint8
	Metrics  []*IndexMetrics
	err      error
}

func (self *IndexInfo) Parse() error {
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

		m := new(IndexMetrics)
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
		self.err = binary.Read(buffer, binary.LittleEndian, &m.Clusters_PF)
		if self.err != nil {
			continue
		}

		self.err = binary.Read(buffer, binary.LittleEndian, &m.Sz_SampleName)
		if self.err != nil {
			continue
		}
		sampleName := make([]byte, m.Sz_SampleName)
		self.err = binary.Read(buffer, binary.LittleEndian, sampleName)
		if self.err != nil {
			continue
		}
		m.SampleName = string(sampleName)

		self.err = binary.Read(buffer, binary.LittleEndian, &m.Sz_ProjectName)
		if self.err != nil {
			continue
		}

		projectName := make([]byte, m.Sz_ProjectName)
		self.err = binary.Read(buffer, binary.LittleEndian, projectName)
		if self.err != nil {
			continue
		}
		m.ProjectName = string(projectName)

		//append
		if self.err == nil {
			self.Metrics = append(self.Metrics, m)
		}
	}
	return self.err
}
