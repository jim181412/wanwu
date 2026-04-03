package depend

import (
	"errors"

	"github.com/gromitlee/depend/v2/rel"
	"gorm.io/gorm"
)

// 单例模式
var _mgr IDependMgr

func Init(db *gorm.DB) error {
	if _mgr != nil {
		return errors.New("depend mgr already init")
	}
	var err error
	_mgr, err = NewDependMgr(db)
	return err
}

// Register 表示typ类型的资源，在被ignore类型的资源依赖时，依然可以执行op操作
func Register(typ rel.ResType, op rel.ResOp, ignores []rel.ResType) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	return _mgr.Register(typ, op, ignores)
}

// AddRelation 记录关系：资源A 被 资源B 依赖
// relTyp长度不超过1；relType为空，指定默认依赖方式
func AddRelation(db *gorm.DB, resA rel.Res, resB rel.Res, relTyp ...rel.Type) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	_relTyp := rel.TypeDefault
	if len(relTyp) == 1 {
		_relTyp = relTyp[0]
	} else if len(relTyp) > 1 {
		return errors.New("len(relTyp) no more than 1")
	}
	return _mgr.AddRelation(db, resA.ID, resA.Typ, _relTyp, resB.ID, resB.Typ)
}

// AddRelations 记录关系：资源A 被 一组资源 依赖
// relTyp长度不超过1；relType为空，指定默认依赖方式
func AddRelations(db *gorm.DB, resA rel.Res, res []rel.Res, relTyp ...rel.Type) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	_relTyp := rel.TypeDefault
	if len(relTyp) == 1 {
		_relTyp = relTyp[0]
	} else if len(relTyp) > 1 {
		return errors.New("len(relTyp) no more than 1")
	}
	if len(res) == 0 {
		return errors.New("res empty")
	}
	return _mgr.AddRelations(db, resA.ID, resA.Typ, _relTyp, res)
}

// AddDepends 记录关系：资源B 依赖 一组资源
// relTyp长度不超过1；relType为空，指定默认依赖方式
func AddDepends(db *gorm.DB, res []rel.Res, resB rel.Res, relTyp ...rel.Type) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	_relTyp := rel.TypeDefault
	if len(relTyp) == 1 {
		_relTyp = relTyp[0]
	} else if len(relTyp) > 1 {
		return errors.New("len(relTyp) no more than 1")
	}
	if len(res) == 0 {
		return errors.New("res empty")
	}
	return _mgr.AddDepends(db, res, _relTyp, resB.ID, resB.Typ)
}

// DelRelation 删除关系：资源A 被 资源B 依赖
// relTyp长度不超过1；relType为空，指定默认依赖方式
func DelRelation(db *gorm.DB, resA rel.Res, resB rel.Res, relTyp ...rel.Type) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	_relTyp := rel.TypeDefault
	if len(relTyp) == 1 {
		_relTyp = relTyp[0]
	} else if len(relTyp) > 1 {
		return errors.New("len(relTyp) no more than 1")
	}
	return _mgr.DelRelation(db, resA.ID, resA.Typ, _relTyp, resB.ID, resB.Typ)
}

// CheckRelation 检查关系：资源A 是否被 资源B 依赖
// relTyp长度不超过1；relType为空，指定默认依赖方式
func CheckRelation(db *gorm.DB, resA rel.Res, resB rel.Res, relTyp ...rel.Type) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	_relTyp := rel.TypeDefault
	if len(relTyp) == 1 {
		_relTyp = relTyp[0]
	} else if len(relTyp) > 1 {
		return errors.New("len(relTyp) no more than 1")
	}
	return _mgr.CheckRelation(db, resA.ID, resA.Typ, _relTyp, resB.ID, resB.Typ)
}

// GetRelations 查询关系：查询某个资源【被依赖】的关系
// relTypes为空，会检查该资源的所有被依赖关系
func GetRelations(db *gorm.DB, res rel.Res, relTypes ...rel.Type) ([]*rel.Relation, error) {
	if _mgr == nil {
		return nil, errors.New("depend mgr not init")
	}
	return _mgr.GetRelations(db, res.ID, res.Typ, relTypes)
}

// GetDependents 查询关系：查询某个资源【所依赖】的关系
// relTypes为空，会检查该资源的所有所依赖关系
func GetDependents(db *gorm.DB, res rel.Res, relTypes ...rel.Type) ([]*rel.Relation, error) {
	if _mgr == nil {
		return nil, errors.New("depend mgr not init")
	}
	return _mgr.GetDependents(db, res.ID, res.Typ, relTypes)
}

// CheckOp 检查是否可以对某个资源执行op操作
func CheckOp(db *gorm.DB, res rel.Res, op rel.ResOp) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	return _mgr.CheckOp(db, res.ID, res.Typ, op)
}

// DelRes 删除某个资源的所有关系记录
func DelRes(db *gorm.DB, res rel.Res) error {
	if _mgr == nil {
		return errors.New("depend mgr not init")
	}
	return _mgr.DelRes(db, res.ID, res.Typ)
}
