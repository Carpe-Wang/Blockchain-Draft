package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type Record struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// 生成随机记录数据
func generateRecords(count int) []Record {
	records := make([]Record, count)
	for i := 0; i < count; i++ {
		records[i] = Record{
			ID:    i,
			Name:  fmt.Sprintf("name_%d", i),
			Value: fmt.Sprintf("value_%d", rand.Intn(1000)),
		}
	}
	return records
}

// JSON序列化并计算大小
func jsonSerialize(records []Record) ([]byte, time.Duration) {
	start := time.Now()
	data, _ := json.Marshal(records)
	duration := time.Since(start)
	return data, duration
}

// JSON反序列化并计时
func jsonDeserialize(data []byte) (time.Duration, []Record) {
	var records []Record
	start := time.Now()
	_ = json.Unmarshal(data, &records)
	duration := time.Since(start)
	return duration, records
}

// Binlog序列化并计算大小
func binlogSerialize(records []Record) ([]byte, time.Duration) {
	var buffer bytes.Buffer
	start := time.Now()
	for _, record := range records {
		binary.Write(&buffer, binary.LittleEndian, int32(record.ID))
		binary.Write(&buffer, binary.LittleEndian, int32(len(record.Name)))
		buffer.WriteString(record.Name)
		binary.Write(&buffer, binary.LittleEndian, int32(len(record.Value)))
		buffer.WriteString(record.Value)
	}
	duration := time.Since(start)
	return buffer.Bytes(), duration
}

// Binlog反序列化并计时
func binlogDeserialize(data []byte) (time.Duration, []Record) {
	buffer := bytes.NewBuffer(data)
	var records []Record
	start := time.Now()
	for buffer.Len() > 0 {
		var id int32
		var nameLen, valueLen int32
		binary.Read(buffer, binary.LittleEndian, &id)
		binary.Read(buffer, binary.LittleEndian, &nameLen)
		name := string(buffer.Next(int(nameLen)))
		binary.Read(buffer, binary.LittleEndian, &valueLen)
		value := string(buffer.Next(int(valueLen)))
		records = append(records, Record{ID: int(id), Name: name, Value: value})
	}
	duration := time.Since(start)
	return duration, records
}

func main() {
	for _, count := range []int{10, 100, 1000, 10000} {
		fmt.Printf("\n=== %d records ===\n", count)
		records := generateRecords(count)

		// JSON 序列化与反序列化
		jsonData, jsonSerializeTime := jsonSerialize(records)
		jsonDeserializeTime, _ := jsonDeserialize(jsonData)
		fmt.Printf("JSON - Size: %d bytes, Serialize Time: %v, Deserialize Time: %v\n",
			len(jsonData), jsonSerializeTime, jsonDeserializeTime)

		// Binlog 序列化与反序列化
		binlogData, binlogSerializeTime := binlogSerialize(records)
		binlogDeserializeTime, _ := binlogDeserialize(binlogData)
		fmt.Printf("Binlog - Size: %d bytes, Serialize Time: %v, Deserialize Time: %v\n",
			len(binlogData), binlogSerializeTime, binlogDeserializeTime)
	}
}
