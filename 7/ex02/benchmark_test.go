// benchmark_test.go
package main

import (
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
)

// Генерируем случайный набор монет для тестов производительности
func generateCoins(n int) []int {
	// Сгенерируем n случайных номиналов от 1 до 1000
	coins := make([]int, n)
	for i := 0; i < n; i++ {
		coins[i] = 1 + rand.Intn(1000)
	}
	return coins
}

func BenchmarkMinCoins(b *testing.B) {
	// Для воспроизводимости зафиксируем seed
	rand.Seed(42)
	coins := generateCoins(1000) // 1000 случайных номиналов
	val := 9999

	// Подготовим файл для записи CPU-профиля
	f, err := os.Create("cpu_profile_minCoins.pprof")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// Запускаем сам бенчмарк
	for i := 0; i < b.N; i++ {
		_ = minCoins(val, coins)
	}
}

func BenchmarkMinCoins2(b *testing.B) {
	rand.Seed(42)
	coins := generateCoins(1000)
	val := 9999

	f, err := os.Create("cpu_profile_minCoins2.pprof")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	for i := 0; i < b.N; i++ {
		_ = minCoins2(val, coins)
	}
}
