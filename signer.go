package env

import "strings"

var _ Signer = &signer{}

type signer struct {
	inner
	prefix   string
	category string
	environ  *environ
}

func newSigner(prefix, category string, environ *environ) Signer {
	s := &signer{
		prefix:   prefix,
		category: category,
		environ:  environ,
	}
	s.inner.lookup = s.lookup
	s.inner.exists = s.exists
	s.inner.iter = s.iter
	return s
}

func (s *signer) lookup(key string) (string, bool) {
	// 相当于使用 prefix 作为分组，category 表示不同类目，
	// 最终形成 prefix_category_key 这样的数据键名称
	value, exists := s.lookup2(s.category, key)
	if exists || s.category == "" {
		return value, exists
	}
	// 当无法通过类目来查找数据时，我们
	// 使用 prefix_key 作为缺省值来查找数据
	return s.lookup2("", key)
}

func (s *signer) lookup2(category, key string) (string, bool) {
	if category != "" {
		key = category + "_" + key
	}
	if s.prefix != "" {
		key = s.prefix + "_" + key
	}
	return s.environ.Lookup(key)
}

func (s *signer) exists(key string) bool {
	// 相当于使用 prefix 作为分组，category 表示不同类目，
	// 最终形成 prefix_category_key 这样的数据键名称
	exists := s.exists2(s.category, key)
	if exists || s.category == "" {
		return exists
	}
	// 当无法通过类目来确定数据是否存在时，我们
	// 使用 prefix_key 作为缺省值来确定数据是否存在
	return s.exists2("", key)
}

func (s *signer) exists2(category, key string) bool {
	if category != "" {
		key = category + "_" + key
	}
	if s.prefix != "" {
		key = s.prefix + "_" + key
	}
	return s.environ.Exists(key)
}

func (s *signer) iter() func() (key string, value string, ok bool) {
	next := s.environ.inner.iter()
	prefix := s.prefix
	if prefix != "" {
		prefix += "_"
	}
	if s.category != "" {
		prefix += s.category + "_"
	}
	var keys, values []string
	var index int
	return func() (key string, value string, ok bool) {
		if next == nil {
			if index >= len(keys) {
				return "", "", false
			}
			defer func() {
				index++
			}()
			return keys[index], values[index], true
		}
		for {
			k, v, b := next()
			if !b {
				return "", "", false
			}
			if strings.HasPrefix(k, prefix) {
				return strings.TrimPrefix(k, prefix), v, true
			}
			if s.prefix != "" && strings.HasPrefix(k, s.prefix) {
				keys = append(keys, strings.TrimPrefix(k, s.prefix))
				values = append(values, value)
			}
		}
	}
}
