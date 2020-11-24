package protocol

import (
	"esptouch/utils/byteutil"
	"net"
	"strconv"
	"strings"
)

type EsptouchGenerator struct {
	mGcBytes2 [][]byte
	mDcBytes2 [][]byte
}

func NewEsptouchGenerator(apSsid, apBssid, apPassword []byte, ipAddress net.IP) *EsptouchGenerator {
	//默认广播IP
	var ipBytes = []byte{255, 255, 255, 255}
	//使用指定IP替换默认广播IP
	if len(ipAddress) != 0 {
		for k, v := range strings.Split(ipAddress.String(), ".") {
			i, _ := strconv.Atoi(v)
			ipBytes[k] = byte(i)
		}
	}
	//组装guideCode
	//生成4个字节，值分别515-512
	//并填充他的长度全部是1
	gc := NewGuideCode()
	gcU81 := gc.GetU8s()
	mGcBytes2 := make([][]byte, len(gcU81))
	for i := 0; i < len(gcU81); i++ {
		mGcBytes2[i] = byteutil.GenSpecBytes(gcU81[i])
	}
	
	//组状dataCode
	dc := NewDatumCode(apSsid, apBssid, apPassword, ipBytes)
	dcU81 := dc.GetU8s()
	mDcBytes2 := make([][]byte, len(dcU81))
	for i := 0; i < len(dcU81); i++ {
		mDcBytes2[i] = byteutil.GenSpecBytes(dcU81[i])
	}
	return &EsptouchGenerator{
		mGcBytes2: mGcBytes2,
		mDcBytes2: mDcBytes2,
	}
}

func (e *EsptouchGenerator) GetGCBytes2() [][]byte {
	return e.mGcBytes2
}
func (e *EsptouchGenerator) GetDCBytes2() [][]byte {
	return e.mDcBytes2
}
