package interop

//RegistrationMetricsOut.bin
import (
	"bufio"
	"encoding/binary"
	"fmt"
	//	"math"
	"os"
)

type SubtileOffsetRegion struct {
	PixelShiftInX     float32
	PixelShiftInY     float32
	RegistrationScore float32
}

type AffineMetrics struct {
	TranslationX   float32
	TranslationY   float32
	MagnificationX float32
	MagnificationY float32
	ShearXY        float32 //the offset in Y as a function of X (float)
	ShearYX        float32 //the offset in X as a function of Y (float)
}

type ChannelMetrics struct {
	Regions []SubtileOffsetRegion
	AffineMetrics
}

type RegistrationSubTileMetrics struct {
	LTC
	Channels []ChannelMetrics
}
type RegistrationMetricsInfo struct {
	Filename           string
	Version            uint8
	SSize              uint16
	NumOfChannels      uint8
	NumberOfSubRegions uint8
	Metrics            []*RegistrationSubTileMetrics
	err                error
}

func NewMetrics(NumOfChannels, NumberOfSubRegions int) *RegistrationSubTileMetrics {
	ret := new(RegistrationSubTileMetrics)
	ret.Channels = make([]ChannelMetrics, NumOfChannels)
	for i, _ := range ret.Channels {
		ret.Channels[i].Regions = make([]SubtileOffsetRegion, NumberOfSubRegions)
	}
	return ret
}

func (self *RegistrationMetricsInfo) Parse() error {
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

	//read  NumOfChannels
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumOfChannels); err != nil {
		return err
	}
	//read  NumberOfSubRegions
	if err := binary.Read(buffer, binary.LittleEndian, &self.NumberOfSubRegions); err != nil {
		return err
	}

	for {

		m := NewMetrics(int(self.NumOfChannels), int(self.NumberOfSubRegions))
		//read LTC
		if err := binary.Read(buffer, binary.LittleEndian, &m.LTC); err != nil {
			self.err = err
			if err.Error() == "EOF" {
				return nil
			}

			return fmt.Errorf("LTC err:%s", err)
		}
		for i, _ := range m.Channels {

			//Read Subtile offset
			for j, _ := range m.Channels[i].Regions {
				if err := binary.Read(buffer, binary.LittleEndian, &m.Channels[i].Regions[j]); err != nil {
					self.err = err
					if err.Error() == "EOF" {
						return nil
					}

					return fmt.Errorf("Regions err:%s", err)
				}
			}

			//Read Affine

			if err := binary.Read(buffer, binary.LittleEndian, &m.Channels[i].AffineMetrics); err != nil {
				self.err = err
				if err.Error() == "EOF" {
					return nil
				}

				return fmt.Errorf("AffineMetrics err:%s", err)
			}
		}

		self.Metrics = append(self.Metrics, m)
	}
	return err
}
