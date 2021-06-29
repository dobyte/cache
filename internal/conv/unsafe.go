/**
 * @Author: wanglin
 * @Author: wanglin@vspn.com
 * @Date: 2021/6/29 19:51
 * @Desc: TODO
 */

package conv

import "unsafe"

func UnsafeStringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}

func UnsafeBytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
