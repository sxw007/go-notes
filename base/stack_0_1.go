package main

func a() *int {
	v := 0 // 若回收a()的所有栈帧回收，此处就会变为空指针，所以变量v不能放到栈上，而是放在堆上
	return &v
}

func main() {
	i := a() // 返回的是变量v的指针
	print(i)
}
