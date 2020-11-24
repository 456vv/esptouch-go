package esptouch

import (
	"errors"
	"github.com/456vv/esptouch-go/protocol"
	"github.com/456vv/esptouch-go/task"
	"github.com/456vv/esptouch-go/utils/byteutil"
	"net"
	"time"
	"context"
	"sync"
)

type EsptouchResult struct {
	BSSID       string
	IP 			net.IP
}

type EsptouchTask struct {
	parameter             	*task.EsptouchParameter
	apSsid                	[]byte
	apPassword            	[]byte
	apBssid               	[]byte
	udpClient             	*net.UDPConn
	mEsptouchResultList   	[]*EsptouchResult
	mBssidTaskSucCountMap 	map[string]int
	mIsInterrupt          	bool
	mIsExecuted           	bool
	mIsSuc                	bool
	wg						sync.WaitGroup
}

func NewEsptouchTask(apSsid, apPassword, apBssid []byte) (*EsptouchTask, error) {
	if apSsid == nil || len(apSsid) == 0 {
		return nil, errors.New("SSID can't be empty")
	}
	if apBssid == nil || len(apBssid) != 6 {
		return nil, errors.New("BSSID is empty or length is not 6")
	}
	if apPassword == nil {
		apPassword = []byte("")
	}
	mParameter := task.NewEsptouchParameter()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: mParameter.GetPortListening(),
	})
	
	if err != nil {
		return nil, err
	}
	return &EsptouchTask{
		parameter:             mParameter,
		apSsid:                apSsid,
		apPassword:            apPassword,
		apBssid:               apBssid,
		udpClient:             conn,
		mEsptouchResultList:   make([]*EsptouchResult, 0),
		mBssidTaskSucCountMap: make(map[string]int),
	}, nil
}

func (p *EsptouchTask) Close() error{
	p.interrupt()
	return p.udpClient.Close()
}

func (p *EsptouchTask) checkTaskValid() {
	if p.mIsExecuted {
		panic("the Esptouch task could be executed only once")
	}
	p.mIsExecuted = true
}
func (p *EsptouchTask) putEsptouchResult(bssid string, ip net.IP) {
	var count int
	if c, ok := p.mBssidTaskSucCountMap[bssid]; ok {
		count = c
	}
	count++
	p.mBssidTaskSucCountMap[bssid] = count
	if !(count >= p.parameter.GetThresholdSucBroadcastCount()) {
		return
	}
	var isExist = false
	for _, esptouchResultInList := range p.mEsptouchResultList {
		if esptouchResultInList.BSSID == bssid {
			isExist = true
			break
		}
	}
	if !isExist {
		p.mEsptouchResultList = append(p.mEsptouchResultList, &EsptouchResult{BSSID:bssid, IP:ip})
	}
}

func (p *EsptouchTask) listenAsync(expectDataLen int) {
	var startTime = time.Now()
	var expectOneByte = byte(len(p.apSsid) + len(p.apPassword) + 9)
	for {
		//1，中断
		//2，结果满足
		if p.mIsInterrupt || len(p.mEsptouchResultList) >= p.parameter.GetExpectTaskResultCount() {
			break
		}
		
		//超时是发送时间+接收时间，因为是同时进行的
		tout := startTime.Add(time.Duration(p.parameter.GetWaitUdpTotalMillisecond())*time.Millisecond)
		p.udpClient.SetReadDeadline(tout)
		var receiveBytes = make([]byte, expectDataLen)
		n, _, err := p.udpClient.ReadFromUDP(receiveBytes)
		if err != nil {
			//1，关闭连接
			//2，超时
			break
		}

		if n > 0 && receiveBytes[0] == expectOneByte {
			var bssid = byteutil.ParseBssid(receiveBytes, p.parameter.GetEsptouchResultOneLen(), p.parameter.GetEsptouchResultMacLen())
			var inetAddress = byteutil.ParseInetAddr(receiveBytes, p.parameter.GetEsptouchResultOneLen()+p.parameter.GetEsptouchResultMacLen(), p.parameter.GetEsptouchResultIpLen())
			p.putEsptouchResult(bssid, inetAddress)
		}
	}
	p.mIsSuc = len(p.mEsptouchResultList) >= p.parameter.GetExpectTaskResultCount()
	p.interrupt()
	p.wg.Done()
}

func (p *EsptouchTask) execute(generator *protocol.EsptouchGenerator) bool {
	gc := generator.GetGCBytes2()
	dc := generator.GetDCBytes2()

	startTime := time.Now().UnixNano() / 1e6
	currentTime := startTime
	for {
		if p.mIsInterrupt {
			break
		}
		//超出所有总超时时间
		if (currentTime - startTime) > p.parameter.GetWaitUdpSendingMillisecond() {
			break
		}
		
		//一直循环的发guidCode，直到2秒后超时
		for !p.mIsInterrupt && ((time.Now().UnixNano()/1e6)-currentTime) < p.parameter.GetTimeoutGuideCodeMillisecond() {
			//每隔8毫秒发送一组guideCode
			p.sendData(gc, 0, int64(len(gc)), p.parameter.GetIntervalGuideCodeMillisecond())
		}
		
		index := 0
		//一直循环的发dataCode，直到4秒后超时
		for !p.mIsInterrupt && ((time.Now().UnixNano()/1e6)-currentTime) < p.parameter.GetTimeoutDataCodeMillisecond() {
			//每隔8毫秒发送一组dataCode
			//每次发3组	
			p.sendData(dc, int64(index), 3, p.parameter.GetIntervalDataCodeMillisecond())
			//1,下一次从4开始
			index = (index + 3) % len(dc)
		}
		currentTime = time.Now().UnixNano() / 1e6
	}
	return p.mIsSuc
}

func (p *EsptouchTask) sendData(data [][]byte, offset, count int64, interval int64) {
	for i := offset; i < offset+count; i++ {
		if len(data[i]) == 0 {
			continue
		}
		p.udpClient.SetWriteDeadline(time.Time{})
		_, _ = p.udpClient.WriteToUDP(data[i], &net.UDPAddr{
			IP:   net.ParseIP(p.parameter.GetTargetHostname()),
			Port: p.parameter.GetTargetPort(),
		})
		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}

func (p *EsptouchTask) interrupt() {
	if !p.mIsInterrupt {
		//设置读写超时
		p.udpClient.SetDeadline(time.Unix(1, 0))
		p.mIsInterrupt = true
	}
}
func (p *EsptouchTask) Interrupt() {
	p.interrupt()
}

func (p *EsptouchTask) localIP() net.IP {
	netInterfaces, _ := net.Interfaces()
	for _, v := range netInterfaces {
		if (v.Flags & net.FlagUp) != 0 {
			addrs, _ := v.Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP
					}
				}
			}
		}
	}
	return nil
}

func (p *EsptouchTask) ExecuteForResultsCtx(ctx context.Context, expectTaskResultCount int) []*EsptouchResult {
	p.checkTaskValid()
	
	go func(){
		select {
		case <-ctx.Done():
			//1，退出发送
			//2，退出接收
			p.interrupt()
		}
	}()
	
	p.wg.Add(1)
	
	//设置等待接收配对设备数量
	p.parameter.SetExpectTaskResultCount(expectTaskResultCount)
	//同步监听接收设备返回的bssid+ip
	go p.listenAsync(p.parameter.GetEsptouchResultTotalLen())

	//生成长编码
	generator := protocol.NewEsptouchGenerator(p.apSsid, p.apBssid, p.apPassword, p.localIP())
	for i := 0; i < p.parameter.GetTotalRepeatItem(); i++ {
		//发送配对数据
		if p.execute(generator) {
			return p.mEsptouchResultList
		}
	}
	//等待接收超时
	p.wg.Wait()
	return p.mEsptouchResultList
}

func (p *EsptouchTask) ExecuteForResults(expectTaskResultCount int) []*EsptouchResult {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return p.ExecuteForResultsCtx(ctx, expectTaskResultCount)
}

func (p *EsptouchTask) SetPackageBroadcast(broadcast bool) {
	p.parameter.SetBroadcast(broadcast)
}
