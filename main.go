package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type Record struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Generate random record data
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

// JSON serialization and calculate size
func jsonSerialize(records []Record) ([]byte, time.Duration) {
	start := time.Now()
	data, _ := json.Marshal(records)
	duration := time.Since(start)
	return data, duration
}

// JSON deserialization and timing
func jsonDeserialize(data []byte) (time.Duration, []Record) {
	var records []Record
	start := time.Now()
	_ = json.Unmarshal(data, &records)
	duration := time.Since(start)
	return duration, records
}

// Binlog serialization and calculate size
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

// Binlog deserialization and timing
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

// Create a line chart
func createLineChart(xValues []int, jsonValues, binlogValues []float64, title, yAxisName string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithYAxisOpts(opts.YAxis{Name: yAxisName}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Number of Records"}),
	)
	line.SetXAxis(xValues).
		AddSeries("JSON", generateLineItems(jsonValues)).
		AddSeries("Binlog", generateLineItems(binlogValues))
	return line
}

func generateLineItems(data []float64) []opts.LineData {
	items := make([]opts.LineData, 0)
	for _, v := range data {
		items = append(items, opts.LineData{Value: v})
	}
	return items
}

func main() {
	recordCounts := []int{10, 100, 1000, 10000, 20000, 30000, 40000, 50000, 60000, 70000}
	var jsonSerializeTimes, jsonDeserializeTimes []float64
	var binlogSerializeTimes, binlogDeserializeTimes []float64
	var jsonSizes, binlogSizes []float64

	for _, count := range recordCounts {
		records := generateRecords(count)

		var totalJsonSerializeTime, totalJsonDeserializeTime time.Duration
		var totalBinlogSerializeTime, totalBinlogDeserializeTime time.Duration
		var jsonSize, binlogSize int

		for i := 0; i < 10; i++ {
			// JSON serialization and deserialization
			jsonData, jsonSerializeTime := jsonSerialize(records)
			jsonDeserializeTime, _ := jsonDeserialize(jsonData)
			totalJsonSerializeTime += jsonSerializeTime
			totalJsonDeserializeTime += jsonDeserializeTime
			jsonSize = len(jsonData) // Record size only once

			// Binlog serialization and deserialization
			binlogData, binlogSerializeTime := binlogSerialize(records)
			binlogDeserializeTime, _ := binlogDeserialize(binlogData)
			totalBinlogSerializeTime += binlogSerializeTime
			totalBinlogDeserializeTime += binlogDeserializeTime
			binlogSize = len(binlogData) // Record size only once
		}

		// Calculate average and convert to seconds
		jsonSerializeTimes = append(jsonSerializeTimes, totalJsonSerializeTime.Seconds()/10)
		jsonDeserializeTimes = append(jsonDeserializeTimes, totalJsonDeserializeTime.Seconds()/10)
		binlogSerializeTimes = append(binlogSerializeTimes, totalBinlogSerializeTime.Seconds()/10)
		binlogDeserializeTimes = append(binlogDeserializeTimes, totalBinlogDeserializeTime.Seconds()/10)

		// Record file sizes (convert to KB)
		jsonSizes = append(jsonSizes, float64(jsonSize)/1024)
		binlogSizes = append(binlogSizes, float64(binlogSize)/1024)
	}

	// Create charts
	page := components.NewPage()
	page.AddCharts(
		createLineChart(recordCounts, jsonSerializeTimes, binlogSerializeTimes, "Average Serialization Time", "Time (seconds)"),
		createLineChart(recordCounts, jsonDeserializeTimes, binlogDeserializeTimes, "Average Deserialization Time", "Time (seconds)"),
		createLineChart(recordCounts, jsonSizes, binlogSizes, "File Size Comparison", "Size (KB)"),
	)

	// Use os.Create instead of open
	f, err := os.Create("serialization_deserialization_chart.html")

	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer f.Close()
	page.Render(f)

	// Get and print the file path
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}
	filePath := currentDir + "/serialization_deserialization_chart.html"
	fmt.Printf("Charts generated and saved as: %s\n", filePath)

	fmt.Println("Charts generated and saved as serialization_deserialization_chart.html")
}
