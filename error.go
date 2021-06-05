/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/6/1 1:01 下午
 * @Desc: TODO
 */

package cache

const Nil = StoreError("store: nil")

type StoreError string

func (e StoreError) Error() string { return string(e) }

func (StoreError) StoreError() {}
