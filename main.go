package main

import (
	"fmt"
	"pager/paging"
	"pager/serializarion/protos"
	"time"

	"google.golang.org/protobuf/proto"
)

type PageFlag byte

const (
	PAGE_FLAG_LEAF = PageFlag(8)
)

func UserDataToCell(userData *protos.UserData) (paging.Cell, error) {
	bytesU1, err := proto.Marshal(userData)
	if err != nil {
		return nil, err
	}
	cell, err := paging.NewCell(bytesU1)
	if err != nil {
		return nil, err
	}
	return cell, nil
}

func CellToUserData(cell paging.Cell) (*protos.UserData, error) {
	ud := &protos.UserData{}
	err := proto.Unmarshal(cell[2:], ud)
	if err != nil {
		return nil, err
	}
	return ud, nil
}

func AddUserData(page paging.Page, userData *protos.UserData) error {
	cell, err := UserDataToCell(userData)
	if err != nil {
		return err
	}

	_, err = page.AddCell(cell, paging.GpsAdIdComparator)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	start := time.Now()
	p := paging.NewPage()
	c1, _ := paging.NewCell([]byte{1, 1, 1, 1, 1, 1, 1, 1})
	c2, _ := paging.NewCell([]byte{2, 2, 2, 2, 2, 2, 2, 2})
	c3, _ := paging.NewCell([]byte{3, 3, 3, 3, 3, 3, 3, 3})
	c4, _ := paging.NewCell([]byte{4, 4, 4, 4, 4, 4, 4, 4})
	cellComparator := func(cell1 paging.Cell, cell2 paging.Cell) bool {
		return cell1.GetPayload()[0] < cell2.GetPayload()[0]
	}
	offset1, _ := p.AddCell(c1, cellComparator)
	offset2, _ := p.AddCell(c2, cellComparator)
	offset3, _ := p.AddCell(c3, cellComparator)
	p.DeleteCell(offset2)
	offset4, _ := p.AddCell(c4, cellComparator)
	fmt.Printf("The whole thing took: %d us\n", time.Since(start).Microseconds())
	fmt.Printf("Cell1 added in %d\n", offset1)
	fmt.Printf("Cell2 added in %d\n", offset2)
	fmt.Printf("Cell3 added in %d\n", offset3)
	fmt.Printf("Cell4 added in %d\n", offset4)

	fmt.Println(p[8100:8192])
}
