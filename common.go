package interop

import (
	"bufio"
	"encoding/binary"
	"os"
)

//LaneTile for common filter usage
type LaneTile struct {
	LaneNum uint16
	TileNum uint16
}

type HeaderInfo struct {
	Version uint8
	SSize   uint8
	Buf     *bufio.Reader
}

func GetHeader(file *os.File) (header *HeaderInfo, err error) {
	header = new(HeaderInfo)

	//	defer file.Close()

	header.Buf = bufio.NewReader(file)

	err = binary.Read(header.Buf, binary.LittleEndian, &header.Version)
	if err != nil {
		return
	}
	err = binary.Read(header.Buf, binary.LittleEndian, &header.SSize)
	if err != nil {
		return
	}
	return
}

func MakeLaneTileMap(tm *[]LaneTile) map[uint16]map[uint16]bool {
	ret := make(map[uint16]map[uint16]bool)
	if tm == nil {
		return ret
	}
	for _, lt := range *tm {
		if _, ok := ret[lt.LaneNum]; !ok {
			ret[lt.LaneNum] = make(map[uint16]bool)
		}
		ret[lt.LaneNum][lt.TileNum] = true
	}
	return ret
}
