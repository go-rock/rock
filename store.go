package rock

import (
	"reflect"
	"sync"
)

type (
	// ValueSetter 表示可以写入键值条的接口。
	ValueSetter interface {
		Set(key string, newValue interface{}) (Entry, bool)
	}

	// Entry 是 Store 中的一条键值记录。
	// immutable 为 true 时，读取会返回值的深拷贝，外部修改不影响存储。
	Entry struct {
		Key       string      `json:"key" msgpack:"key" yaml:"Key" toml:"Value"`
		ValueRaw  interface{} `json:"value" msgpack:"value" yaml:"Value" toml:"Value"`
		immutable bool        // if true then it can't change by its caller.
	}

	// Store 保持与现有代码的兼容性，但添加性能优化
	Store struct {
		entries []Entry
		index   map[string]int // 添加索引以提高查找性能
		mu      sync.RWMutex   // 添加互斥锁以支持并发访问
	}
)

// 创建新的Store实例
func NewStore() *Store {
	return &Store{
		entries: make([]Entry, 0),
		index:   make(map[string]int),
	}
}

// initLocked 初始化底层数据结构，调用方需持有写锁。
func (r *Store) initLocked() {
	if r.entries == nil {
		r.entries = make([]Entry, 0)
	}
	if r.index == nil {
		r.index = make(map[string]int)
	}
}

func (r *Store) Set(key string, value interface{}) (Entry, bool) {
	return r.Save(key, value, false)
}

// Save 保存键值条目；immutable 为 true 时读取返回深拷贝。
func (r *Store) Save(key string, value interface{}, immutable bool) (Entry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.initLocked()

	// 检查key是否已存在
	if idx, exists := r.index[key]; exists {
		// 更新现有条目
		r.entries[idx] = Entry{
			Key:       key,
			ValueRaw:  value,
			immutable: immutable,
		}
		return r.entries[idx], true
	}

	// 添加新条目
	kv := Entry{
		Key:       key,
		ValueRaw:  value,
		immutable: immutable,
	}
	idx := len(r.entries)
	r.entries = append(r.entries, kv)
	r.index[key] = idx
	return kv, true
}

// Get returns the entry's value based on its key.
// If not found returns nil.
func (r *Store) Get(key string) interface{} {
	return r.GetDefault(key, nil)
}

// GetDefault returns the entry's value based on its key.
// If not found returns "def".
// This function checks for immutability as well, the rest don't.
func (r *Store) GetDefault(key string, def interface{}) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.getEntryLocked(key)
	if !ok || v.ValueRaw == nil {
		return def
	}
	vv := v.Value()
	if vv == nil {
		return def
	}
	return vv
}

// Value 返回条目的值；对 immutable 条目返回深拷贝。
func (e Entry) Value() interface{} {
	if e.ValueRaw == nil {
		return nil
	}
	if e.immutable {
		// take its value, no pointer even if set with a reference.
		vv := reflect.Indirect(reflect.ValueOf(e.ValueRaw))

		// return copy of that slice
		if vv.Type().Kind() == reflect.Slice {
			newSlice := reflect.MakeSlice(vv.Type(), vv.Len(), vv.Cap())
			reflect.Copy(newSlice, vv)
			return newSlice.Interface()
		}
		// return a copy of that map
		if vv.Type().Kind() == reflect.Map {
			newMap := reflect.MakeMap(vv.Type())
			for _, k := range vv.MapKeys() {
				newMap.SetMapIndex(k, vv.MapIndex(k))
			}
			return newMap.Interface()
		}
		// if was *value it will return value{}.
		return vv.Interface()
	}
	return e.ValueRaw
}

var emptyEntry Entry

// getEntryLocked 按 key 查找条目，调用方需持有读锁。
// 对 nil 的 index/entries 读取是安全的（返回零值）。
func (r *Store) getEntryLocked(key string) (Entry, bool) {
	// 使用索引快速查找
	if idx, exists := r.index[key]; exists && idx < len(r.entries) {
		return r.entries[idx], true
	}

	return emptyEntry, false
}

// GetEntry 线程安全地返回 key 对应的条目副本。
// 如果没有找到则返回空的 Entry 和 false。
func (r *Store) GetEntry(key string) (Entry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.getEntryLocked(key)
}
