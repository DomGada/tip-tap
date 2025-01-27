package shared

func Sizeof(bits int) int {
	num := 1
	toAdd := 0
	if num%2 == 1 {
		toAdd = 1
	}
	for range bits {
		num = num << 1
		num += toAdd
	}
	return num
}
