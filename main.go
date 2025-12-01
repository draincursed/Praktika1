package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL      = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval   = 10 * time.Second
	errorThreshold = 3
)

func main() {
	errorCount := 0
	
	for {
		stats, err := fetchStats()
		if err != nil {
			errorCount++
			if errorCount >= errorThreshold {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(pollInterval)
			continue
		}
		
		errorCount = 0
		checkMetrics(stats)
		time.Sleep(pollInterval)
	}
}

func fetchStats() ([]int64, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	
	scanner := bufio.NewScanner(resp.Body)
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty response")
	}
	
	line := strings.TrimSpace(scanner.Text())
	values := strings.Split(line, ",")
	
	if len(values) < 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(values))
	}
	
	stats := make([]int64, 7)
	for i := 0; i < 7; i++ {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(values[i]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid value '%s': %v", values[i], err)
		}
		stats[i] = int64(parsed)
	}
	
	return stats, nil
}

func checkMetrics(stats []int64) {
	// Load Average - целочисленное сравнение
	if stats[0] > 30 {
		fmt.Printf("Load Average is too high: %d\n", stats[0])
	}
	
	// Memory - используем прямое сравнение float процентов
	if stats[1] > 0 {
		memoryUsagePercent := (float64(stats[2]) / float64(stats[1])) * 100
		// Прямое сравнение float с порогом
		if memoryUsagePercent > 80.0 {
			fmt.Printf("Memory usage too high: %.0f%%\n", memoryUsagePercent)
		}
	}
	
	// Disk - используем прямое сравнение float процентов
	if stats[3] > 0 {
		diskUsagePercent := (float64(stats[4]) / float64(stats[3])) * 100
		// Прямое сравнение float с порогом
		if diskUsagePercent > 90.0 {
			freeSpaceMB := (stats[3] - stats[4]) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeSpaceMB)
		}
	}
	
	// Network - используем прямое сравнение float процентов
	if stats[5] > 0 {
		networkUsagePercent := (float64(stats[6]) / float64(stats[5])) * 100
		// Прямое сравнение float с порогом
		if networkUsagePercent > 90.0 {
			availableBandwidthBytes := stats[5] - stats[6]
			availableBandwidthMbits := availableBandwidthBytes / 1000000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", availableBandwidthMbits)
		}
	}
}
