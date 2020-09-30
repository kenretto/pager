package driver

import (
	"fmt"
	"github.com/kenretto/pager"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// Gorm 查询
type Gorm struct {
	conn  *gorm.DB
	query string
	args  []interface{}
	limit int
	skip  int
	index string
	sorts string
}

// NewGormDriver gorm 分页实现
func NewGormDriver(conn *gorm.DB) *Gorm {
	return &Gorm{
		conn: conn,
	}
}

// WithDB 选择另外的数据库
func (orm *Gorm) WithDB(db *gorm.DB) *Gorm {
	orm.conn = db
	return orm
}

// Where 构建查询条件
func (orm *Gorm) Where(kv pager.Where) {
	for k, v := range kv {
		if reflect.ValueOf(v).Kind() == reflect.Slice {
			orm.query += fmt.Sprintf("AND `%s` in(?) ", k)
		} else {
			orm.query += fmt.Sprintf("AND `%s` = ? ", k)
		}
		orm.args = append(orm.args, v)
	}
}

// Range 范围查询条件
func (orm *Gorm) Range(r pager.Range) {
	for k, v := range r {
		if val, ok := v[pager.Gte]; ok {
			orm.query += fmt.Sprintf("AND `%s` > ? ", k)
			orm.args = append(orm.args, val)
		}
		if val, ok := v[pager.Lte]; ok {
			orm.query += fmt.Sprintf("AND `%s` < ? ", k)
			orm.args = append(orm.args, val)
		}
	}
}

// Limit 每页数量
func (orm *Gorm) Limit(limit int) {
	orm.limit = limit
}

// Skip 跳过数量
func (orm *Gorm) Skip(skip int) {
	orm.skip = skip
}

// Index table 表名
func (orm *Gorm) Index(index string) {
	orm.index = index
}

// Sort 排序
func (orm *Gorm) Sort(sortInfo []pager.SortInfo) {
	var sorts []string
	for _, sort := range sortInfo {
		if sort.Sort == pager.Asc {
			sorts = append(sorts, fmt.Sprintf("%s ASC", sort.Key))
		} else {
			sorts = append(sorts, fmt.Sprintf("%s DESC", sort.Key))
		}
	}
	orm.sorts = strings.Join(sorts, ",")
}

// Find 从数据库查询数据
func (orm *Gorm) Find(data interface{}) {
	orm.query = strings.TrimPrefix(orm.query, "AND")
	var query = orm.conn.Table(orm.index).Where(orm.query, orm.args...).Limit(orm.limit).Offset(orm.skip)
	if orm.sorts != "" {
		query = query.Order(orm.sorts)
	}
	query.Find(data)
}

// SetTyp sql 对数据查询的类型不敏感
func (orm *Gorm) SetTyp(typ reflect.Type) {}

// Count 计算指定查询条件的总数量
func (orm *Gorm) Count() int64 {
	var count int64
	orm.conn.Table(orm.index).Where(orm.query, orm.args...).Count(&count)
	return count
}
