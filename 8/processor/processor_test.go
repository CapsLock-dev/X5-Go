package processor

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	path := filepath.Join("..", "testdata", "orders-a.txt")
	got := analyzeFile(path)
	if got.Err != nil || got.Path != path || got.Lines != 3 || got.Words != 5 {
		t.Fatalf("результат = %+v, ожидаем этот путь, 3 строки и 5 слов", got)
	}
}

func TestProcessFiles(t *testing.T) {
	paths := []string{
		filepath.Join("..", "testdata", "orders-a.txt"),
		filepath.Join("..", "testdata", "orders-b.txt"),
		filepath.Join("..", "testdata", "empty.txt"),
	}
	want := map[string][2]int{
		paths[0]: {3, 5},
		paths[1]: {2, 4},
		paths[2]: {0, 0},
	}

	for got := range ProcessFiles(paths, 2) {
		stats, found := want[got.Path]
		if !found {
			t.Fatalf("лишний результат для %q", got.Path)
		}
		if got.Err != nil || got.Lines != stats[0] || got.Words != stats[1] {
			t.Fatalf("для %q получили %+v, ожидаем строк и слов %v", got.Path, got, stats)
		}
		delete(want, got.Path)
	}
	if len(want) > 0 {
		t.Fatalf("не получены результаты: %v", want)
	}
}

func TestInvalidFile(t *testing.T) {
	paths := []string{
		filepath.Join("..", "testdata", "orders-a.txt"),
		filepath.Join("..", "testdata", "invalid_file"),
		filepath.Join("..", "testdata", "empty.txt"),
	}

	want := map[string]Result{
		paths[0]: {paths[0], 3, 5, nil},
		paths[1]: {paths[1], 0, 0, errors.New("test")},
		paths[2]: {paths[2], 0, 0, nil},
	}

	for got := range ProcessFiles(paths, 2) {
		stats, found := want[got.Path]
		if !found {
			t.Fatalf("лишний результат для %q", got.Path)
		}
		if stats.Err != nil {
			if got.Err == nil {
				t.Fatalf("для %q ожидали ошибку", got.Path)
			}
		} else if got.Err != nil {
			t.Fatalf("для %q не ожидали ошибку, но получили: %v", got.Path, got.Err)
		}

		if got.Lines != stats.Lines || got.Words != stats.Words {
			t.Fatalf("для %q получили %+v, ожидаем строк и слов %v", got.Path, got, stats)
		}
		delete(want, got.Path)
	}
	if len(want) > 0 {
		t.Fatalf("не получены результаты: %v", want)
	}
}

func TestProcessFilesEmpty(t *testing.T) {
	paths := []string{}
	for range ProcessFiles(paths, 2) {
		t.Fatal("не ожидали результатов")
	}
}

func TestFileDuplicate(t *testing.T) {
	paths := []string{
		filepath.Join("..", "testdata", "orders-a.txt"),
		filepath.Join("..", "testdata", "orders-b.txt"),
		filepath.Join("..", "testdata", "empty.txt"),
		filepath.Join("..", "testdata", "orders-a.txt"),
		filepath.Join("..", "testdata", "orders-a.txt"),
	}
	want := map[string][2]int{
		paths[0]: {3, 5},
		paths[1]: {2, 4},
		paths[2]: {0, 0},
	}
	count := 0

	for got := range ProcessFiles(paths, 2) {
		stats := want[got.Path]
		if got.Err != nil || got.Lines != stats[0] || got.Words != stats[1] {
			t.Fatalf("для %q получили %+v, ожидаем строк и слов %v", got.Path, got, stats)
		}
		count++
	}
	if count < len(paths) {
		t.Fatalf("дубликаты не обработаны")
	}
}
