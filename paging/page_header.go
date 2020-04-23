package paging

import "github.com/abdolsamad/paging/serializarion"

type PageFlag byte

const (
	PAGE_FLAG_INT_KEY   = PageFlag(1)
	PAGE_FLAG_ZERO_DATA = PageFlag(2)
	PAGE_FLAG_LEAF_DATA = PageFlag(4)
	PAGE_FLAG_LEAF      = PageFlag(8)

	PAGE_SIZE_UNIT   = 4 * 1024
	PAGE_SIZE        = 2 * PAGE_SIZE_UNIT //Page size in bytes
	PAGE_HEADER_SIZE = 8                  //Page header size in bytes
)

type RawPageHeader []byte

func NewRawPageHeader(flag PageFlag) RawPageHeader {
	pageHeader := RawPageHeader(make([]byte, PAGE_HEADER_SIZE))
	pageHeader.SetFlag(flag)
	pageHeader.SetCellsCount(0)
	pageHeader.SetFirstCellOffset(InvalidOffset)
	pageHeader.SetFirstFreeBlockOffset(0)
	pageHeader.SetFragmentedBytesCount(0)
	return pageHeader
}

type PageHeader struct { // 8 byte
	Flags                PageFlag // 1 byte bytes: 0
	CellsCount           uint16   // 2 byte bytes: [1,2]
	FirstCellOffset      uint16   // 2 byte bytes [3,4]
	FirstFreeBlockOffset uint16   // 2 byte bytes [5-6]
	FragmentedBytesCount uint8    // 1 byte bytes 7
}

func (rph RawPageHeader) GetFlag() PageFlag {
	return PageFlag(rph[0])
}

func (rph RawPageHeader) GetCellsCount() uint16 {
	return serializarion.DeserializeUint16(rph[1:3])
}

func (rph RawPageHeader) GetFirstCellOffset() uint16 {
	return serializarion.DeserializeUint16(rph[3:5])
}

func (rph RawPageHeader) GetFirstFreeBlockOffset() uint16 {
	return serializarion.DeserializeUint16(rph[5:7])
}

func (rph RawPageHeader) GetFragmentedBytesCount() uint8 {
	return serializarion.DeserializeUint8(rph[7])
}

func (rph RawPageHeader) SetFlag(pageFlag PageFlag) {
	rph[0] = byte(pageFlag)
}

func (rph RawPageHeader) SetCellsCount(cellsCount uint16) {
	bts := serializarion.SerializeUint16(cellsCount)
	rph[1] = bts[0]
	rph[2] = bts[1]
}

func (rph RawPageHeader) IncreaseFragmentedBytesCount(size uint8) {
	//TODO: defragment if necessary
	fragmentedBytesCount := rph.GetFragmentedBytesCount()
	rph.SetFragmentedBytesCount(fragmentedBytesCount + size)
}

func (rph RawPageHeader) DecreaseFragmentedBytesCount(size uint8) {
	fragmentedBytesCount := rph.GetFragmentedBytesCount()
	rph.SetFragmentedBytesCount(fragmentedBytesCount - size)
}

func (rph RawPageHeader) IncreaseCellsCount() {
	cellsCount := rph.GetCellsCount()
	rph.SetCellsCount(cellsCount + 1)
}
func (rph RawPageHeader) DecreaseCellsCount() {
	cellsCount := rph.GetCellsCount()
	rph.SetCellsCount(cellsCount - 1)
}

func (rph RawPageHeader) SetFirstCellOffset(firstCellOffset uint16) {
	bts := serializarion.SerializeUint16(firstCellOffset)
	rph[3] = bts[0]
	rph[4] = bts[1]
}

func (rph RawPageHeader) SetFirstFreeBlockOffset(firstFreeBlockOffset uint16) {
	bts := serializarion.SerializeUint16(firstFreeBlockOffset)
	rph[5] = bts[0]
	rph[6] = bts[1]
}

func (rph RawPageHeader) SetFragmentedBytesCount(fragmentedBytesCount uint8) {
	rph[7] = fragmentedBytesCount
}
