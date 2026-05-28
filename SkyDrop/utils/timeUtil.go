package utils

import (
	"fmt"
	"strconv"
	"time"
)

func GetCurrentUnixTime(strTime string) int64 {

	LOC, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		LOC = time.FixedZone("CST-8", 8*3600)
	}
	strDate := time.Now().Format("2006-01-02")
	strDateTime := fmt.Sprintf("%s %s:00", strDate, strTime)
	//fmt.Println(strDateTime)
	//要转换成时间日期的格式模板（go诞生时间，模板必须是这个时间）
	timeTemplate := "2006-01-02 15:04:05"
	unixTime, _ := time.ParseInLocation(timeTemplate, strDateTime, LOC)
	return unixTime.Unix()
}

func GetMillisecond() string {
	ts := time.Now().UnixNano() / 1e6
	return strconv.FormatInt(ts, 10)
}