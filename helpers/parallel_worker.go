package H

import (
	"fmt"
	"runtime"
	"sync"
)

// runJob processes a single item from the data slice using the provided processFunc.
// It updates the results slice with the processed result in a thread-safe manner.
//
// Parameters:
//
//	mu - A mutex used to ensure exclusive access to the shared last index.
//	last - A pointer to an integer representing the last processed index.
//	data - A pointer to the slice of input data to be processed.
//	results - A pointer to the slice where processed results are stored.
//	processFunc - A function that takes an element of type T and returns a result of type R.
func runJob[T any, R any](mu *sync.Mutex, last *int, data *[]T, results *[]R, processFunc func(T) R) {
	if data == nil || last == nil || results == nil || mu == nil {
		return
	}

	mu.Lock()
	current := *last
	current++
	*last = current
	mu.Unlock()
	if current >= len(*data) {
		return
	}
	(*results)[current] = processFunc((*data)[current])
}

// ParallelWorker processes elements in a slice concurrently using a specified number of workers.
// It applies the provided processFunc to each element in the slice and returns a new slice of results.
//
// Parameters:
//
//	data - The slice of input data to be processed.
//	workers - The number of worker goroutines to use for processing. If workers is less than or equal to 0,
//	          it defaults to 1. The number of workers is capped at twice the number of CPU cores.
//	processFunc - A function that takes an element of type T and returns a result of type R.
//
// The function ensures thread-safe access to shared resources and recovers from any panics that occur
// during the processing of slice elements. It waits for all worker goroutines to complete before returning
// the processed results.
func ParallelWorker[T any, R any](
	data []T,
	workers int,
	processFunc func(T) R,
) []R {
	if workers <= 0 {
		workers = 1
	}
	cpuCount := runtime.NumCPU() * 2
	if workers > cpuCount {
		workers = cpuCount
	}

	results := make([]R, len(data))
	var wg sync.WaitGroup
	var mu sync.Mutex

	last := -1

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for last < len(data) {
				func() {
					defer func() { // Recover from panic
						if r := recover(); r != nil {
							fmt.Println("ParallelWorker Recover: ", r)
						}
					}()
					runJob(&mu, &last, &data, &results, processFunc)
				}()
			}
		}()
	}

	wg.Wait()

	return results
}

// runMapJob processes a single item from the data map using the provided processFunc.
// It updates the results map with the processed result in a thread-safe manner.
//
// Parameters:
//
//	mu - A mutex used to ensure exclusive access to the shared last index.
//	last - A pointer to an integer representing the last processed index.
//	keys - A pointer to the slice of keys in the map.
//	data - A pointer to the map of input data to be processed.
//	processFunc - A function that takes an element of type T and returns a result of type T.
func runMapJob[T any](mu *sync.Mutex, last *int, keys *[]string, data *map[string]T, processFunc func(T) T) {
	if data == nil || last == nil || keys == nil || mu == nil {
		return
	}

	mu.Lock()
	current := *last
	current++
	*last = current
	if current >= len(*data) {
		mu.Unlock()
		return
	}
	mu.Unlock()

	key := (*keys)[current]
	mu.Lock()
	dataToProcess, exists := (*data)[key]
	mu.Unlock()

	if !exists {
		return
	}
	result := processFunc(dataToProcess)

	mu.Lock()
	(*data)[key] = result
	mu.Unlock()
}

// ParallelMapWorker processes elements in a map concurrently using a specified number of workers.
// It applies the provided processFunc to each element in the map and updates the map in place.
//
// Parameters:
//
//	data - A pointer to the map of input data to be processed.
//	workers - The number of worker goroutines to use for processing. If workers is less than or equal to 0,
//	          it defaults to 1. The number of workers is capped at twice the number of CPU cores.
//	processFunc - A function that takes an element of type T and returns a processed result of the same type.
//
// The function ensures thread-safe access to shared resources and recovers from any panics that occur
// during the processing of map elements. It waits for all worker goroutines to complete before returning.
func ParallelMapWorker[T any](
	data *map[string]T,
	workers int,
	processFunc func(T) T,
) {
	if workers <= 0 {
		workers = 1
	}
	cpuCount := runtime.NumCPU() * 2
	if workers > cpuCount {
		workers = cpuCount
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	var keys []string
	for k := range *data {
		keys = append(keys, k)
	}

	last := -1

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for last < len(*data) {
				func() {
					defer func() { // Recover from panic
						if r := recover(); r != nil {
							fmt.Println("ParallelWorker Recover: ", r)
						}
					}()
					runMapJob(&mu, &last, &keys, data, processFunc)
				}()
			}
		}()
	}

	wg.Wait()
}
