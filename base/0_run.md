# Go程序是如何运行的？

## Go 程序的入口？

- runtime/rt0_XXX.s（汇编）

  1 ./src/runtime/rt0_linux_amd64.s

  ```plan9_x86
  #include "textflag.h"
  
  TEXT _rt0_amd64_linux(SB),NOSPLIT,$-8
      JMP	_rt0_amd64(SB)
  
  TEXT _rt0_amd64_linux_lib(SB),NOSPLIT,$0
      JMP	_rt0_amd64_lib(SB)
  
  ```
  2 ./src/runtime/as_amd64.s:15（不管windows还是Linux都会调_rt0_amd64()）
  ```plan9_x86
  // 主要是把argc、argv放到寄存器里面，
  TEXT _rt0_amd64(SB),NOSPLIT,$-8
      MOVQ	0(SP), DI	// argc
      LEAQ	8(SP), SI	// argv（执行参数）
      JMP	runtime·rt0_go(SB)
  ```

