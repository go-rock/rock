package rock

import (
	"fmt"
	"sync"
	"testing"
)

// 验证 Store 在高并发读写下没有数据竞态（配合 go test -race）。
func TestStoreConcurrent(t *testing.T) {
	store := NewStore()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("k%d", id)
			for j := 0; j < 100; j++ {
				store.Set(key, j)
				store.Get(key)
				store.GetDefault(key, -1)
				store.Save(key+"_imm", []string{"a", "b"}, true)
				_, _ = store.GetEntry(key)
			}
		}(i)
	}
	wg.Wait()
}

// 零值 Store（如 Ctx.values = Store{} 之后）在并发读写时也不应 panic。
func TestStoreZeroValueConcurrent(t *testing.T) {
	var store Store // 零值：entries/index 为 nil

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("z%d", id)
			for j := 0; j < 100; j++ {
				store.Set(key, j)
				_ = store.Get(key)
				_, _ = store.GetEntry(key)
			}
		}(i)
	}
	wg.Wait()
}
