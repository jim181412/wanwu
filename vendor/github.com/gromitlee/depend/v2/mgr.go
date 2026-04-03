package depend

import (
	"github.com/gromitlee/depend/v2/internal"
	"github.com/gromitlee/depend/v2/rel"
	"gorm.io/gorm"
)

type IDependMgr interface {
	Register(typ rel.ResType, op rel.ResOp, ignores []rel.ResType) error
	AddRelation(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, idB string, typB rel.ResType) error
	AddRelations(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, res []rel.Res) error
	AddDepends(db *gorm.DB, res []rel.Res, relTyp rel.Type, idB string, typB rel.ResType) error
	DelRelation(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, idB string, typB rel.ResType) error
	CheckRelation(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, idB string, typB rel.ResType) error
	GetRelations(db *gorm.DB, id string, typ rel.ResType, relTypes []rel.Type) ([]*rel.Relation, error)
	GetDependents(db *gorm.DB, id string, typ rel.ResType, relTypes []rel.Type) ([]*rel.Relation, error)
	CheckOp(db *gorm.DB, id string, typ rel.ResType, op rel.ResOp) error
	DelRes(db *gorm.DB, id string, typ rel.ResType) error
}

func NewDependMgr(db *gorm.DB) (*internal.Mgr, error) {
	if err := db.AutoMigrate(rel.Relation{}); err != nil {
		return nil, err
	}
	return internal.NewMgr(), nil
}
