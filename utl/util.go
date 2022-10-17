package utl

import (
	"encoding/base64"
	"github.com/labstack/echo/v4"
	echoLog "github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"net/http"
	"runtime"
	"time"
	"unsafe"
)

func GetCurrentTimeStamp() int64 {
	return time.Now().Unix()
}

func TimeCost(logger *echoLog.Logger, fn string, req, resp interface{}) func() {
	t := time.Now()
	return func() {
		if err := recover(); err != nil {
			buffer := make([]byte, 1024*10)
			n := runtime.Stack(buffer, false)
			logger.Infof("%v, panic, err:%v, stack:%v", fn, err, string(buffer[:n]))
		}
		logger.Infof("%v, timeCost:%v, req:%v, resp:%v", fn, time.Since(t), req, resp)
	}
}

func PanicHandler(logger *echoLog.Logger, fn string) {
	if err := recover(); err != nil {
		buffer := make([]byte, 1024*10)
		n := runtime.Stack(buffer, false)
		logger.Infof("%v, panic, err:%v, stack:%v", fn, err, string(buffer[:n]))
	}
}

func WrapperErr(ctx echo.Context, err error) error {
	return ctx.String(http.StatusOK, err.Error())
}

func ToObjectId(s string) (primitive.ObjectID, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return primitive.ObjectIDFromHex(string(data))
}
func FromObjectId(id primitive.ObjectID) string {
	return base64.StdEncoding.EncodeToString([]byte(id.Hex()))
}

const (
	letters      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers      = "0123456789"
	lowerLetters = "abcdefghijklmnopqrstuvwxyz"
	upperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 6 bits to represent a letter index
	//letterIdBits = 6
	// All 1-bits as many as letterIdBits
	//letterIdMask = 1<<letterIdBits - 1
	//letterIdMax  = 63 / letterIdBits
)

type RandMode int

const (
	RandInt RandMode = iota
	RandString
	RandIntString
	RandLowerString
	RandUpperString
)

func GetRandString(mode RandMode, n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	bits, str := 0, ""
	switch mode {
	case RandInt:
		bits, str = 4, numbers
	case RandString:
		bits, str = 6, letters
	case RandIntString:
		bits, str = 7, numbers+letters
	case RandLowerString:
		bits, str = 4, lowerLetters
	case RandUpperString:
		bits, str = 4, upperLetters
	}
	mask, max, l := 1<<bits-1, 63/bits, len(str)
	for i, cache, remain := n-1, src.Int63(), max; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), max
		}
		b[i] = str[int(cache&int64(mask))%l]
		i--
		cache >>= bits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))

	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	//for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
	//	if remain == 0 {
	//		cache, remain = src.Int63(), letterIdMax
	//	}
	//	if idx := int(cache & letterIdMask); idx < len(letters) {
	//		b[i] = letters[idx]
	//		i--
	//	}
	//	cache >>= letterIdBits
	//	remain--
	//}
	//return *(*string)(unsafe.Pointer(&b))
}
