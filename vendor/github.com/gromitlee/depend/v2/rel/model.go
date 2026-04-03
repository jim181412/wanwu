package rel

// Relation 资源A 被 资源B 依赖
type Relation struct {
	ID int64 `gorm:"primary_key"`
	// 资源A id
	IDA string `gorm:"index:idx_relation_id_a;uniqueIndex:idx_depend;not null"`
	// 资源A 类型
	TypA ResType `gorm:"index:idx_relation_typ_a;uniqueIndex:idx_depend;not null"`
	// 依赖方式
	RelTyp Type `gorm:"index:idx_relation_rel_typ;uniqueIndex:idx_depend;not null"`
	// 资源B id
	IDB string `gorm:"index:idx_relation_id_b;uniqueIndex:idx_depend;not null"`
	// 资源B 类型
	TypB ResType `gorm:"index:idx_relation_typ_b;uniqueIndex:idx_depend;not null"`
}
