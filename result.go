/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/24 9:43 上午
 * @Desc: result interface define
 */

package cache

import (
	"strconv"
	"time"
	
	"github.com/dobyte/cache/internal/conv"
)

type Result interface {
	// Err Return a error from result.
	Err() error
	// String Return a value of type string from the result.
	String() string
	// Val Return a value of type string from the result.
	Val() string
	// Result Return a value of type string and error from the result.
	Result() (string, error)
	// Bytes Return a value of type byte and error from the result.
	Bytes() ([]byte, error)
	// Bool Return a value of type bool and error from the result.
	Bool() (bool, error)
	// Int Return a value of type int and error from the result.
	Int() (int, error)
	// Int64 Return a value of type int64 and error from the result.
	Int64() (int64, error)
	// Uint64 Return a value of type uint64 and error from the result.
	Uint64() (uint64, error)
	// Float32 Return a value of type float32 and error from the result.
	Float32() (float32, error)
	// Float64 Return a value of type float64 and error from the result.
	Float64() (float64, error)
	// Time Return a value of type time and error from the result.
	Time() (time.Time, error)
	// Scan Convert the value from the result into a complex data structure.
	Scan(val interface{}) error
}

type result struct {
	err      error
	writeErr error
	val      string
}

func NewResult(val string, errs ...error) Result {
	r := new(result)
	r.val = val
	
	if len(errs) > 0 {
		r.err = errs[0]
	}
	
	if len(errs) > 1 {
		r.writeErr = errs[1]
	}
	
	return r
}

// Err Return a error from result.
func (r *result) Err() error {
	return r.err
}

// String Return a value of type string from the result.
func (r *result) String() string {
	return r.val
}

// Val Return a value of type string from the result.
func (r *result) Val() string {
	return r.val
}

// Result Return a value of type string and error from the result.
func (r *result) Result() (string, error) {
	return r.Val(), r.err
}

// Bytes Return a value of type []byte and error from the result.
func (r *result) Bytes() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return conv.UnsafeStringToBytes(r.Val()), nil
}

// Bool Return a value of type bool and error from the result.
func (r *result) Bool() (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	return strconv.ParseBool(r.Val())
}

// Int Return a value of type int and error from the result.
func (r *result) Int() (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return strconv.Atoi(r.Val())
}

// Int64 Return a value of type int64 and error from the result.
func (r *result) Int64() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return strconv.ParseInt(r.Val(), 10, 64)
}

// Uint64 Return a value of type uint64 and error from the result.
func (r *result) Uint64() (uint64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return strconv.ParseUint(r.Val(), 10, 64)
}

// Float32 Return a value of type float32 and error from the result.
func (r *result) Float32() (float32, error) {
	if r.err != nil {
		return 0, r.err
	}
	f, err := strconv.ParseFloat(r.Val(), 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

// Float64 Return a value of type float64 and error from the result.
func (r *result) Float64() (float64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return strconv.ParseFloat(r.Val(), 64)
}

// Time Return a value of type time and error from the result.
func (r *result) Time() (time.Time, error) {
	if r.err != nil {
		return time.Time{}, r.err
	}
	return time.Parse(time.RFC3339Nano, r.Val())
}

// Scan Convert the value from the result into a complex data structure.
func (r *result) Scan(val interface{}) error {
	if r.err != nil {
		return r.err
	}
	
	return conv.Scan([]byte(r.Val()), val)
}