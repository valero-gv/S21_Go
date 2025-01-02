package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
)

func mean(nums []int) float64 {
	var sum int
	for _, num := range nums {
		sum += num
	}
	return float64(sum) / float64(len(nums))
}

func median(numbers []int) float64 {
	n := len(numbers)

	if n%2 == 0 {
		return float64(numbers[n/2-1]+numbers[n/2]) / 2.0
	}

	return float64(numbers[n/2])
}

func mode(numbers []int) int {
	freq := make(map[int]int)
	var mode int
	maxFreq := 0
	for _, num := range numbers {
		freq[num] += 1
		if freq[num] > maxFreq {
			mode = num
			maxFreq = freq[num]
		} else if freq[num] == maxFreq && num < mode {
			mode = num
		}
	}
	return mode
}

func standardDeviation(numbers []int) float64 {
	m := mean(numbers)
	var sumSqDiff float64
	var diff float64
	for _, num := range numbers {
		diff = float64(num) - m
		sumSqDiff += diff * diff
	}

	meanSqDiff := sumSqDiff / float64(len(numbers))
	return math.Sqrt(meanSqDiff)
}

func main() {
	meanFlag := flag.Bool("mean", false, "Calculate mean")
	medianFlag := flag.Bool("median", false, "Calculate median")
	modeFlag := flag.Bool("mode", false, "Calculate mode")
	sdFlag := flag.Bool("sd", false, "Calculate standard deviation")

	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	var numbers []int

	for scanner.Scan() {
		num, err := strconv.Atoi(scanner.Text())
		if err != nil {
			fmt.Println("Error converting input to integer:", err)
			continue
		}
		numbers = append(numbers, num)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	slices.Sort(numbers)

	anyFlagsSet := *meanFlag || *medianFlag || *modeFlag || *sdFlag

	if !anyFlagsSet || *meanFlag {
		fmt.Printf("Mean: %.2f\n", mean(numbers))
	}
	if !anyFlagsSet || *medianFlag {
		fmt.Printf("Median: %.2f\n", median(numbers))
	}
	if !anyFlagsSet || *modeFlag {
		fmt.Printf("Mode: %d\n", mode(numbers))
	}
	if !anyFlagsSet || *sdFlag {
		fmt.Printf("SD: %.2f\n", standardDeviation(numbers))
	}
}
