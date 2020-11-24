package main

import (
	"encoding/hex"
	"esptouch"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
	"context"
)

var (
	apSsid     	string
	apBssid    	string
	apPassword 	string
	mode       	bool
	num        	int
	tout		int
)

func init() {
	flag.StringVar(&apSsid, "ssid", "", "AP's SSID")
	flag.StringVar(&apBssid, "bssid", "", "AP's BSSID. such like 4C:50:77:73:37:B0")
	flag.StringVar(&apPassword, "psk", "", "AP's Password")
	flag.IntVar(&num, "num", 1, "Num of device to config")
	flag.BoolVar(&mode, "broadcast", false, "use broadcast mode?")
	flag.IntVar(&tout, "tout", 1*60, "timeout unit second")
	flag.Parse()
}

func main() {
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		return
	}
	
	bssidBytes := make([]byte, 0)
	bssidList := strings.Split(apBssid, ":")
	for _, v := range bssidList {
		bt, _ := hex.DecodeString(v)
		bssidBytes = append(bssidBytes, bt...)
	}
	task, err := esptouch.NewEsptouchTask([]byte(apSsid), []byte(apPassword), bssidBytes)
	if err != nil {
		panic(err)
	}
	defer task.Close()
	task.SetPackageBroadcast(mode)
	log.Println("SmartConfig run.")
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * time.Duration(tout))
	defer cancel()
	rList := task.ExecuteForResultsCtx(ctx, num)
	log.Println("Finished. totalCount:", len(rList))
	for _, v := range rList {
		fmt.Println(v)
	}
}
