package env

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Signer interface {
	Lookup(key string) (string, bool)
	Exists(key string) bool
	String(key string, fallback ...string) string
	Bytes(key string, fallback ...[]byte) []byte
	Int(key string, fallback ...int) int
	Duration(key string, fallback ...time.Duration) time.Duration
	Bool(key string, fallback ...bool) bool
	List(key string, fallback ...[]string) []string
	Map(prefix string) map[string]string
	Where(filter func(name, value string) bool) map[string]string
	Fill(structure any) error
}

type Environ interface {
	Signer
	Load(filenames ...string) error
	Signed(prefix, category string) Signer
	Clean()
}

var (
	// 缓存的环境变量
	env = New().(*environ)
	// 环境变量文件 `.env` 所处的目录
	// 一般位于程序的工作目录
	root string
)

// Init 加载运行目录下的 .env 文件
func Init(root ...string) error {
	var dir string
	if len(root) > 0 {
		dir = root[0]
	}
	if dir == "" {
		dir = "."
	}
	return InitWithDir(dir)
}

// InitWithDir 加载指定录下的 .env 文件
func InitWithDir(dir string) (err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			root = ""
			env.Clean()
		} else {
			root = dir
		}
	}()

	// 重置缓存的环境变量
	root = ""
	env.Clean()

	// 加载系统的环境变量
	result := make(map[string]string)
	for _, value := range os.Environ() {
		parts := strings.SplitN(value, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		result[key] = val
	}
	env.Save(result)

	// 加载 .env 和 .env.local 文件
	err = loadEnv(dir, "")
	if err != nil {
		return err
	}

	// 加载与运行环境相关的环境变量
	appEnv := String("APP_ENV", "prod")
	if len(appEnv) > 0 {
		// 加载 .env.{APP_ENV} 和 .env.{APP_ENV}.local 文件
		err = loadEnv(dir, "."+strings.ToLower(appEnv))
		if err != nil {
			return err
		}
	}

	return
}

func loadEnv(dir, env string) error {
	filename := filepath.Join(dir, ".env"+env)
	if err := Load(filename); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if err := Load(filename + ".local"); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}

// Load 加载指定的环境变量文件
func Load(filenames ...string) error {
	return env.Load(filenames...)
}

func Signed(prefix, category string) Signer {
	return env.Signed(prefix, category)
}

// Path 基于初始化目录获取目录
func Path(path ...string) string {
	switch len(path) {
	case 0:
		return root
	case 1:
		return filepath.Join(root, path[0])
	default:
		return filepath.Join(root, filepath.Join(path...))
	}
}

// IsEnv 判断应用环境是否与给出的一致
func IsEnv(env string) bool {
	return String("APP_ENV") == env
}

// Lookup 查看配置
func Lookup(name string) (string, bool) {
	return env.Lookup(name)
}

// Exists 配置是否存在
func Exists(name string) bool {
	return env.Exists(name)
}

// String 取字符串值
func String(name string, value ...string) string {
	return env.String(name, value...)
}

// Bytes 取二进制值
func Bytes(name string, value ...[]byte) []byte {
	return env.Bytes(name, value...)
}

// Int 取整型值
func Int(name string, value ...int) int {
	return env.Int(name, value...)
}

func Duration(name string, value ...time.Duration) time.Duration {
	return env.Duration(name, value...)
}

func Bool(name string, value ...bool) bool {
	return env.Bool(name, value...)
}

// List 将值按 `,` 分割并返回
func List(name string, fallback ...[]string) []string {
	return env.List(name, fallback...)
}

func Map(prefix string) map[string]string {
	return env.Map(prefix)
}

func Where(filter func(name string, value string) bool) map[string]string {
	return env.Where(filter)
}

// Fill 将环境变量填充到指定结构体
func Fill(structure any) error {
	return env.Fill(structure)
}

// All 返回所有值
func All() map[string]string {
	return env.Where(func(name, value string) bool {
		return true
	})
}
