package ex00

import (
	"testing"
)

func TestMinCoins2_OptimalOnNonCanonicalSet(t *testing.T) {
	coins := []int{1, 3, 4}
	val := 6
	got := MinCoins2(val, coins)
	// Оптимальный результат: [3, 3] (или то же самое по порядку).
	// Но порядок в нашем результате может быть любым (например, [3,3] или [3,3]),
	// мы проверим лишь что длина = 2 и сумма = 6.
	if len(got) != 2 {
		t.Errorf("minCoins2(%d, %v) дал набор %v (длина %d), ожидается 2 монеты",
			val, coins, got, len(got))
	}
	sum := 0
	for _, c := range got {
		sum += c
	}
	if sum != val {
		t.Errorf("Сумма монет %v = %d, а надо %d", got, sum, val)
	}
}

func TestMinCoins2_UnsortedAndDuplicates(t *testing.T) {
	coins := []int{3, 3, 1, 4, 1}
	val := 6
	got := MinCoins2(val, coins)
	// Оптимальный результат: 2 монеты (3 и 3).
	if len(got) != 2 {
		t.Errorf("minCoins2(%d, %v) дал набор %v (длина %d), ожидается 2 монеты",
			val, coins, got, len(got))
	}
	sum := 0
	for _, c := range got {
		sum += c
	}
	if sum != val {
		t.Errorf("Сумма монет %v = %d, а надо %d", got, sum, val)
	}
}

func TestMinCoins2_EmptyCoins(t *testing.T) {
	val := 10
	coins := []int{}
	got := MinCoins2(val, coins)
	if len(got) != 0 {
		t.Errorf("minCoins2(%d, %v) = %v, ожидается пустой набор", val, coins, got)
	}
}
