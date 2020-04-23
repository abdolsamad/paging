package test

import (
	"testing"

	"github.com/abdolsamad/paging/paging"
)

func TestUpdateFirstCellsOffsetAfterDelete(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})
	p.AddCell(cell0, comparator)
	offset1, _ := p.AddCell(cell1, comparator)
	offset2, _ := p.AddCell(cell2, comparator)
	p.DeleteCell(offset2)
	got := p.GetRawHeader().GetFirstCellOffset()
	expected := offset1
	if got != expected {
		t.Errorf("Expected first cell offset of %d but got %d instead", expected, got)
	}
}

func TestUpdateCellsCountAfterDelete(t *testing.T) {
	p := paging.NewPage()

	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})
	p.AddCell(cell0, comparator)
	p.AddCell(cell1, comparator)
	offset2, _ := p.AddCell(cell2, comparator)
	p.DeleteCell(offset2)

	expected := uint16(2)
	got := p.GetRawHeader().GetCellsCount()
	if got != expected {
		t.Errorf("Expected cell count of %d but got %d instead", expected, got)
	}
}

func TestDeleteUpdateFragmentsCount(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})

	offset0, _ := p.AddCell(cell0, comparator)
	p.AddCell(cell1, comparator)
	p.AddCell(cell2, comparator)

	p.DeleteCell(offset0)
	expected := uint8(0)
	got := p.GetRawHeader().GetFragmentedBytesCount()
	if got != expected {
		t.Errorf("Expected fragmented bytes count of %d but got %d instead", expected, got)
	}
	cell3, _ := paging.NewCell([]byte{1, 2})
	p.AddCell(cell3, comparator)
	expected = uint8(2)
	got = p.GetRawHeader().GetFragmentedBytesCount()
	if got != expected {
		t.Errorf("Expected fragmented bytes count of %d but got %d instead", expected, got)
	}
}

func TestDeleteUpdateFreeHeadersOffset(t *testing.T) {
	p := paging.NewPage()
	cell0, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell1, _ := paging.NewCell([]byte{1, 2, 3, 4})
	cell2, _ := paging.NewCell([]byte{1, 2, 3, 4})
	p.AddCell(cell0, comparator)
	offset1, _ := p.AddCell(cell1, comparator)
	p.AddCell(cell2, comparator)

	p.DeleteCell(offset1)
	expected := offset1
	got := p.GetRawHeader().GetFirstFreeBlockOffset()
	if got != expected {
		t.Errorf("Expected first free block offset of %d but got %d instead", expected, got)
	}
}
