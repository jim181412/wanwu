package rel

type Res struct {
	ID  string
	Typ ResType
}

// ResType 资源类型 (用户自定义)
type ResType int32

// ResOp 资源操作 (用户自定义)
type ResOp string

// 预定义一些常用操作
const (
	OpDelete ResOp = "DELETE" // 如果希望某种资源数据可以从depend系统中被删除，则至少需要对该资源的某个状态注册DELETE操作
	OpEdit   ResOp = "EDIT"
)

// Type 依赖方式 (用户自定义)
type Type int32

// 预定义一些常用依赖方式
const (
	TypeDefault Type = 0 // 通常资源A只以default一种方式被资源B依赖；当资源A可以被资源B以多种方式依赖时，需要用户自定义
)
