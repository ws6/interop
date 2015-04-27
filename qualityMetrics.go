package interop

import (
	"bufio"
	"encoding/binary"
	"os"
)

//withou qbinning <= version 4
type QMetrics struct {
	LaneNum     uint16
	TileNum     uint16
	Cycle       uint16
	NumClusters [50]uint32
}
type QbinConfig struct {
	LowerBound  []uint8
	UpperBound  []uint8
	ReMapScores []uint8
}

type QMetricsInfo struct {
	Filename   string
	Version    uint8
	SSize      uint8
	EnableQbin bool
	NumQscores uint8
	QbinConfig QbinConfig
	Metrics    []*QMetrics
	err        error
}

func (self *QMetricsInfo) ParseQbin(buffer *bufio.Reader) error {
	err := binary.Read(buffer, binary.LittleEndian, &self.NumQscores)
	if err != nil {
		self.err = err
		return err
	}
	self.QbinConfig.LowerBound = make([]uint8, self.NumQscores)
	err = binary.Read(buffer, binary.LittleEndian, &self.QbinConfig.LowerBound)
	if err != nil {
		self.err = err
		return err
	}
	self.QbinConfig.UpperBound = make([]uint8, self.NumQscores)
	err = binary.Read(buffer, binary.LittleEndian, &self.QbinConfig.UpperBound)
	if err != nil {
		self.err = err
		return err
	}

	self.QbinConfig.ReMapScores = make([]uint8, self.NumQscores)
	err = binary.Read(buffer, binary.LittleEndian, &self.QbinConfig.ReMapScores)
	if err != nil {
		self.err = err
		return err
	}
	return self.ParseNonQbin(buffer)
}
func (self *QMetricsInfo) ParseNonQbin(buffer *bufio.Reader) error {
	if self.err != nil {
		return self.err
	}
	for {
		m := new(QMetrics)
		err := binary.Read(buffer, binary.LittleEndian, m)
		if err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}
			break
		}
		self.Metrics = append(self.Metrics, m)
	}
	return self.err
}

func (self *QMetricsInfo) Parse() error {
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
	self.EnableQbin = false
	if self.Version > 4 {
		//		self.EnableQbin = true
		var enableQbined uint8
		err := binary.Read(header.Buf, binary.LittleEndian, &enableQbined)
		if err != nil {
			self.err = err
			return self.err
		}
		if enableQbined == 1 {
			self.EnableQbin = true
			return self.ParseQbin(header.Buf)
		}
	}
	return self.ParseNonQbin(header.Buf)
}
