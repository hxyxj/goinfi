package main

import (
	"fmt"
	"goinfi/lua"
	"io"
	"os"
	"strings"
)

type Point struct {
	X int
	Y int
}

func (p *Point) SumXY() int {
	return p.X + p.Y
}

type DoublePoint struct {
	P1 Point
	P2 Point
}

type Rect struct {
	Left   int
	Top    int
	Width  int
	Height int
}

type allMyStruct struct {
	*Point
	*Rect
	*DoublePoint
}

func NewPoint(x, y int) *Point {
	return &Point{x, y}
}

func NewDoublePoint() *DoublePoint {
	return new(DoublePoint)
}

func NewIntSlice() []int {
	return []int{1, 2, 3, 4}
}

func NewStrIntMap() map[string]int {
	return map[string]int{
		"a": 1, "b": 2, "c": 3,
	}
}

// this is a raw function
func GetHello(state lua.State) int {
	state.Pushstring("hello world")
	return 1
}

func WrongRawFunc1(state *lua.State) int {
	return 0
}

func WrongRawFunc2(state lua.State) (int, int) {
	return 0, 0
}

func WrongRawFunc3(i int, state lua.State) int {
	return 0
}

func Test1() {
	vm := lua.NewVM()
	defer vm.Close()

	vm.Openlibs()

	vm.AddStructList(allMyStruct{})

	vm.AddFunc("NewPoint", NewPoint)
	vm.AddFunc("NewDoublePoint", NewDoublePoint)
	vm.AddFunc("NewIntSlice", NewIntSlice)
	vm.AddFunc("NewStrIntMap", NewStrIntMap)
	vm.AddFunc("go.golang", GetHello)

	_, err := vm.AddFunc("go.golang.GetHello", GetHello)
	fmt.Println("AddFunc", err)

	var ES = func(s string) {
		_, err := vm.EvalStringWithError(s, 0)
		if err != nil {
			fmt.Println("> ES", err)
		}
	}
	var EB = func(buf io.Reader) {
		_, err := vm.EvalBufferWithError(buf, 0)
		if err != nil {
			fmt.Println("> EB", err)
		}
	}

	ES("pt = NewPoint(1, 2)")

	ES("print(pt.X, pt.Y)")

	ES("print(pt.SumXY)")

	// string reader
	EB(strings.NewReader("print(pt:SumXY())"))

	// file reader
	f, err := os.Open("baselib.lua")
	if err != nil {
		fmt.Println(err)
	} else {
		defer f.Close()
		EB(f)
	}

	ES("map = NewStrIntMap(); print('map[a]', #map, map['c'], map['x'])")
	ES("map['c']=nil; print('map[c]', map['c'])")
	ES("map['c']=4; print('map[c]', map['c'])")
	ES("map[1]=4")
	ES("map['1']='4'")

	ES("slice = NewIntSlice(); print('slice[0]', #slice, slice[0])")
	ES("slice[0]=100; print(slice[0])")
	ES("slice[-1]=200")
	ES("slice['a']=200")

	ES("dp = NewDoublePoint(); print('dp.P1.X', dp.P1_X)")
	ES("dp.P1_X = 100; print('dp.P1.X', dp.P1_X)")
	ES("dp.P1_X = '100'")
	ES("dp.P1_K = 100")

	ok, err := vm.AddFunc("", WrongRawFunc1)
	fmt.Println("WrongRawFunc1", ok, err)

	ok, err = vm.AddFunc("", WrongRawFunc2)
	fmt.Println("WrongRawFunc2", ok, err)

	ok, err = vm.AddFunc("", WrongRawFunc3)
	fmt.Println("WrongRawFunc3", ok, err)

	ES("print(go.golang.GetHello())")
	ES("print(go.golang())")

	ES("for _, k in ipairs(golang.Keys(map)) do print(k) end")
	ES("print(golang.HasKey(map, 1))")
	ES("print(golang.HasKey(map, 'k'))")
	ES("print(golang.HasKey(map, 'a'))")
}

func main() {
	Test1()
}

func oldTest() {
	/*
		vm.AddFunc("foo", func() {
			fmt.Println("this is function foo")
		})

		vm.AddFunc("myadd", func(a, b int) int {
			return a+b
		})

		vm.AddFunc("myconcat", func(a, b string) string {
			return a + "," + b
		})

		vm.AddFunc("get2d", func() Point {
			return Point {10, 10}
		})

		vm.AddFunc("add2d", func(a *Point, b *Point) Point {
			return Point { a.X+b.X, a.Y+b.Y }
		})

		vm.Dostring("print('foo', pcall(function() foo() end))")
		vm.Dostring("print('myadd', pcall(function() print('result=', myadd(1, 2)) end))")
		vm.Dostring("print('myconcat', pcall(function() print(myconcat('1', '2')) end))")
		vm.Dostring("print('add2d', pcall(function() print(add2d(get2d(), get2d())) end))")
	*/
}
