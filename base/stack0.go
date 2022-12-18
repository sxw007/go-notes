package main

func sum(a, b int) int {
	s := 0
	s = a + b
	return s
}

func main() {
	a := 3
	b := 5
	print(sum(a, b))
}
