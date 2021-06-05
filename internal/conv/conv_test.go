/**
 * @Author: wanglin
 * @Email: wanglin@vspn.com
 * @Date: 2021/6/5 11:37 上午
 * @Desc: TODO
 */

package conv_test

import (
	"fmt"
	"testing"
	
	"github.com/dobyte/cache/internal/conv"
)

func TestBytesToString(t *testing.T) {
	b := []byte("abcdefg")
	
	fmt.Println(b)
	
	fmt.Println(conv.String(b))
	//fmt.Println(conv.BytesToString(b))
}
