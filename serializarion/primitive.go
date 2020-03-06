package serializarion

import "encoding/binary"

func DeserializeUint8(bts byte) uint8 {
	return uint8(bts)
}
func DeserializeUint16(bts []byte) uint16 {
	return binary.LittleEndian.Uint16(bts)
}
func DeserializeUint32(bts []byte) uint32 {
	return binary.LittleEndian.Uint32(bts)
}

func DeserializeUint64(bts []byte) uint64 {
	return binary.LittleEndian.Uint64(bts)
}

func SerializeUint8(bts uint8) byte {
	return byte(bts)
}
func SerializeUint16(value uint16) []byte {
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, value)
	return result
}
func SerializeUint32(value uint32) []byte {
	result := make([]byte, 4)
	binary.LittleEndian.PutUint32(result, value)
	return result
}

func SerializeUint64(value uint64) []byte {
	result := make([]byte, 8)
	binary.LittleEndian.PutUint64(result, value)
	return result
}
