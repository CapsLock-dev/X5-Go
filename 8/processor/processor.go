package processor

import "sync"

// ProcessFiles возвращает по результату на каждый переданный путь.
// Получатель читает канал до закрытия.
func ProcessFiles(paths []string, workers int) <-chan Result {
	if workers < 1 {
		workers = 1
	}
	results := make(chan Result)
	jobs := make(chan string)
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for path := range jobs {
				res := analyzeFile(path)
				results <- res
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, path := range paths {
			jobs <- path
		}
	}()
	go func() {
		defer close(results)
		wg.Wait()
	}()
	// TODO: распределите пути между workers обработчиками через канал jobs.
	// Для каждого пути вызовите analyzeFile и отправьте результат в results.
	// Определите, кто закроет каждый канал и как для этого использовать WaitGroup.
	return results
}
