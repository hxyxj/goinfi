package main

import (
	"fmt"
	"os"
	"io"
	//"strings"
	// "bytes"
	"goinfi/lua"
)

func Test1() {
	vm := lua.NewVM()
	defer vm.Close()
	vm.Openlibs()
	var EB = func(buf io.Reader) {
		ok, err := vm.ExecBuffer(buf)
		if err != nil {
			fmt.Println("> EB", ok, err)
		}
	}
	var ES = func(s string) {
		ok, err := vm.ExecString(s)
		if err != nil {
			fmt.Println("> ES", ok, err)
		}
	}

	f, err := os.Open("example/test1.lua")
	if err != nil {
		fmt.Println(err)
	} else {
		defer f.Close()
		EB(f)
	}

	ES(`
	bin = pack.Pack(1, 2, {a=1,b=2,c=3, 1, 2})
	print('#bin', #bin)
	print(sys.dump({pack.Unpack(bin)}))
	`)

	ES(`
	bin = pack.Pack('key', {a=1})
	print('#bin', #bin)
	print(sys.dump({pack.Unpack(bin)}))
	`)

	ES(`
	bin = pack.Pack({a={b={c={d={e='depth'}}}}})
	print('#bin', #bin)
	print(sys.dump({pack.Unpack(bin)}))
	`)
}

func main() {
	Test1()
}
