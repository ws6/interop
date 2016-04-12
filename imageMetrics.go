package interop

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

type ImageMetrics struct {
	LTC
	//	for version 2
	MinContrasts []uint16 // multipel channels
	MaxContrasts []uint16 // multipel channels

	//for version 1 sections
	ChannelId   uint16 //0: A; 1: C;2: G;3: T
	MinContrast uint16
	MaxContrast uint16
}

type ImageInfo struct {
	Filename      string
	Version       uint8
	SSize         uint8
	NumOfChannels uint8
	Metrics       []*ImageMetrics
	err           error
}

func NewImageMetrics(NumOfChannels uint8) *ImageMetrics {
	ret := new(ImageMetrics)
	ret.MinContrasts = make([]uint16, NumOfChannels)
	ret.MaxContrasts = make([]uint16, NumOfChannels)
	return ret
}

func (self *ImageInfo) Parse() error {
	if self.err != nil {
		return self.err
	}
	file, err := os.Open(self.Filename)
	if err != nil {
		self.err = err
		return self.err
	}
	defer file.Close()
	buffer := bufio.NewReader(file)

	//read version
	if err := binary.Read(buffer, binary.LittleEndian, &self.Version); err != nil {
		return err
	}

	//read length of each record
	if err := binary.Read(buffer, binary.LittleEndian, &self.SSize); err != nil {
		return err
	}

	if self.Version == uint8(1) {

		for {
			em := new(ImageMetrics)

			if err = binary.Read(buffer, binary.LittleEndian, &em.LTC); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				break
			}
			if err = binary.Read(buffer, binary.LittleEndian, &em.ChannelId); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				break
			}
			if err = binary.Read(buffer, binary.LittleEndian, &em.MinContrast); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				break
			}
			if err = binary.Read(buffer, binary.LittleEndian, &em.MaxContrast); err != nil {
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
	if self.Version == uint8(2) {
		//read length of each record
		if err := binary.Read(buffer, binary.LittleEndian, &self.NumOfChannels); err != nil {
			return err
		}

		for {
			m := NewImageMetrics(self.NumOfChannels)

			if err := binary.Read(buffer, binary.LittleEndian, &m.LTC); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}

				return fmt.Errorf("LTC err:%s", err)
			}

			if err := binary.Read(buffer, binary.LittleEndian, &m.MinContrasts); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				return fmt.Errorf("MinContrast err:%s", err)
			}

			if err := binary.Read(buffer, binary.LittleEndian, &m.MaxContrasts); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}
				return fmt.Errorf("MaxContrast err:%s", err)
			}

			self.Metrics = append(self.Metrics, m)
		}
		return self.err
	}

	return fmt.Errorf("unsuppored version %d", self.Version)

}
