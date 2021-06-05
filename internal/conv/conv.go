/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/5/26 12:41 下午
 * @Desc: TODO
 */

package conv

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func String(any interface{}) string {
	switch v := any.(type) {
	case nil:
		return ""
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case time.Time:
		return v.String()
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.String()
	default:
		if v == nil {
			return ""
		}
		
		if i, ok := v.(stringInterface); ok {
			return i.String()
		}
		
		if i, ok := v.(errorInterface); ok {
			return i.Error()
		}
		
		var (
			rv   = reflect.ValueOf(v)
			kind = rv.Kind()
		)
		
		switch kind {
		case reflect.Chan,
			reflect.Map,
			reflect.Slice,
			reflect.Func,
			reflect.Ptr,
			reflect.Interface,
			reflect.UnsafePointer:
			if rv.IsNil() {
				return ""
			}
		case reflect.String:
			return rv.String()
		}
		
		if kind == reflect.Ptr {
			return String(rv.Elem().Interface())
		}
		
		if b, e := json.Marshal(v); e != nil {
			return fmt.Sprint(v)
		} else {
			return string(b)
		}
	}
}

func Bytes(any interface{}) []byte {
	if any == nil {
		return nil
	}
	switch v := any.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	default:
		return nil
	}
}

func Scan(b []byte, any interface{}) error {
	switch v := any.(type) {
	case nil:
		return fmt.Errorf("Scan(nil)")
	case *string:
		*v = String(b)
		return nil
	case *[]byte:
		*v = b
		return nil
	case *int:
		var err error
		*v, err = strconv.Atoi(String(b))
		return err
	case *int8:
		n, err := strconv.ParseInt(String(b), 10, 8)
		if err != nil {
			return err
		}
		*v = int8(n)
		return nil
	case *int16:
		n, err := strconv.ParseInt(String(b), 10, 16)
		if err != nil {
			return err
		}
		*v = int16(n)
		return nil
	case *int32:
		n, err := strconv.ParseInt(String(b), 10, 32)
		if err != nil {
			return err
		}
		*v = int32(n)
		return nil
	case *int64:
		n, err := strconv.ParseInt(String(b), 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *uint:
		n, err := strconv.ParseUint(String(b), 10, 64)
		if err != nil {
			return err
		}
		*v = uint(n)
		return nil
	case *uint8:
		n, err := strconv.ParseUint(String(b), 10, 8)
		if err != nil {
			return err
		}
		*v = uint8(n)
		return nil
	case *uint16:
		n, err := strconv.ParseUint(String(b), 10, 16)
		if err != nil {
			return err
		}
		*v = uint16(n)
		return nil
	case *uint32:
		n, err := strconv.ParseUint(String(b), 10, 32)
		if err != nil {
			return err
		}
		*v = uint32(n)
		return nil
	case *uint64:
		n, err := strconv.ParseUint(String(b), 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *float32:
		n, err := strconv.ParseFloat(String(b), 32)
		if err != nil {
			return err
		}
		*v = float32(n)
		return err
	case *float64:
		var err error
		*v, err = strconv.ParseFloat(String(b), 64)
		return err
	case *bool:
		*v = len(b) == 1 && b[0] == '1'
		return nil
	case *time.Time:
		var err error
		*v, err = time.Parse(time.RFC3339Nano, String(b))
		return err
	case encoding.BinaryUnmarshaler:
		return v.UnmarshalBinary(b)
	default:
		return fmt.Errorf(
			"can't unmarshal %T (consider implementing BinaryUnmarshaler)", v)
	}
}

//func Int(any interface{}) int {
//	if any == nil {
//		return 0
//	}
//	if v, ok := any.(int); ok {
//		return v
//	}
//	return int(Int64(any))
//}
//
//func Int8(any interface{}) int8 {
//	if any == nil {
//		return 0
//	}
//	if v, ok := any.(int8); ok {
//		return v
//	}
//	return int8(Int64(any))
//}
//
//func Int16(any interface{}) int16 {
//	if any == nil {
//		return 0
//	}
//	if v, ok := any.(int16); ok {
//		return v
//	}
//	return int16(Int64(any))
//}
//
//func Int32(any interface{}) int32 {
//	if any == nil {
//		return 0
//	}
//	if v, ok := any.(int32); ok {
//		return v
//	}
//	return int32(Int64(any))
//}
//
//func Int64(any interface{}) int64 {
//	if any == nil {
//		return 0
//	}
//
//	switch v := any.(type) {
//	case int:
//		return int64(v)
//	case int8:
//		return int64(v)
//	case int16:
//		return int64(v)
//	case int32:
//		return int64(v)
//	case int64:
//		return v
//	case uint:
//		return int64(v)
//	case uint8:
//		return int64(v)
//	case uint16:
//		return int64(v)
//	case uint32:
//		return int64(v)
//	case uint64:
//		return int64(v)
//	case bool:
//		if v {
//			return 1
//		}
//		return 0
//	case []byte:
//		return 0
//	default:
//		s := String(v)
//		isMinus := false
//		if len(s) > 0 {
//			if s[0] == '-' {
//				isMinus = true
//				s = s[1:]
//			} else if s[0] == '+' {
//				s = s[1:]
//			}
//		}
//		// Hexadecimal
//		if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
//			if v, e := strconv.ParseInt(s[2:], 16, 64); e == nil {
//				if isMinus {
//					return -v
//				}
//				return v
//			}
//		}
//		// Octal
//		if len(s) > 1 && s[0] == '0' {
//			if v, e := strconv.ParseInt(s[1:], 8, 64); e == nil {
//				if isMinus {
//					return -v
//				}
//				return v
//			}
//		}
//		// Decimal
//		if v, e := strconv.ParseInt(s, 10, 64); e == nil {
//			if isMinus {
//				return -v
//			}
//			return v
//		}
//		// Float64
//		return int64(Float64(v))
//	}
//}
