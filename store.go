package rock

import "reflect"

type (
	ValueSetter interface {
		Set(key string, newValue interface{}) (Entry, bool)
	}
	Entry struct {
		Key       string      `json:"key" msgpack:"key" yaml:"Key" toml:"Value"`
		ValueRaw  interface{} `json:"value" msgpack:"value" yaml:"Value" toml:"Value"`
		immutable bool        // if true then it can't change by its caller.
	}
	Store []Entry
)

func (r *Store) Set(key string, value interface{}) (Entry, bool) {
	return r.Save(key, value, false)
}

func (r *Store) Save(key string, value interface{}, immutable bool) (Entry, bool) {
	args := *r
	// n := len(args)
	// add
	kv := Entry{
		Key:       key,
		ValueRaw:  value,
		immutable: immutable,
	}
	*r = append(args, kv)
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
	args := *r
	n := len(args)
	for i := 0; i < n; i++ {
		if kv := args[i]; kv.Key == key {
			return kv, true
		}
	}

	return emptyEntry, false
}
