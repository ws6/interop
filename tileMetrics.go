package interop

import (
	"encoding/binary"
	"os"
)

type TileMetrics struct {
	LaneNum     uint16
	TileNum     uint16
	MetricCode  uint16
	MetricValue float32
}

type TileInfo struct {
	Filename string
	Version  uint8
	SSize    uint8
	Metrics  []*TileMetrics
	err      error
}

func (self *TileInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.Filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	header, err := GetHeader(file)

	if err != nil {
		self.err = err
		return self.err
	}
	self.Version = header.Version
	self.SSize = header.SSize

	for {
		em := new(TileMetrics)
		err = binary.Read(header.Buf, binary.LittleEndian, em)
		if err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			break
		}

		self.Metrics = append(self.Metrics, em)
	}
	return self.err
}
