package lesson2

func FizzBuzz(n int) string {
	switch {
	case n <= 0:
		return "invalid"
	case n%3 == 0 && n%5 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	}
	return ""
}

func Sum(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	sum := int64(0)
	for _, n := range values {
		sum += n
	}
	return sum
}

func Min(values []int64) int64 {
	if (len(values) == 0) {return 0}
	min := int64(values[0])
	for _, n := range values {
		if n < min {
			min = n
		}
	}
	return min
}

func Max(values []int64) int64 {
	if (len(values) == 0) {return 0}
	max := int64(values[0])
	for _, n := range values {
		if n > max {
			max = n
		}
	}
	return max
}

func Average(values []int64) float64 {
	if (len(values) == 0) {return 0}
	sum := int64(0)
	for _, n := range values {
		sum += n
	}
	return float64(sum) / float64(len(values))
}

func Category(amount int64) string {
	switch {
	case amount < 0:
		return "invalid"
	case amount < 10000:
		return "small"
	case amount < 50000:
		return "regular"
	default:
		return "large"
	}
}

func RuneCount(text string) int {
	count := int(0)
	for range text {
		count++
	}
	return count
}
