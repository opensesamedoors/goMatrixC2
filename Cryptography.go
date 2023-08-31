package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func createRandomWord() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	var sb strings.Builder
	for i := 0; i < 4; i++ {
		sb.WriteByte(alphabet[rand.Intn(len(alphabet))])
	}

	return sb.String()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetRandomString(n int) string {
	str := "abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	for i := 0; i < n; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}

// Todo: refactor and fix polling duration issues
func sleepWithBitterness(seconds string, bitterness string) {
	sleepDuration, err := time.ParseDuration(seconds + "s")
	if err != nil {
		fmt.Println("Error parsing sleep duration:", err)
		return
	}

	bitternessPercent, err := strconv.Atoi(bitterness)
	if err != nil {
		fmt.Println("Error parsing bitterness percent:", err)
		return
	}

	rand.Seed(time.Now().UnixNano())
	jitter := time.Duration(rand.Intn(bitternessPercent)) * time.Millisecond
	sleepDuration += jitter
}
