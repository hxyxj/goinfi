// Copyright 2013 Jerry Hongy.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lua

import (
	"fmt"
	"goinfi/base"
	"reflect"
	"testing"
)

type Runner struct {
	vm *VM
	t  *testing.T
}

func NewRunner(t *testing.T) *Runner {
	r := new(Runner)
	vm := NewVM()
	vm.Openlibs()
	r.vm = vm
	r.t = t
	return r
}

func (r *Runner) End() {
	r.vm.Close()
}

func (r *Runner) E(s string) []interface{} {
	value, err := r.vm.EvalStringWithError(s)
	if err != nil {
		r.t.Errorf("eval error: %v", err)
	}
	return value
}

func (r *Runner) E_MustError(s string) []interface{} {
	value, err := r.vm.EvalStringWithError(s)
	if err == nil {
		r.t.Errorf("must error: %v", err)
	}
	return value
}

func (r *Runner) AssertEqual(a interface{}, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		r.t.Errorf("%v, %v no equal", a, b)
		return
	}
}

func (r *Runner) AssertNoEqual(a interface{}, b interface{}) {
	if reflect.DeepEqual(a, b) {
		r.t.Errorf("%v, %v no equal", a, b)
		return
	}
}

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
		"a": 100, "b": 200, "c": 300,
	}
}

// this is a raw function
func GetHello(state State) int {
	state.Pushstring("hello")
	return 1
}

type allMyStruct struct {
	*Point
	*Rect
	*DoublePoint
}

func TestLua_base(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}
	var expect []interface{}

	//
	// add struct type
	//
	r.vm.AddStructList(struct {
		*Point
		*Rect
		*DoublePoint
	}{})

	//
	// add function
	//
	r.vm.AddFuncList("test", []base.KeyValue{
		{"NewPoint", NewPoint},
		{"NewDoublePoint", NewDoublePoint},
		{"NewIntSlice", NewIntSlice},
		{"NewStrIntMap", NewStrIntMap},
		{"hello.world.GetHello", GetHello},
	})

	result = r.E(`
		return test.NewPoint ~= nil, test.NewDoublePoint ~= nil,
			test.NewIntSlice ~= nil, test.NewStrIntMap ~= nil,
			test.NewStrIntMap ~= nil, test.hello.world.GetHello ~= nil
	`)
	expect = []interface{}{true, true, true, true, true, true}
	r.AssertEqual(result, expect)

	//
	// call function
	//
	result = r.E(`
		word = test.hello.world.GetHello()
		return word
	`)
	expect = []interface{}{"hello"}
	r.AssertEqual(result, expect)

	//
	// create object
	//
	result = r.E(`
		obj = test.NewPoint(1, 2)
		return obj.X, obj.Y
	`)
	expect = []interface{}{1.0, 2.0}
	r.AssertEqual(result, expect)

	//
	// call method
	//
	result = r.E(`
		return obj:SumXY()
	`)
	expect = []interface{}{3.0}
	r.AssertEqual(result, expect)

	//
	//  map
	//
	result = r.E(`
		map = test.NewStrIntMap()
		return #map, map.c, map.not_exist_field
	`)
	expect = []interface{}{3.0, 300.0, nil}
	r.AssertEqual(result, expect)

	result = r.E(`
		map.c = 400
		return map.c
	`)
	expect = []interface{}{400.0}
	r.AssertEqual(result, expect)

	result = r.E_MustError(`
		map[1] = 4
		return map.c
	`)
	expect = []interface{}{}
	r.AssertEqual(result, expect)

	//
	// slice
	//
	result = r.E(`
		slice = test.NewIntSlice()
		return #slice, slice[0], slice[1], slice[2], slice[3]
	`)
	expect = []interface{}{4.0, 1.0, 2.0, 3.0, 4.0}
	r.AssertEqual(result, expect)

	result = r.E(`
		slice[0] = 100
		return slice[0]
	`)
	expect = []interface{}{100.0}
	r.AssertEqual(result, expect)

	result = r.E_MustError(`
		slice[-1] = 200
		return slice[-1]
	`)
	expect = []interface{}{}
	r.AssertEqual(result, expect)

	result = r.E_MustError(`
		slice['key'] = 200
		return slice['key']
	`)
	expect = []interface{}{}
	r.AssertEqual(result, expect)

	//
	// embed struct
	//
	result = r.E(`
		doublePoint = test.NewDoublePoint()
		return doublePoint.P1_X
	`)
	expect = []interface{}{0.0}
	r.AssertEqual(result, expect)

	result = r.E(`
		doublePoint.P1_X = 123
		doublePoint.P2_X = 456
		return doublePoint.P1_X, doublePoint.P2_X
	`)
	expect = []interface{}{123.0, 456.0}
	r.AssertEqual(result, expect)

	result = r.E_MustError(`
		doublePoint.P1_K = 789
		return doublePoint.P1_K
	`)
	expect = []interface{}{}
	r.AssertEqual(result, expect)
}

func wrongRawFunc1(state *State) int {
	return 0
}

func wrongRawFunc2(state State) (int, int) {
	return 0, 0
}

func wrongRawFunc3(i int, state State) int {
	return 0
}

func TestLua_rawfunc(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var ok bool
	var err error

	ok, err = r.vm.AddFunc("WrongRawFunc1", wrongRawFunc1)
	r.AssertEqual(ok, false)
	r.AssertNoEqual(err, nil)

	ok, err = r.vm.AddFunc("WrongRawFunc2", wrongRawFunc2)
	r.AssertEqual(ok, false)
	r.AssertNoEqual(err, nil)

	ok, err = r.vm.AddFunc("WrongRawFunc3", wrongRawFunc3)
	r.AssertEqual(ok, false)
	r.AssertNoEqual(err, nil)
}

func TestLua_luafunc(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}
	var expect []interface{}
	var fn *Function
	var err error

	// int as arg
	result = r.E(`
		return function() return 1 end
	`)
	fn = result[0].(*Function)
	result, err = fn.Call()
	r.AssertEqual(err, nil)
	expect = []interface{}{1.0}
	r.AssertEqual(result, expect)
	fn.Release()

	// string and int as arg, return multi-value
	result = r.E(`
		return function(s, i) return s, i end
	`)
	fn = result[0].(*Function)
	result, err = fn.Call("s", 2)
	r.AssertEqual(err, nil)
	expect = []interface{}{"s", 2.0}
	r.AssertEqual(result, expect)
	fn.Release()

	result = r.E(`
		return function(s, i) return s, i end
	`)
	fn = result[0].(*Function)
	result, err = fn.Call("s", 2)
	r.AssertEqual(err, nil)
	expect = []interface{}{"s", 2.0}
	r.AssertEqual(result, expect)
	fn.Release()

	// sum
	result = r.E(`
		return function(a, b) return a+b end
	`)
	fn = result[0].(*Function)
	result, err = fn.Call(23, 8)
	r.AssertEqual(err, nil)
	expect = []interface{}{31.0}
	r.AssertEqual(result, expect)
	fn.Release()

	// return table
	result = r.E(`
		return function(a, b) return {a, b} end
	`)
	fn = result[0].(*Function)
	result, err = fn.Call("value1", "value2")
	r.AssertEqual(err, nil)
	T := result[0].(*Table)
	r.AssertEqual(T.Get(1), "value1")
	r.AssertEqual(T.Get(2), "value2")
	T.Release()
	fn.Release()

	// String()
	result = r.E(`
		return function() end
	`)
	fn = result[0].(*Function)
	s := fmt.Sprintf("%v", fn.String())
	r.AssertNoEqual(len(s), 0)
	fn.Release()
}

func TestLua_luatable(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}
	//var expect []interface{}
	var tbl *Table
	//var err error

	// get table
	result = r.E(`
		T1 = { 'v1', 'v2', 'v3' }
		return T1
	`)
	tbl = result[0].(*Table)
	r.AssertEqual(tbl.Get(1), "v1")
	r.AssertEqual(tbl.Get(2), "v2")
	r.AssertEqual(tbl.Get(3), "v3")
	tbl.Release()

	// set table 1
	result = r.E(`
		T2 = { }
		return T2
	`)
	tbl = result[0].(*Table)
	tbl.Set(1, "my1")
	tbl.Set("key", "myvalue")
	r.AssertEqual(tbl.Get(1), "my1")
	r.AssertEqual(tbl.Get("key"), "myvalue")
	tbl.Release()

	// set table 2
	result = r.E(`
		T3 = { }
		return T3
	`)
	tbl = result[0].(*Table)
	ok, err := tbl.Set(nil, "my1")
	// will return: false, invalid key type for lua type
	r.AssertEqual(ok, false)
	r.AssertNoEqual(err, nil)
	tbl.Release()

	// sub table
	result = r.E(`
		T4 = {
			name = "foo",
			info = {
				email = "foo@bar.com",
				place = "cn",
				birth = "2012-12-21",
			}
		}
		return T4
	`)
	tbl = result[0].(*Table)
	info := tbl.Get("info")
	r.AssertNoEqual(info, nil)
	tinfo := info.(*Table)
	r.AssertEqual(tinfo.Get("place"), "cn")
	tinfo.Release()
	tbl.Release()
}

func TestLua_vararg(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}

	SumAll := func(init int, values ...int) int {
		for i := 0; i < len(values); i++ {
			init += values[i]
		}
		return init
	}

	r.vm.AddFunc("SumAll", SumAll)

	result = r.E(`
		return SumAll(1000, 2, 3, 4, 5, 6)
	`)
	r.AssertEqual(result[0], 1020.0)
}

func TestLua_interface(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}
	var expect []interface{}

	CallMe := func(a ...interface{}) []interface{} {
		return a
	}

	r.vm.AddFunc("CallMe", CallMe)

	result = r.E(`
		return CallMe(1, true, 'abc', nil)
	`)
	expect = []interface{}{1.0, true, "abc", nil}
	r.AssertEqual(result[0], expect)
}

func TestLua_keyvalue(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}
	var expect []interface{}

	r.vm.AddFunc("CallMe", func(kvs []base.KeyValue) ([]base.KeyValue, []interface{}) {
		array := make([]interface{}, 0)
		for _, kv := range kvs {
			array = append(array, kv.Key)
			array = append(array, kv.Value)
		}
		return kvs, array
	})

	result = r.E(`
		t = { a=1, b=2 }
		return CallMe(t)
	`)

	r.AssertEqual(reflect.TypeOf(result[0]), reflect.TypeOf(theNullSliceKeyValue))
	r.AssertEqual(result[0].([]base.KeyValue)[0].Key, "a")
	r.AssertEqual(result[0].([]base.KeyValue)[0].Value, 1.0)
	r.AssertEqual(result[0].([]base.KeyValue)[1].Key, "b")
	r.AssertEqual(result[0].([]base.KeyValue)[1].Value, 2.0)

	expect = []interface{}{"a", 1.0, "b", 2.0}
	r.AssertEqual(result[1], expect)
}
