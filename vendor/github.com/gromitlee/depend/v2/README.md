# depend

Resource Depend and Access Management

- [x] 支持完全由用户自定义的资源、资源类型(`ResType`)、资源操作(`ResOp`)、依赖方式(`Type`)
- [x] 提供两种使用方式：全局单例方式与管理器方式
- [x] 提供基于[gorm](https://github.com/go-gorm/gorm)的数据存储
- [x] Goroutine Safe & Developer Friendly

## 安装

```bash
go get github.com/gromitlee/depend/v2
```

## 引入

```go
import "github.com/gromitlee/depend/v2"
```

## 使用方式1：全局单例
API详见[api.go](api.go)

示例详见[examples/depend_test.go](examples/depend_test.go)

```go
depend.Init(db)
```


## 使用方式2：管理器
API详见[mgr.go](mgr.go)

示例同上

```go
mgr, err := NewDependMgr(db)
```