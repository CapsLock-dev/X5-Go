package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if len(words) == 0 {return ""}
	first := []rune(words[0])
	for i, curr := range first {
		for _, other_str := range words[1:] {
			other := []rune(other_str)
			if i >= len(other) || other[i] != curr {
				return string(first[:i])
			}
		}
	}
	return words[0]
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	m := make(map[int]int)
	for i := 0; i<len(input.Nums); i++ {
		comp := input.Target - input.Nums[i]
		index, ok := m[comp]
		if ok {
			return [2]int{index, i}
		}
		m[input.Nums[i]] = i
	}
	return [2]int{}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	m := make(map[int]int)
	for i := 0; i<len(nums); i++ {
		_, ok := m[nums[i]]
		if ok {
			return true;
		}
		m[nums[i]] = i;
	}
	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if len(nums) == 0 {return []int{}}
	res := make([]int, 0, len(nums))
	res = append(res, nums[0])
	for i := 1; i<len(nums); i++ {
		if nums[i] != nums[i-1] {
			res = append(res, nums[i])
		}
	}
	return res
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	runes := []rune(text)
	st := make([]rune, 0, len(runes));
	for _, c := range runes {
		switch c {
		case '(', '[', '{':
			st = append(st, c);
		case ')', ']', '}':
			if len(st) == 0 {return false}
			top := st[len(st)-1]
			if (c == ')' && top != '(') || 
			   (c == ']' && top != '[') ||
			   (c == '}' && top != '{') {
				return false	
			}
			st = st[:len(st)-1]
		}

	}
	return len(st) == 0;
}
