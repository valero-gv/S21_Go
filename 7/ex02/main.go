package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sort"
)

func minCoins(val int, coins []int) []int {
	res := make([]int, 0)
	i := len(coins) - 1
	for i >= 0 {
		for val >= coins[i] {
			val -= coins[i]
			res = append(res, coins[i])
		}
		i -= 1
	}
	return res
}

func minCoins2(val int, coins []int) []int {
	if val <= 0 || len(coins) == 0 {
		// Если нужно набрать 0 или монет нет - возвращаем пустой набор
		return nil
	}

	// Удалим дубли, отсортируем
	uniqueCoins := make(map[int]bool)
	for _, c := range coins {
		uniqueCoins[c] = true
	}
	var sortedCoins []int
	for c := range uniqueCoins {
		sortedCoins = append(sortedCoins, c)
	}
	sort.Ints(sortedCoins)

	// Создаём dp-массив
	// dp[i] = минимальное число монет для i
	// choice[i] = какая последняя монета была выбрана, чтобы получить dp[i]
	dp := make([]int, val+1)
	choice := make([]int, val+1)
	const INF = 1_000_000_000

	for i := 1; i <= val; i++ {
		dp[i] = INF
		choice[i] = -1
	}

	dp[0] = 0 // 0 монет для суммы 0

	for i := 1; i <= val; i++ {
		for _, c := range sortedCoins {
			if c <= i && dp[i-c] != INF && dp[i-c]+1 < dp[i] {
				dp[i] = dp[i-c] + 1
				choice[i] = c
			}
		}
	}

	// Если dp[val] так и осталось INF, значит собрать сумму val не можем
	if dp[val] == INF {
		return nil
	}

	// Восстанавливаем набор монет из choice
	var res []int
	cur := val
	for cur > 0 {
		c := choice[cur]
		if c == -1 {
			// Такое в принципе уже не должно случиться, но на всякий случай
			break
		}
		res = append(res, c)
		cur -= c
	}

	return res
}

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	val := 13
	coins := []int{1, 5, 10}
	fmt.Println(minCoins2(val, coins))

	// Block forever
	select {}
}
