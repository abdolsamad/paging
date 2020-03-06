package gps_ad_id

import "github.com/google/uuid"

type GpsAdId []byte

func NewFromString(gpsAdId string) (GpsAdId, error) {
	uuidObject, err := uuid.Parse(gpsAdId)
	if err != nil {
		return nil, err
	}
	return uuidObject.MarshalBinary()
}

func (gai GpsAdId) ToString() (string, error) {
	uuidObject, err := uuid.FromBytes(gai)
	if err != nil {
		return "", err
	}
	return uuidObject.String(), nil
}
func (gai GpsAdId) Lt(right GpsAdId) bool {
	for i := byte(0); i < 16; i++ {
		if gai[i] < right[i] {
			return true
		} else if gai[i] > right[i] {
			return false
		}
	}
	return false
}
func (gai GpsAdId) Gte(right GpsAdId) bool {
	return !gai.Lt(right)
}
func (gai GpsAdId) Gt(right GpsAdId) bool {
	for i := byte(0); i < 16; i++ {
		if gai[i] > right[i] {
			return true
		} else if gai[i] < right[i] {
			return false
		}
	}
	return false
}

func (gai GpsAdId) Lte(right GpsAdId) bool {
	return !gai.Gt(right)
}

func (gai GpsAdId) Equals(right GpsAdId) bool {
	for i := byte(0); i < 16; i++ {
		if gai[i] != right[i] {
			return false
		}
	}
	return true
}
