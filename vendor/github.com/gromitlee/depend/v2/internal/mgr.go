package internal

import (
	"fmt"
	"sync"

	"github.com/gromitlee/depend/v2/rel"
	"gorm.io/gorm"
)

type Mgr struct {
	mutex sync.Mutex
	rs    []*res
}

type res struct {
	typ rel.ResType
	ops []*opIgnores
}

// 不会影响操作的资源类型集合
type opIgnores struct {
	op      rel.ResOp
	ignores []rel.ResType
}

func (r *res) check(op rel.ResOp) (bool, []rel.ResType) {
	for _, _opIgnores := range r.ops {
		if _opIgnores.op == op {
			return true, _opIgnores.ignores
		}
	}
	return false, nil
}

func NewMgr() *Mgr {
	return &Mgr{}
}

func (mgr *Mgr) Register(typ rel.ResType, op rel.ResOp, ignores []rel.ResType) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var r *res
	for _, _r := range mgr.rs {
		if _r.typ == typ {
			r = _r
			break
		}
	}
	if r == nil {
		r = &res{typ: typ}
		mgr.rs = append(mgr.rs, r)
	}

	var o *opIgnores
	for _, _opIgnores := range r.ops {
		if _opIgnores.op == op {
			o = _opIgnores
			break
		}
	}
	if o == nil {
		o = &opIgnores{op: op}
		r.ops = append(r.ops, o)
	}

	for _, ignore := range ignores {
		for _, _ignore := range o.ignores {
			if _ignore == ignore {
				return fmt.Errorf("res type %v op %v ignore %v registered", typ, op, ignore)
			}
		}
		o.ignores = append(o.ignores, ignore)
	}

	return nil
}

func (mgr *Mgr) AddRelation(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, idB string, typB rel.ResType) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var existA bool
	for _, r := range mgr.rs {
		if r.typ == typA {
			existA = true
			break
		}
	}
	if !existA {
		return fmt.Errorf("unknown res %v type %v", idA, typA)
	}

	return db.Create(&rel.Relation{
		IDA:    idA,
		TypA:   typA,
		RelTyp: relTyp,
		IDB:    idB,
		TypB:   typB,
	}).Error
}

func (mgr *Mgr) AddRelations(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, res []rel.Res) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var existA bool
	for _, r := range mgr.rs {
		if r.typ == typA {
			existA = true
			break
		}
	}
	if !existA {
		return fmt.Errorf("unknown res %v type %v", idA, typA)
	}

	var relations []*rel.Relation
	for _, r := range res {
		relations = append(relations, &rel.Relation{
			IDA:    idA,
			TypA:   typA,
			RelTyp: relTyp,
			IDB:    r.ID,
			TypB:   r.Typ,
		})
	}
	return db.Create(relations).Error
}

func (mgr *Mgr) AddDepends(db *gorm.DB, res []rel.Res, relTyp rel.Type, idB string, typB rel.ResType) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	for _, r := range res {
		var exist bool
		for _, _r := range mgr.rs {
			if _r.typ == r.Typ {
				exist = true
				break
			}
		}
		if !exist {
			return fmt.Errorf("unknown res %v type %v", r.ID, r.Typ)
		}
	}

	var relations []*rel.Relation
	for _, r := range res {
		relations = append(relations, &rel.Relation{
			IDA:    r.ID,
			TypA:   r.Typ,
			RelTyp: relTyp,
			IDB:    idB,
			TypB:   typB,
		})
	}
	return db.Create(relations).Error
}

func (mgr *Mgr) DelRelation(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, idB string, typB rel.ResType) error {
	return db.Where("id_a = ? AND typ_a = ? AND rel_typ = ? AND id_b = ? AND typ_b = ?",
		idA, typA, relTyp, idB, typB).Delete(&rel.Relation{}).Error
}

func (mgr *Mgr) CheckRelation(db *gorm.DB, idA string, typA rel.ResType, relTyp rel.Type, idB string, typB rel.ResType) error {
	return db.Where("id_a = ? AND typ_a = ? AND rel_typ = ? AND id_b = ? AND typ_b = ?",
		idA, typA, relTyp, idB, typB).First(&rel.Relation{}).Error
}

// GetRelations 查询某个资源【被依赖】的关系；relTypes为空，会检查该资源的所有被依赖关系
func (mgr *Mgr) GetRelations(db *gorm.DB, id string, typ rel.ResType, relTypes []rel.Type) ([]*rel.Relation, error) {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var exist bool
	for _, r := range mgr.rs {
		if r.typ == typ {
			exist = true
			break
		}
	}
	if !exist {
		return nil, fmt.Errorf("unknown res %v type %v", id, typ)
	}

	var relations []*rel.Relation
	var err error
	if len(relTypes) > 0 {
		err = db.Where("id_a = ? AND typ_a = ? AND rel_typ IN ?", id, typ, relTypes).Find(&relations).Error
	} else {
		err = db.Where("id_a = ? AND typ_a = ?", id, typ).Find(&relations).Error
	}
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// GetDependents 查询某个资源【所依赖】的关系：relTypes为空，会检查该资源的所有所依赖关系
func (mgr *Mgr) GetDependents(db *gorm.DB, id string, typ rel.ResType, relTypes []rel.Type) ([]*rel.Relation, error) {
	var dependents []*rel.Relation
	var err error
	if len(relTypes) > 0 {
		err = db.Where("id_b = ? AND typ_b = ? AND rel_typ IN ?", id, typ, relTypes).Find(&dependents).Error
	} else {
		err = db.Where("id_b = ? AND typ_b = ?", id, typ).Find(&dependents).Error
	}
	if err != nil {
		return nil, err
	}
	return dependents, nil
}

// CheckOp 检查是否可以对某个资源执行op操作
func (mgr *Mgr) CheckOp(db *gorm.DB, id string, typ rel.ResType, op rel.ResOp) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var r *res
	for _, _r := range mgr.rs {
		if _r.typ == typ {
			r = _r
			break
		}
	}
	if r == nil {
		return fmt.Errorf("unknown res type %v", typ)
	}

	ok, ignores := r.check(op)
	if !ok {
		return fmt.Errorf("res %v type %v unknown op %v", id, typ, op)
	}

	var err error
	if len(ignores) > 0 {
		err = db.Where("id_a = ? AND typ_a = ? AND typ_b NOT IN ?", id, typ, ignores).First(&rel.Relation{}).Error
	} else {
		err = db.Where("id_a = ? AND typ_a = ?", id, typ).First(&rel.Relation{}).Error
	}
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		return nil
	}
	return fmt.Errorf("res %v type %v op %v has relation", id, typ, op)
}

// DelRes 删除某个资源的所有关系记录；会检查该资源的所有依赖关系
func (mgr *Mgr) DelRes(db *gorm.DB, id string, typ rel.ResType) error {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()

	var r *res
	for _, _r := range mgr.rs {
		if _r.typ == typ {
			r = _r
			break
		}
	}
	if r == nil {
		return fmt.Errorf("unknown res %v type %v", id, typ)
	}

	ok, ignores := r.check(rel.OpDelete)
	if !ok {
		return fmt.Errorf("res %v type %v unknown op %v", id, typ, rel.OpDelete)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var err error
		if len(ignores) > 0 {
			err = db.Where("id_a = ? AND typ_a = ? AND typ_b NOT IN ?", id, typ, ignores).First(&rel.Relation{}).Error
		} else {
			err = db.Where("id_a = ? AND typ_a = ?", id, typ).First(&rel.Relation{}).Error
		}
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
		} else {
			return fmt.Errorf("res %v type %v op %v has relation", id, typ, rel.OpDelete)
		}
		// 删除资源作为被依赖方
		if err := tx.Where("id_a = ? AND typ_a = ?", id, typ).Delete(&rel.Relation{}).Error; err != nil {
			return err
		}
		// 删除资源作为依赖方
		if err := tx.Where("id_b = ? AND typ_b = ?", id, typ).Delete(&rel.Relation{}).Error; err != nil {
			return err
		}
		return nil
	})
}
