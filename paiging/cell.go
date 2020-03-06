package paiging

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"pager/serializarion"
	"pager/serializarion/gps_ad_id"
	"pager/serializarion/protos"
)

type Cell []byte

type CellComparator func(cell1 Cell, cell2 Cell) bool

func NewCell(payload []byte) (Cell, error) {
	cellSize := len(payload)
	if cellSize > 65535 {
		return nil, fmt.Errorf("too large payload")
	}
	cellSizeBytes := serializarion.SerializeUint16(uint16(cellSize))
	cellBytes := append(cellSizeBytes, payload...)
	return cellBytes, nil
}
func (c Cell) GetSize() uint16 {
	return c.GetPayloadSize() + 2
}
func (c Cell) GetPayloadSize() uint16 {
	return serializarion.DeserializeUint16(c[:2])
}
func (c Cell) GetPayload() []byte {
	return c[2:]
}

//TODO: change this and make it more generic
func GpsAdIdComparator(cell1 Cell, cell2 Cell) bool {
	userData1 := &protos.UserData{}
	userData2 := &protos.UserData{}
	err := proto.Unmarshal(cell1.GetPayload(), userData1)
	if err != nil {
		return false
	}
	err = proto.Unmarshal(cell2.GetPayload(), userData2)
	if err != nil {
		return false
	}
	gpsAdId1 := gps_ad_id.GpsAdId(userData1.GetGpsAdId())
	gpsAdId2 := gps_ad_id.GpsAdId(userData2.GetGpsAdId())
	return gpsAdId1.Gt(gpsAdId2)
}
