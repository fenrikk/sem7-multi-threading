package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	student    = "Stukalov"
	raceRounds = 100000
)

var unsafeCounter int

func main() {
	fmt.Println("Студент:", student)
	fmt.Println("Практична робота 1, варіант 4")

	data := make([]int, 10)
	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range data {
		data[i] = source.Intn(26)
	}
	fmt.Println("Масив:", data)
	fmt.Println()

	var (
		wg      sync.WaitGroup
		greater []int
		mean    float64
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		local := make([]int, 0)
		for _, value := range data {
			if value > 10 {
				local = append(local, value)
			}
		}
		for i := 0; i < raceRounds; i++ {
			unsafeCounter++
		}
		greater = local
	}()
	go func() {
		defer wg.Done()
		sum := 0
		for _, value := range data {
			sum += value
		}
		mean = float64(sum) / float64(len(data))
		for i := 0; i < raceRounds; i++ {
			unsafeCounter++
		}
	}()
	wg.Wait()

	fmt.Println("T0, числа більші за 10:", greater)
	fmt.Printf("T1, середнє арифметичне: %.2f\n", mean)
	fmt.Printf("Спільний лічильник без mutex: очікувалось %d, отримано %d\n", raceRounds*2, unsafeCounter)

	checkSum := 0
	checkCount := 0
	for _, value := range data {
		checkSum += value
		if value > 10 {
			checkCount++
		}
	}
	checkMean := float64(checkSum) / float64(len(data))
	same := checkCount == len(greater) && checkMean == mean
	if same {
		fmt.Println("Звірка з головною горутиною: збігається")
	} else {
		fmt.Println("Звірка з головною горутиною: не збігається")
	}
}
