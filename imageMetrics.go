package interop

import (
	"encoding/binary"
	"os"
)

type ImageMetrics struct {
	LaneNum     uint16
	TileNum     uint16
	Cycle       uint16
	ChannelId   uint16 //0: A; 1: C;2: G;3: T
	MinContrast uint16
	MaxContrast uint16
}

type ImageInfo struct {
	filename string
	Version  uint8
	SSize    uint8
	Metrics  []*ImageMetrics
	err      error
}

func (self *ImageInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.filename)
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
		em := new(ImageMetrics)
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
