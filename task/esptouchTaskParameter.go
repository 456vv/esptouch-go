package task

import "fmt"

type EsptouchParameter struct {
	datagramCount                 int
	IntervalGuideCodeMillisecond int64
	IntervalDataCodeMillisecond  int64
	TimeoutGuideCodeMillisecond  int64
	TimeoutDataCodeMillisecond   int64
	TotalRepeatItem              int
	EsptouchResultOneLen         int
	EsptouchResultMacLen         int
	EsptouchResultIpLen          int
	EsptouchResultTotalLen       int
	PortListening                int
	TargetPort                   int
	WaitUdpReceivingMillisecond  int64
	WaitUdpSendingMillisecond    int64
	ThresholdSucBroadcastCount   int
	ExpectTaskResultCount        int
	Broadcast                    bool
}

func NewEsptouchParameter() *EsptouchParameter {
	return &EsptouchParameter{
		datagramCount:                 	0,
		IntervalGuideCodeMillisecond: 	8,
		IntervalDataCodeMillisecond:  	8,
		TimeoutGuideCodeMillisecond:  	2000,
		TimeoutDataCodeMillisecond:   	4000,
		TotalRepeatItem:              	1,
		EsptouchResultOneLen:         	1,
		EsptouchResultMacLen:         	6,
		EsptouchResultIpLen:          	4,
		EsptouchResultTotalLen:       	1 + 6 + 4,
		PortListening:                	18266,
		TargetPort:                   	7001,
		WaitUdpReceivingMillisecond:  	15000,
		WaitUdpSendingMillisecond:    	45000,
		ThresholdSucBroadcastCount:   	1,
		ExpectTaskResultCount:        	1,
		Broadcast:                    	true,
	}
}

func (p *EsptouchParameter) nextDatagramCount() int {
	p.datagramCount++
	return 1 + (p.datagramCount-1)%100
}
func (p *EsptouchParameter) GetTargetHostname() string {
	if p.Broadcast {
		//广播
		return "255.255.255.255"
	} else {
		//组播
		//这是乐鑫后面加上去的
		count := p.nextDatagramCount()
		return fmt.Sprintf("234.%d.%d.%d", count, count, count)
	}
}
