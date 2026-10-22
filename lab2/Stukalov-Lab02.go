package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	student      = "Stukalov"
	goroutines   = 4
	perGoroutine = 200000
	valueRange   = 32
	shardCount   = 8
)

func value(id, i int) int {
	return (id*131 + i*17) % valueRange
}

func expectedTotal() int {
	return goroutines * perGoroutine
}

func runUnsafe() {
	fmt.Println("A. Небезпечна версія: спільна map без sync.Mutex")
	fmt.Printf("Потоків: %d, подій на потік: %d\n", goroutines, perGoroutine)
	_ = os.Stdout.Sync()

	freq := make(map[int]int, valueRange)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for id := 0; id < goroutines; id++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				freq[value(id, i)]++
			}
		}(id)
	}
	wg.Wait()
	fmt.Println("Сума частот:", sumMap(freq))
}

func runLockedEachTime() (int, time.Duration) {
	freq := make(map[int]int, valueRange)
	var mu sync.Mutex
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(goroutines)
	for id := 0; id < goroutines; id++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				key := value(id, i)
				mu.Lock()
				freq[key]++
				mu.Unlock()
			}
		}(id)
	}
	wg.Wait()
	return sumMap(freq), time.Since(start)
}

func runLocalBatch() (int, time.Duration) {
	freq := make(map[int]int, valueRange)
	var mu sync.Mutex
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(goroutines)
	for id := 0; id < goroutines; id++ {
		go func(id int) {
			defer wg.Done()
			local := make(map[int]int, valueRange)
			for i := 0; i < perGoroutine; i++ {
				local[value(id, i)]++
			}
			mu.Lock()
			for key, count := range local {
				freq[key] += count
			}
			mu.Unlock()
		}(id)
	}
	wg.Wait()
	return sumMap(freq), time.Since(start)
}

type shard struct {
	mu   sync.Mutex
	freq map[int]int
}

func runSharded() (int, time.Duration) {
	shards := make([]shard, shardCount)
	for i := range shards {
		shards[i].freq = make(map[int]int)
	}
	var wg sync.WaitGroup
	start := time.Now()
	wg.Add(goroutines)
	for id := 0; id < goroutines; id++ {
		go func(id int) {
			defer wg.Done()
			local := make(map[int]int, valueRange)
			for i := 0; i < perGoroutine; i++ {
				local[value(id, i)]++
			}
			grouped := make([][][2]int, shardCount)
			for key, count := range local {
				index := key % shardCount
				grouped[index] = append(grouped[index], [2]int{key, count})
			}
			for index, pairs := range grouped {
				if len(pairs) == 0 {
					continue
				}
				shards[index].mu.Lock()
				for _, pair := range pairs {
					shards[index].freq[pair[0]] += pair[1]
				}
				shards[index].mu.Unlock()
			}
		}(id)
	}
	wg.Wait()
	total := 0
	for i := range shards {
		total += sumMap(shards[i].freq)
	}
	return total, time.Since(start)
}

func sumMap(freq map[int]int) int {
	total := 0
	for _, count := range freq {
		total += count
	}
	return total
}

func printResult(title string, total int, elapsed time.Duration) {
	status := "збігається"
	if total != expectedTotal() {
		status = "не збігається"
	}
	fmt.Printf("%s\n  сума частот: %d (%s), час: %s\n", title, total, status, elapsed)
}

func runSafe() {
	fmt.Println("B. Безпечна версія: sync.Mutex і короткі критичні секції")
	fmt.Printf("Потоків: %d, подій на потік: %d, значень: 0..%d\n", goroutines, perGoroutine, valueRange-1)
	fmt.Println("Очікувана сума частот:", expectedTotal())
	fmt.Println()

	total, elapsed := runLockedEachTime()
	printResult("1) Один mutex на кожен інкремент", total, elapsed)

	total, elapsed = runLocalBatch()
	printResult("2) Локальна map + одне злиття під mutex", total, elapsed)

	total, elapsed = runSharded()
	printResult("3) Локальна map + mutex на шард", total, elapsed)
}

func main() {
	fmt.Println("Студент:", student)
	fmt.Println("Практична робота 2, варіант 4. Частотний словник")
	fmt.Println()

	mode := "safe"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	switch mode {
	case "unsafe":
		runUnsafe()
	case "safe":
		runSafe()
	default:
		fmt.Println("Режим: safe або unsafe")
		os.Exit(2)
	}
}
