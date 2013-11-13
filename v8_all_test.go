package v8

import (
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var Default = NewEngine()

func init() {
	// traceDispose = true
	rand.Seed(time.Now().UnixNano())
	go func() {
		for {
			input, err := ioutil.ReadFile("test.cmd")

			if err == nil && len(input) > 0 {
				ioutil.WriteFile("test.cmd", []byte(""), 0744)

				cmd := strings.Trim(string(input), " \n\r\t")

				var p *pprof.Profile

				switch cmd {
				case "lookup goroutine":
					p = pprof.Lookup("goroutine")
				case "lookup heap":
					p = pprof.Lookup("heap")
				case "lookup threadcreate":
					p = pprof.Lookup("threadcreate")
				default:
					println("unknow command: '" + cmd + "'")
				}

				if p != nil {
					file, err := os.Create("test.out")
					if err != nil {
						println("couldn't create test.out")
					} else {
						p.WriteTo(file, 2)
					}
				}
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

func rand_sched(max int) {
	for j := rand.Intn(max); j > 0; j-- {
		runtime.Gosched()
	}
}

func Test_HelloWorld(t *testing.T) {
	context := Default.NewContext()
	script := context.Compile("'Hello ' + 'World!'", nil, nil)
	value := script.Run(context)
	result := value.ToString()

	if result != "Hello World!" {
		t.FailNow()
	}

	runtime.GC()
	//println(result)
}

func Test_PreCompile(t *testing.T) {
	code := "'Hello ' + 'PreCompile!'"
	scriptData1 := Default.PreCompile(code)

	data := scriptData1.Data()
	scriptData2 := NewScriptData(data)

	context := Default.NewContext()
	script := context.Compile(code, nil, scriptData2)
	value := script.Run(context)
	result := value.ToString()

	if result != "Hello PreCompile!" {
		t.FailNow()
	}

	runtime.GC()
	//println(result)
}

func Test_Object(t *testing.T) {
	context := Default.NewContext()
	script := context.Compile("a={'a':123};", nil, nil)
	value := script.Run(context)
	result := value.ToObject()

	if prop := result.GetProperty("a"); prop != nil {
		if !prop.IsNumber() || prop.GetNumber() != 123 {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if !result.SetProperty("a", Default.True(), PA_None) {
		t.FailNow()
	}

	if prop := result.GetProperty("a"); prop != nil {
		if !prop.IsBoolean() || !prop.IsTrue() {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if !result.SetProperty("b", Default.False(), PA_None) {
		t.FailNow()
	}

	if prop := result.GetProperty("b"); prop != nil {
		if !prop.IsBoolean() || !prop.IsFalse() {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if elem := result.GetElement(0); elem != nil {
		if !elem.IsUndefined() {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if !result.SetElement(0, Default.True()) {
		t.FailNow()
	}

	if elem := result.GetElement(0); elem != nil {
		if !elem.IsTrue() {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	runtime.GC()
}

func Test_Array(t *testing.T) {
	context := Default.NewContext()
	script := context.Compile("[1,2,3]", nil, nil)
	value := script.Run(context)
	result := value.ToArray()

	if result.Length() != 3 {
		t.FailNow()
	}

	if elem := result.GetElement(0); elem != nil {
		if !elem.IsNumber() || elem.GetNumber() != 1 {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if elem := result.GetElement(1); elem != nil {
		if !elem.IsNumber() || elem.GetNumber() != 2 {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if elem := result.GetElement(2); elem != nil {
		if !elem.IsNumber() || elem.GetNumber() != 3 {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	if !result.SetElement(0, Default.True()) {
		t.FailNow()
	}

	if elem := result.GetElement(0); elem != nil {
		if !elem.IsTrue() {
			t.FailNow()
		}
	} else {
		t.FailNow()
	}

	runtime.GC()
}

func Test_TypeCheck(t *testing.T) {
	// TODO
}

func Test_SpecialValues(t *testing.T) {
	if !Default.Undefined().IsUndefined() {
		t.FailNow()
	}

	if !Default.Null().IsNull() {
		t.FailNow()
	}

	if !Default.True().IsTrue() {
		t.FailNow()
	}

	if !Default.False().IsFalse() {
		t.FailNow()
	}
}

func Test_UnderscoreJS(t *testing.T) {
	// Need download underscore.js from:
	// https://raw.github.com/jashkenas/underscore/master/underscore.js
	code, err := ioutil.ReadFile("underscore.js")

	if err != nil {
		return
	}

	context := Default.NewContext()
	script := context.Compile(string(code), nil, nil)
	script.Run(context)

	test := "_.find([1, 2, 3, 4, 5, 6], function(num){ return num % 2 == 0; });"
	testScript := context.Compile(test, nil, nil)
	value := testScript.Run(context)

	if value == nil || value.IsNumber() == false {
		t.FailNow()
	}

	result := value.GetNumber()

	if result != 2 {
		t.FailNow()
	}
}

func Test_ThreadSafe1(t *testing.T) {
	fail := false

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			context := Default.NewContext()
			script := context.Compile("'Hello ' + 'World!'", nil, nil)
			value := script.Run(context)
			result := value.ToString()
			fail = fail || result != "Hello World!"
			runtime.GC()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

func Test_ThreadSafe2(t *testing.T) {
	fail := false
	context := Default.NewContext()

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			rand_sched(200)

			script := context.Compile("'Hello ' + 'World!'", nil, nil)
			value := script.Run(context)
			result := value.ToString()
			fail = fail || result != "Hello World!"
			runtime.GC()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

func Test_ThreadSafe3(t *testing.T) {
	fail := false
	context := Default.NewContext()
	script := context.Compile("'Hello ' + 'World!'", nil, nil)

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			rand_sched(200)

			value := script.Run(context)
			result := value.ToString()
			fail = fail || result != "Hello World!"
			runtime.GC()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

func Test_ThreadSafe4(t *testing.T) {
	fail := false
	context := Default.NewContext()
	script := context.Compile("'Hello ' + 'World!'", nil, nil)
	value := script.Run(context)

	wg := new(sync.WaitGroup)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			rand_sched(200)

			result := value.ToString()
			fail = fail || result != "Hello World!"
			runtime.GC()
			wg.Done()
		}()
	}
	wg.Wait()
	runtime.GC()

	if fail {
		t.FailNow()
	}
}

func Test_ThreadSafe5(t *testing.T) {
	fail := false
	gonum := 100
	contextChan := make(chan *Context, gonum*2)
	scriptChan := make(chan *Script, gonum)
	valueChan := make(chan *Value, gonum)

	for i := 0; i < gonum*2; i++ {
		go func() {
			rand_sched(200)

			contextChan <- Default.NewContext()
		}()
	}

	for i := 0; i < gonum; i++ {
		go func() {
			rand_sched(200)

			context := <-contextChan
			scriptChan <- context.Compile("'Hello ' + 'World!'", nil, nil)
		}()
	}

	for i := 0; i < gonum; i++ {
		go func() {
			rand_sched(200)

			context := <-contextChan
			script := <-scriptChan
			valueChan <- script.Run(context)
		}()
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < gonum; i++ {
		wg.Add(1)
		go func() {
			rand_sched(200)

			value := <-valueChan
			result := value.ToString()
			fail = fail || result != "Hello World!"
			runtime.GC()
			wg.Done()
		}()
	}
	wg.Wait()

	runtime.GC()

	if fail {
		t.FailNow()
	}
}

func Benchmark_NewContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Default.NewContext()
	}

	runtime.GC()
}

func Benchmark_Compile(b *testing.B) {
	b.StartTimer()
	context := Default.NewContext()
	scripts := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		scripts[i] = `function myfunc(a, b) { 
			return 'Hello ' + '` + strconv.Itoa(i) + `' + a + b
		}`
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		context.Compile(scripts[i], nil, nil)
	}

	runtime.GC()
}

func Benchmark_PreCompile(b *testing.B) {
	b.StartTimer()
	context := Default.NewContext()
	scripts := make([]string, b.N)
	scriptDatas := make([]*ScriptData, b.N)
	for i := 0; i < b.N; i++ {
		scripts[i] = `function myfunc(a, b) { 
			return 'Hello ' + '` + strconv.Itoa(i) + `' + a + b
		}`
		scriptDatas[i] = Default.PreCompile(scripts[i])
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		context.Compile(scripts[i], nil, scriptDatas[i])
	}

	runtime.GC()
}

func Benchmark_RunScript(b *testing.B) {
	b.StartTimer()
	context := Default.NewContext()
	script := context.Compile("'Hello ' + 'World!'", nil, nil)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		script.Run(context)
	}

	runtime.GC()
}
