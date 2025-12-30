package rock

import (
	"reflect"
	"sync"
)

type (
	ValueSetter interface {
		Set(key string, newValue interface{}) (Entry, bool)
	}
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

// 初始化Store以适配现有代码
func (s *Store) init() {
	if s.entries == nil {
		s.entries = make([]Entry, 0)
	}
	if s.index == nil {
		s.index = make(map[string]int)
	}
}

func (r *Store) Set(key string, value interface{}) (Entry, bool) {
	return r.Save(key, value, false)
}

func (r *Store) Save(key string, value interface{}, immutable bool) (Entry, bool) {
	r.init()
	r.mu.Lock()
	defer r.mu.Unlock()
	
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
	r.init()
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	v, ok := r.GetEntry(key)
	if !ok || v.ValueRaw == nil {
		return def
	}
	vv := v.Value()
	if vv == nil {
		return def
	}
	return vv
}

// Value returns the value of the entry,
// respects the immutable.
func (e Entry) Value() interface{} {
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

// GetEntry returns a pointer to the "Entry" found with the given "key"
// if nothing found then it returns an empty Entry and false.
func (r *Store) GetEntry(key string) (Entry, bool) {
	r.init()
	
	// 使用索引快速查找
	if idx, exists := r.index[key]; exists && idx < len(r.entries) {
		return r.entries[idx], true
	}
	
	return emptyEntry, false
}
