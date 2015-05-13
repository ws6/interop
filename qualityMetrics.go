package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

//common struct Lane,Tile and Cycle
type LTC struct {
	LaneNum uint16
	TileNum uint16
	Cycle   uint16
}

type QMetrics struct {
	LTC
	NumClusters [50]uint32 //first 50 clusters by cycle score Q1 through Q50 (uint32
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
	NumQscores uint8 //number of quality score bins, B
	QbinConfig QbinConfig
	Metrics    []*QMetrics
	err        error
}

func (self *QMetricsInfo) Error() string {
	if self.err != nil {
		if self.err.Error() == "EOF" {
			return ""
		}
		return self.err.Error()
	}
	return ""
}

func (self *QMetricsInfo) ParseQbinConfig(buffer *bufio.Reader) error {
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
	return nil
}

func (self *QMetricsInfo) ParseQbin(buffer *bufio.Reader) error {
	if err := self.ParseQbinConfig(buffer); err != nil {
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

func boundCheck(arr []uint8, errTag string) error {
	for i, v := range arr {
		if v >= 50 {
			return fmt.Errorf("%s: %d >=50 at %d", errTag, v, i)
		}
	}
	return nil
}

func (self *QMetricsInfo) ValidateQbinConfig() error {
	if self.NumQscores > 50 {
		return fmt.Errorf("qbin szie:%d >50", self.NumQscores)
	}

	if err := boundCheck(self.QbinConfig.LowerBound, "lower bound"); err != nil {
		return err
	}
	if err := boundCheck(self.QbinConfig.UpperBound, "upper bound"); err != nil {
		return err
	}
	if err := boundCheck(self.QbinConfig.ReMapScores, "remap score"); err != nil {
		return err
	}
	return nil
}

func (self *QMetricsInfo) ParseVersion6(buffer *bufio.Reader) error {
	if self.err = self.ParseQbinConfig(buffer); self.err != nil {
		return self.err
	}
	if self.EnableQbin {
		if self.err = self.ValidateQbinConfig(); self.err != nil {
			return self.err
		}
	}

	ok := true
	n := uint32(0)
	for {
		if self.err != nil {
			if self.err.Error() == "EOF" {
				self.err = nil
			}
			break
		}
		m := new(QMetrics)
		if self.EnableQbin {
			if self.err = binary.Read(buffer, binary.LittleEndian, &m.LTC); self.err != nil {
				continue
			}
			ok = true
			for i := uint8(0); i < self.NumQscores; i++ {
				n = 0
				if self.err = binary.Read(buffer, binary.LittleEndian, &n); self.err != nil {
					ok = false
					break
				}
				m.NumClusters[self.QbinConfig.ReMapScores[i]] = n
			}
			if !ok {
				continue
			}

		} else {
			if self.err = binary.Read(buffer, binary.LittleEndian, m); self.err != nil {
				continue
			}
		}

		if self.err == nil {
			self.Metrics = append(self.Metrics, m)
		}

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

	if self.Version == 4 {
		return self.ParseNonQbin(header.Buf)
	}

	if self.Version == 5 {
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
	if self.Version == 6 {
		var enableQbined uint8
		if self.err = binary.Read(header.Buf, binary.LittleEndian, &enableQbined); self.err != nil {
			return self.err
		}

		if enableQbined == 1 {
			self.EnableQbin = true
			return self.ParseVersion6(header.Buf)
		}
	}
	return self.ParseNonQbin(header.Buf)
}
