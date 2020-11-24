package byteutil

import (
	"fmt"
	"net"
	"strings"
)

func SplitUint8To2Bytes(char uint8) [2]byte {
	//0x12=>{0001, 0010}
	bt := [2]byte{0, 0}
	bt[0] = (char & 0xff) >> 4 //High
	bt[1] = char & 0x0f
	return bt
}

func Combine2BytesToOne(high, low byte) uint8 {
	//height==01,low==10
	//high<<4==010000,low==0010
	//00010000|00000010==00010010
	return high<<4 | low
}

func CovertByte2Uint8(b byte) uint16 {
	//0x1&0xff==0x01
	return uint16(b & 0xff)
}
func Combine2bytesToU16(h, l byte) uint16 {
	highU8 := uint16(CovertByte2Uint8(h))
	lowU8 := CovertByte2Uint8(l)
	return highU8<<8 | uint16(lowU8)
}

func GenSpecBytes(length uint16) []byte {
	bt := make([]byte, 0)
	for i := uint16(0); i < length; i++ {
		bt = append(bt, '1')
	}
	return bt
}

func ParseBssid(bssidBytes []byte, offset, count int) string {
	bytes := bssidBytes[offset : offset+count]
	bssid := fmt.Sprintf("% 02x", bytes)
	return strings.ReplaceAll(bssid, " ", ":")
}

func ParseInetAddr(inetAddrBytes []byte, offset, count int) net.IP {
	bytes := inetAddrBytes[offset : offset+count]
	var ip = make([]byte, count)
	copy(ip, bytes)
	return ip
}
