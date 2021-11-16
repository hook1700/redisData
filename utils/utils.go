package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"redisData/pkg/logger"
	"strconv"
	"strings"
	"time"
)

//Strval 把输入类型转化成字符串类型
func Strval(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

// GetSleepTime 输入int64,转化成时间类型
func GetSleepTime(i int64) time.Duration {
	return time.Duration(i * 1)
}

// GetGenerateId 获取分布式ID
func GetGenerateId() string {
	s := uuid.NewV4().String()
	return s
}

// ParseGzip 解压Gzip数据
func ParseGzip(data []byte) ([]byte, error) {

	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, data)
	r, err := gzip.NewReader(b)
	if err != nil {
		fmt.Printf("[ParseGzip] NewReader error: %v, maybe data is ungzip", err)
		return nil, err
	} else {

		defer r.Close()
		undatas, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Printf("[ParseGzip]  ioutil.ReadAll error: %v", err)

			return nil, err
		}
		return undatas, nil
	}
}

// JSONToMap 把json数据妆花成map
func JSONToMap(str string) map[string]interface{} {

	var tempMap = make(map[string]interface{})

	err := json.Unmarshal([]byte(str), &tempMap)

	if err != nil {
		logger.Error(err)
	}

	return tempMap
}

// Split 分割字符串，以切片的方式返回
func Split(s, sep string) (result []string) {
	i := strings.Index(s, sep)

	for i > -1 {
		result = append(result, s[:i])
		s = s[i+len(sep):] // 这里使用len(sep)获取sep的长度
		i = strings.Index(s, sep)
	}
	result = append(result, s)
	return
}
