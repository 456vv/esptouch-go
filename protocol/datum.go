package protocol

import (
	"github.com/456vv/esptouch-go/utils"
	"github.com/456vv/esptouch-go/utils/byteutil"
)

import (
	"container/list"
)

const (
	EXTRA_LEN      = 40
	EXTRA_HEAD_LEN = 5
)

type DatumCode struct {
	mDataCodes *list.List
}

func NewDatumCode(apSsid, apBssid, apPassword, ipAddress []byte) *DatumCode {
	totalXor := uint16(0)

	crc := utils.NewCRC8()
	apPwdLen := uint16(len(apPassword))
	crc.Update(apSsid, 0, len(apSsid))
	apSsidCrc := uint16(crc.GetValue())

	crc.Reset()
	apSsidLen := uint16(len(apSsid))
	crc.Update(apBssid, 0, len(apBssid))
	apBssidCrc := uint16(crc.GetValue())


	ipLen := len(ipAddress)

	totalLen := uint16(EXTRA_HEAD_LEN + uint16(ipLen) + apPwdLen + apSsidLen)

	//build data codes
	mDatacodes := list.New()
	mDatacodes.PushBack(NewDataCode(totalLen, 0))
	totalXor ^= totalLen
	mDatacodes.PushBack(NewDataCode(apPwdLen, 1))
	totalXor ^= apPwdLen
	mDatacodes.PushBack(NewDataCode(apSsidCrc, 2))
	totalXor ^= apSsidCrc
	mDatacodes.PushBack(NewDataCode(apBssidCrc, 3))
	totalXor ^= apBssidCrc
	// ESPDataCode 4 is null
	for i := (0); i < ipLen; i++ {
		c := byteutil.CovertByte2Uint8(ipAddress[i])
		totalXor ^= c
		mDatacodes.PushBack(NewDataCode(c, int(i+EXTRA_HEAD_LEN)))
	}
	for i := 0; i < len(apPassword); i++ {
		c := byteutil.CovertByte2Uint8(apPassword[i])
		totalXor ^= c
		mDatacodes.PushBack(NewDataCode(c, int(i+EXTRA_HEAD_LEN+int(ipLen))))
	}
	// totalXor will xor apSsidChars no matter whether the ssid is hidden
	for i := 0; i < len(apSsid); i++ {
		c := byteutil.CovertByte2Uint8(apSsid[i])
		totalXor ^= c
		mDatacodes.PushBack(NewDataCode(c, int(i+EXTRA_HEAD_LEN+int(ipLen)+int(apPwdLen))))
	}
	// add total xor last
	for element, i := mDatacodes.Front(), 0; element != nil; element, i = element.Next(), i+1 {
		if i == 4 {
			mDatacodes.InsertBefore(NewDataCode(totalXor, 4), element)
			break
		}
	}
	// add bssid
	bssidInsertIndex := EXTRA_HEAD_LEN
	for i := 0; i < len(apBssid); i++ {
		index := int(totalLen) + i
		c := byteutil.CovertByte2Uint8(apBssid[i])
		dc := NewDataCode(c, index)
		if bssidInsertIndex >= mDatacodes.Len() {
			mDatacodes.PushBack(dc)
		} else {
			for element, i := mDatacodes.Front(), 0; element != nil; element, i = element.Next(), i+1 {
				if i == bssidInsertIndex {
					mDatacodes.InsertAfter(dc, element)
					break
				}
			}
		}
		bssidInsertIndex += 4
	}
	return &DatumCode{mDataCodes: mDatacodes}
}

func (d *DatumCode) GetBytes() []byte {
	datumCode := make([]byte, d.mDataCodes.Len()*DATA_CODE_LEN)
	index := 0
	for element, i := d.mDataCodes.Front(), 0; element != nil; element, i = element.Next(), i+1 {
		if dc, ok := element.Value.(*DataCode); ok {
			for _, b := range dc.GetBytes() {
				datumCode[index] = b
				index++
			}
		}
	}
	return datumCode
}

func (d *DatumCode) GetU8s() []uint16 {
	dataBytes := d.GetBytes()
	bLen := len(dataBytes) / 2
	dataU8s := make([]uint16, bLen)
	var high, low byte
	for i := 0; i < bLen; i++ {
		high = dataBytes[i*2]
		low = dataBytes[i*2+1]
		dataU8s[i] = byteutil.Combine2bytesToU16(high, low) + EXTRA_LEN
	}
	return dataU8s
}
