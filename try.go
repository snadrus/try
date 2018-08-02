package try

import (
	"bytes"
	"fmt"
	"runtime"
)

type Except struct {
	deprecations map[string]struct{}
	callstack    string
}

type RichError struct {
	error
	Callstack string
}

func (r RichError) ErrorWithCallstack() string {
	return r.Error() + " " + r.Callstack
}

func (h *Except) PanicIf(err error) {
	if err != nil {
		h.callstack = getStack()
		panic(err)
	}
}

func (h *Except) ImDeprecated() {
	h.deprecations[getStack()] = struct{}{}
}

func getStack() string {
	s := make([]byte, 0, 1024*10)
	runtime.Stack(s, true)
	return bytes.NewBuffer(s).String()
}

// Try/Catch convenient panics.
// Your functions can take ErrorHook
// It provides ErrorHook which should not have any references taken to it (should not outlive variable lifetime)
func Try(f func(*Except), deprecationHandler func(string)) (rich RichError) {
	err := &Except{deprecations: map[string]struct{}{}}
	defer func() {
		if v := recover(); v != nil {
			e, ok := v.(error)
			if !ok {
				e = fmt.Errorf("%v", v)
			}
			if err.callstack == "" {
				err.callstack = getStack()
			}
			rich = RichError{
				error:     e,
				Callstack: err.callstack,
			}
		}
		if deprecationHandler != nil {
			var b *bytes.Buffer
			for k := range err.deprecations {
				if b == nil {
					b = bytes.NewBufferString(k)
				} else {
					b.WriteString(k)
				}
			}
			s := ""
			if b != nil {
				s = b.String()
			}
			deprecationHandler(s)
		}
	}()
	f(err)
	return
}
