package club

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/hcl"
)

//-----------------------------------------------------------------------------

var (
	errNotAvailable = errors.New("N/A")
)

// Here .
func Here(skip ...int) (funcName, fileName string, fileLine int, callerErr error) {
	sk := 1
	if len(skip) > 0 && skip[0] > 1 {
		sk = skip[0]
	}
	var pc uintptr
	var ok bool
	pc, fileName, fileLine, ok = runtime.Caller(sk)
	if !ok {
		callerErr = errNotAvailable
		return
	}
	fn := runtime.FuncForPC(pc)
	name := fn.Name()
	ix := strings.LastIndex(name, ".")
	if ix > 0 && (ix+1) < len(name) {
		name = name[ix+1:]
	}
	funcName = name
	nd, nf := filepath.Split(fileName)
	fileName = filepath.Join(filepath.Base(nd), nf)
	return
}

//-----------------------------------------------------------------------------

// TimerScope .
func TimerScope(name string, opCount ...int) func() {
	if name == "" {
		funcName, fileName, fileLine, err := Here(2)
		if err != nil {
			name = "N/A"
		} else {
			name = fmt.Sprintf("%s %02d %s()", fileName, fileLine, funcName)
		}
	}
	log.Println(name, `started`)
	start := time.Now()
	return func() {
		elapsed := time.Now().Sub(start)
		log.Printf("%s took %v", name, elapsed)
		if len(opCount) == 0 {
			return
		}

		N := opCount[0]
		if N <= 0 {
			return
		}

		E := float64(elapsed)
		FRC := E / float64(N)

		log.Printf("op/sec %.2f", float64(N)/(E/float64(time.Second)))

		switch {
		case FRC > float64(time.Second):
			log.Printf("sec/op %.2f", (E/float64(time.Second))/float64(N))
		case FRC > float64(time.Millisecond):
			log.Printf("milli-sec/op %.2f", (E/float64(time.Millisecond))/float64(N))
		case FRC > float64(time.Microsecond):
			log.Printf("micro-sec/op %.2f", (E/float64(time.Microsecond))/float64(N))
		default:
			log.Printf("nano-sec/op %.2f", (E/float64(time.Nanosecond))/float64(N))
		}
	}
}

//-----------------------------------------------------------------------------

type _err string

func (v _err) Error() string { return string(v) }

// Errorf value type (string) error
func Errorf(format string, a ...interface{}) error {
	return _err(fmt.Sprintf(format, a...))
}

//-----------------------------------------------------------------------------

// PolishYeKaf .
func PolishYeKaf(s string) (res string) {
	res = s
	res = strings.Replace(
		res,
		"ي",
		"ی",
		-1)
	res = strings.Replace(
		res,
		"ك",
		"ک",
		-1)

	return
}

// IranTime .
func IranTime(source time.Time) time.Time {
	var dest time.Time
	loc, err := time.LoadLocation(`Asia/Tehran`)
	if err == nil {
		dest = source.In(loc)
	} else {
		dest = source
	}
	return dest
}

// IranNow .
func IranNow() time.Time {
	return IranTime(time.Now())
}

//-----------------------------------------------------------------------------

// OnSignal .
func OnSignal(f func(), sig ...os.Signal) {
	if f == nil {
		return
	}
	sigc := make(chan os.Signal, 1)
	if len(sig) > 0 {
		signal.Notify(sigc, sig...)
	} else {
		signal.Notify(sigc,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGSTOP,
			syscall.SIGABRT,
			syscall.SIGTSTP,
			syscall.SIGKILL)
	}
	go func() {
		<-sigc
		f()
	}()
}

//-----------------------------------------------------------------------------

// Supervise take care of restarts (in case of panic or error)
func Supervise(action func() error, intensity int, period ...time.Duration) {
	if intensity == 0 {
		return
	}
	if intensity > 0 {
		intensity--
	}
	dt := time.Second * 3
	if len(period) > 0 && period[0] > 0 {
		dt = period[0]
	}
	retry := func() {
		time.Sleep(dt)
		go Supervise(action, intensity, dt)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			retry()
		}
	}()
	if err := action(); err != nil {
		log.Println(err)
		retry()
	}
}

//-----------------------------------------------------------------------------

func defaultAppNameHandler() string {
	return filepath.Base(os.Args[0])
}

func defaultConfNameHandler() string {
	fp := fmt.Sprintf("%s.conf", defaultAppNameHandler())
	if _, err := os.Stat(fp); err != nil {
		fp = "app.conf"
	}
	return fp
}

// LoadHCL loads hcl conf file. default conf file names (if filePath not provided)
// in the same directory are <appname>.conf and if not fount app.conf
func LoadHCL(ptr interface{}, filePath ...string) error {
	var fp string
	if len(filePath) > 0 {
		fp = filePath[0]
	}
	if fp == "" {
		fp = defaultConfNameHandler()
	}
	cn, err := ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	err = hcl.Unmarshal(cn, ptr)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

var (
	ぁappCtx    context.Context
	ぁappCancel context.CancelFunc
	ぁappWG     = &sync.WaitGroup{}
)

func init() {
	ぁappCtx, ぁappCancel = context.WithCancel(context.Background())
	OnSignal(func() { ぁappCancel() })
}

// Ctx global context
func Ctx() context.Context { return ぁappCtx }

// WaitGroupScope use as
//	defer WaitGroupScope()()
func WaitGroupScope() func() {
	ぁappWG.Add(1)
	return func() { ぁappWG.Done() }
}

//-----------------------------------------------------------------------------

// Finit .
func Finit(timeout time.Duration, cancelApp ...bool) {
	if len(cancelApp) > 0 && cancelApp[0] {
		ぁappCancel()
	}
	<-ぁappCtx.Done()

	done := make(chan struct{})
	go func() {
		defer close(done)
		ぁappWG.Wait()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		log.Println(fmt.Errorf("TIMEOUT"))
	}
}

//-----------------------------------------------------------------------------

// import (
// 	"errors"
// 	"fmt"
// 	"log"
// 	"path/filepath"
// 	"runtime"
// 	"strings"
// 	"time"
// )

// //-----------------------------------------------------------------------------

// // Here .
// func Here(skip ...int) (funcName, fileName string, fileLine int, callerErr error) {
// 	sk := 1
// 	if len(skip) > 0 && skip[0] > 1 {
// 		sk = skip[0]
// 	}
// 	var pc uintptr
// 	var ok bool
// 	pc, fileName, fileLine, ok = runtime.Caller(sk)
// 	if !ok {
// 		callerErr = errNotAvailable
// 		return
// 	}
// 	fn := runtime.FuncForPC(pc)
// 	name := fn.Name()
// 	ix := strings.LastIndex(name, ".")
// 	if ix > 0 && (ix+1) < len(name) {
// 		name = name[ix+1:]
// 	}
// 	funcName = name
// 	nd, nf := filepath.Split(fileName)
// 	fileName = filepath.Join(filepath.Base(nd), nf)
// 	return
// }

// //-----------------------------------------------------------------------------

// // TimerScope .
// func TimerScope(name string, opCount ...int) func() {
// 	if name == "" {
// 		funcName, fileName, fileLine, err := Here(2)
// 		if err != nil {
// 			name = "N/A"
// 		} else {
// 			name = fmt.Sprintf("%s %02d %s()", fileName, fileLine, funcName)
// 		}
// 	}
// 	log.Println(name, `started`)
// 	start := time.Now()
// 	return func() {
// 		elapsed := time.Now().Sub(start)
// 		log.Printf("%s took %v", name, elapsed)
// 		if len(opCount) == 0 {
// 			return
// 		}

// 		N := opCount[0]
// 		if N <= 0 {
// 			return
// 		}

// 		E := float64(elapsed)
// 		FRC := E / float64(N)

// 		log.Printf("op/sec %.2f", float64(N)/(E/float64(time.Second)))

// 		switch {
// 		case FRC > float64(time.Second):
// 			log.Printf("sec/op %.2f", (E/float64(time.Second))/float64(N))
// 		case FRC > float64(time.Millisecond):
// 			log.Printf("milli-sec/op %.2f", (E/float64(time.Millisecond))/float64(N))
// 		case FRC > float64(time.Microsecond):
// 			log.Printf("micro-sec/op %.2f", (E/float64(time.Microsecond))/float64(N))
// 		default:
// 			log.Printf("nano-sec/op %.2f", (E/float64(time.Nanosecond))/float64(N))
// 		}
// 	}
// }

// //-----------------------------------------------------------------------------

// // import (
// // 	"context"
// // 	"fmt"
// // 	"log"
// // 	"os"
// // 	"os/signal"
// // 	"runtime"
// // 	"syscall"
// // 	"time"

// // 	"github.com/dc0d/goroutines"
// // )

// // //-----------------------------------------------------------------------------

// // func init() {
// // 	AppCtx, AppCancel = context.WithCancel(context.Background())
// // 	CallOnSignal(func() { AppCancel() })
// // 	AppPool = NewGroup()
// // }

// // //-----------------------------------------------------------------------------

// // // Finit use to prevent app from stopping/exiting
// // func Finit(timeout time.Duration, nap ...time.Duration) {
// // 	<-AppCtx.Done()
// // 	werr := goroutines.New().
// // 		WaitStart().
// // 		WaitGo(timeout).
// // 		Go(func() {
// // 			AppPool.Wait()
// // 		})
// // 	if werr != nil {
// // 		log.Println(`error:`, werr)
// // 	}
// // 	s := time.Millisecond * 300
// // 	if len(nap) > 0 {
// // 		s = nap[0]
// // 	}
// // 	if s <= 0 {
// // 		s = time.Millisecond * 300
// // 	}
// // 	<-time.After(s)
// // }

// // //-----------------------------------------------------------------------------

// // // Recover recovers from panic and returns error,
// // // or returns the provided error
// // func Recover(f func() error, verbose ...bool) (err error) {
// // 	vrb := false
// // 	if len(verbose) > 0 {
// // 		vrb = verbose[0]
// // 	}

// // 	defer func() {
// // 		if e := recover(); e != nil {
// // 			if !vrb {
// // 				err = Error(fmt.Sprintf("%v", e))
// // 				return
// // 			}

// // 			trace := make([]byte, 1<<16)
// // 			n := runtime.Stack(trace, true)

// // 			s := fmt.Sprintf("error: %v, n: %d, traced: %s", e, n, trace[:n])
// // 			err = Error(s)
// // 		}
// // 	}()

// // 	err = f()

// // 	return
// // }

// // //-----------------------------------------------------------------------------

// // // CallOnSignal calls function on os signal
// // func CallOnSignal(f func(), sig ...os.Signal) {
// // 	if f == nil {
// // 		return
// // 	}
// // 	sigc := make(chan os.Signal, 1)
// // 	if len(sig) > 0 {
// // 		signal.Notify(sigc, sig...)
// // 	} else {
// // 		signal.Notify(sigc,
// // 			syscall.SIGINT,
// // 			syscall.SIGTERM,
// // 			syscall.SIGQUIT,
// // 			syscall.SIGSTOP,
// // 			syscall.SIGABRT,
// // 			syscall.SIGTSTP,
// // 			syscall.SIGKILL)
// // 	}
// // 	go func() {
// // 		<-sigc
// // 		f()
// // 	}()
// // }

// // //-----------------------------------------------------------------------------
