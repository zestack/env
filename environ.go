package env

import (
	"sync"
	"sync/atomic"

	"github.com/joho/godotenv"
)

var _ Signer = &environ{}

type environ struct {
	inner
	keys   []string
	values []string
	mu     sync.RWMutex
}

func New() Environ {
	e := &environ{}
	e.inner.lookup = e.lookup
	e.inner.exists = e.exists
	e.inner.iter = e.iter
	return e
}

// Load 加载环境变量文件
func (e *environ) Load(filenames ...string) error {
	data, err := godotenv.Read(filenames...)
	if err == nil {
		e.Save(data)
	}
	return err
}

// Save 保存数据到缓存的环境变量里面
func (e *environ) Save(data map[string]string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for key, value := range data {
		if i := e.index(key); i > -1 {
			e.values[i] = value
		} else {
			e.keys = append(e.keys, key)
			e.values = append(e.values, value)
		}
	}
}

func (e *environ) Signed(prefix, category string) Signer {
	return newSigner(prefix, category, e)
}

func (e *environ) index(key string) int {
	if e.keys != nil {
		for i, s := range e.keys {
			if s == key {
				return i
			}
		}
	}
	return -1
}

// 查看环境变量值，如果不存在或值为空，返回的第二个参数的值则为false。
func (e *environ) lookup(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if i := e.index(key); i > -1 {
		v := e.values[i]
		return v, len(v) > 0
	}
	return "", false
}

// 判断环境变量是否存在
func (e *environ) exists(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.index(key) > -1
}

func (e *environ) iter() func() (key string, value string, ok bool) {
	var pos int32 = -1
	return func() (key string, value string, ok bool) {
		index := int(atomic.AddInt32(&pos, 1))
		if index >= len(e.keys) {
			return "", "", false
		}
		e.mu.RLock()
		defer e.mu.RUnlock()
		return e.keys[index], e.values[index], true
	}
}

func (e *environ) Clean() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.keys = nil
	e.values = nil
}
