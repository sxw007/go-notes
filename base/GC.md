# 垃圾回收（Garbage Collection）

---

## 栈内存（协程栈、调用栈）

``堆上的栈，Go的协程栈位于堆内存上``

### 作用

<img title="" src="images/1_1_1.png" alt="stack" width="200">

```
每个协程第一个栈帧为 goexit()

每次调用其他函数会插入一个栈帧

用户的main方法首先会开辟一个main.main的栈帧

栈帧首先记录栈基址（就是指从哪个方法调用进来的）方便返回的时候知道返回地址在哪
开辟调用方法的返回值，return就是将返回值写回上一个栈帧预留的空间
```

- 协程的执行路径（do1() → do2()）
- 局部变量（方法内部声明的变量会记录在协程栈中）
- 函数传参（方法间的参数传递，例如do2()需要一个入参，do1()是通过栈内存把参数传递给do2()）
- 函数返回值（do2()有返回值给do1()，用的也是栈内存传递）

- 栈内存（协程栈、调用栈）
- 堆内存
- 垃圾回收



### 位置

- Go协程栈位于Go堆内存上（Go的特殊设计，C++，C#的栈区和堆区是分开的）
  - 通过GC来释放
- Go堆内存位于操作系统虚拟内存上（操作系统会给每个进程分配一块虚拟内存）


### 结构


```shell
go build -gcflags -S stack0.go
```
```go
package main

func sum(a, b int) int {
  s := 0
  s = a + b
  return s
}

func main()  {
  a := 3
  b := 5
  print(sum(a, b))
}
```

![stack demo](./images/1_1_2.png)

> 往后就是清理sum函数返回值、sum函数参数...，再给print开栈帧

### 总结

- 协程栈记录了协程的执行现场
- 协程栈还负责记录局部变量，传递参数和返回值
- Go使用参数拷贝传递（值传递）
  - sum函数传参的时候回开辟2个新的空间，将5、3拷贝进去
  - 推荐在代码中的结构体参数使用指针（节约内存）
    - 传递结构体时：会拷贝结构体中的全部内容
    - 传递结构体指针时：会拷贝结构体指针


### 思考

>初始大小2~4k

- 协程栈不够大怎么办？
  - 局部变量太大
  - 栈帧太多 

---

## 逃逸分析（从栈逃逸到堆上）

- 不是所有的变量都能放在协程栈上
- 栈帧回收后，需要继续使用的变量
- 太大的变量

### 指针逃逸

> 函数返回了对象的指针

```go
package main

func a() *int {
	v := 0 // 若回收a()的所有栈帧回收，此处就会变为空指针，所以变量v不能放到栈上，而是放在堆上
	return &v
}

func main() {
	i := a() // 返回的是变量v的指针
	print(i)
}
```

### 空接口逃逸

> 如果函数的参数为 interface{}，函数的实参很可能会逃逸
> 因为 interface{} 类型的函数往往会使用反射（反射要求对象是在堆上），未使用反射则不会逃逸

```go
package main

import "fmt"

func b() {
  i := 0 // 因为下面的 fmt.Println() 接收的是 interface{}，i会逃逸到堆上
  fmt.Println(i) // func Println(a ...interface{}) (n int, err error) {...}
}

func main() {
  b()
}
```

### 大变量逃逸

- 过大的变量会导致栈空间不足
- 64位机器中，一般没超过64KB的变量会逃逸

### 栈扩容

> 栈空间是从堆中申请的，可以多申请

- Go 栈的初始空间为2KB
- 在函数调用前判断栈空间（morestack），必要时堆栈进行扩容
- 早期使用分段栈，后期使用连续栈

#### 分段栈

> 1.13之前使用
> 
> 优点：没有空间浪费
> 
> 缺点：栈指针会在不连续的空间跳转（影响性能）

![stack demo](./images/1_1_3.png)

#### 连续栈

> 优点：空间一直连续
> 
> 缺点：伸缩时的开销大
> 
> 当空间不足时扩容，变为原来的2倍（老的栈空间不足时，会找一块2倍大的栈空间并拷贝过去）
> 
> 当空间使用率不足1/4时缩容，变为原来的1/2

![stack demo](./images/1_1_4.png)

./src/runtime/stubs.go:312 （使用汇编实现）
```go
func morestack() // 以64位为例：./src/runtime/asm_amd64.s
func morestack_noctxt()
```


## 堆内存

### 操作系统虚拟内存

- × 不是win的“虚拟内存”（内存不够的时候拿硬盘做虚拟内存）
- √ 操作系统给应用提供的虚拟内存空间
  - 系统会给每个进程一个虚拟的内存空间，而不是直接的物理内存，操作系统管理这些虚拟内存空间映射到物理内存空间
  - 背后是物理内存，也有可能有磁盘
- Linux获取虚拟内存：mmap、madvice

#### Linux（64位）

> 若虚拟内存超过物理内存(64GB)就是内存溢出（OOM），操作系统会杀掉进程

![stack demo](./images/1_1_5.png)

#### heapArena

- 在64位操作系统中(win除外) Go 每次申请的虚拟内存单元为64MB（以heapArena为单元申请，一次64MB，释放也是一次64MB）
- 最多有4,194,304个虚拟内存单元（2^20，刚好可以占满256TB）
- 所有的heapArena组成了mheap（Go堆内存）

![stack demo](./images/1_1_6.png)

> go heap 会按照arena的大小增长，每次预留arena大小整数倍的虚拟地址空间。arena的大小与平台相关，除了windows，其他系统64位的平台下arena的大小都是64M。在32位的平台中，为了使go heap比较连续，没有碎片，当程序启动的时候就会先预留一大块虚拟地址空间，如果这些空间都被用完了，才会每次按照arena大小整数倍去预留虚拟地址空间。

```
       Platform  Addr bits  Arena size  L1 entries   L2 entries
 --------------  ---------  ----------  ----------  -----------
       */64-bit         48        64MB           1    4M (32MB)
 windows/64-bit         48         4MB          64    1M  (8MB)
       */32-bit         32         4MB           1  1024  (4KB)
     */mips(le)         31         4MB           1   512  (2KB)
```

./src/runtime/mheap.go:229
```go
// 62行，mheap
type mheap struct {
	// ...
	// ↓ ↓ ↓ ↓ 157行 ↓ ↓ ↓ ↓
    arenas [1 << arenaL1Bits]*[1 << arenaL2Bits]*heapArena // 记录向操作系统申请的所有内存单元
	// ...
}

// 229行，这个结构体描述了一个64MB的内存单元（不是一个结构体64MB），记录向操作系统申请64MB虚拟内存的信息
// bitmap、pageMarks、pageSpecials都与GC有关
type heapArena struct {
    bitmap [heapArenaBitmapBytes]byte // 用于记录这个arena中有哪些位置有指针
    spans [pagesPerArena]*mspan  // 内存管理单元
    pageInUse [pagesPerArena / 8]uint8
    pageMarks [pagesPerArena / 8]uint8
    pageSpecials [pagesPerArena / 8]uint8
    checkmarks *checkmarksMap
    zeroedBase uintptr
}
```

#### 内存管理单元

##### 分级分配思想
为了减少每个对象会放入可容纳该对象的最小的区域内
![stack demo](./images/1_1_7.png)

##### mspan

``上面所述的“级”就是 “内存管理单元 mspan”``

- 根据隔离适应策略，使用内存时的最小单位为mspan
- 每个mspan为N个大小相同的“格子”
- Go中一共有67种mspan

> 源码详情：./src/runtime/sizeclasses.go

```
   级别    对象大小  格子的大小   对象数  页面尾部浪费   最大浪费
 class  bytes/obj  bytes/span  objects  tail waste  max waste
     1          8        8192     1024           0     87.50%
     2         16        8192      512           0     43.75%
     3         32        8192      256           0     46.88%
    ...
    10        128        8192       64           0     11.72%
    11        144        8192       56         128     11.82%
    ...
    37       1792       16384        9         256     15.57%
    38       2048        8192        4           0     12.45%
    39       2304       16384        7         256     12.46%
    ...
    66      28672       57344        2           0      4.91%
    67      32768       32768        1           0     12.50%
```
> 因为mspan管理内存的最小单位是页面，而页面的大小不一定是size class大小的倍数，这会导致一些内存被浪费
> 
> 例如下图中一个mspan划分成若干个slot用于分配，但是mspan占用页面的大小不能被slot的大小整除，所以有一个tail waste

![stack demo](./images/1_1_8.png)

> mspan：./src/runtime/mheap.go:384

#### 中心索引 mcentral

- 136个mcentral结构体
  - 68个组需要GC扫描的mspan（堆中的对象）
  - 68个组不需要GC扫描的mspan（常量）




## 垃圾回收（GC）
