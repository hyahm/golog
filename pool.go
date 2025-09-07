// Copyright (c) 2023 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package pool provides internal pool utilities.
package golog

// A Pool is a generic wrapper around [sync.Pool] to provide strongly-typed
// object pooling.
//
// Note that SA6002 (ref: https://staticcheck.io/docs/checks/#SA6002) will
// not be detected, so all internal pool use must take care to only store
// pointer types.
// var pool chan *msgLog

// const POOLCOUNT = 1500

// var count int32

// func init() {
// 	// 最大1500个内存池
// 	pool = make(chan *msgLog, POOLCOUNT)
// }

// func GetPool() *msgLog {
// 	ticker := time.NewTicker(1 * time.Microsecond * 10)

// 	defer ticker.Stop() // 主函数退出前停止 Ticker，防止 goroutine 泄漏

// 	for {
// 		select {
// 		case ml := <-pool:
// 			atomic.AddInt32(&count, -1)
// 			return ml
// 		case <-ticker.C:
// 			return &msgLog{}
// 		}
// 	}
// }

// func PutPool(ml *msgLog) {
// 	// 如果缓存满了怎么办
// 	select {
// 	case pool <- ml:
// 		atomic.AddInt32(&count, 1)
// 	default:
// 		// fmt.Println("pool is full--------", atomic.LoadInt32(&count))
// 		// 处理 pool 已满的情况，例如：
// 		// 1. 丢弃对象
// 		// 2. 记录日志
// 		// 3. 尝试其他存储方式
// 	}
// }
