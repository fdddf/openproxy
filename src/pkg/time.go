package pkg

import "time"

func GetTimestamp() int64 {
	return time.Now().Unix()
}
