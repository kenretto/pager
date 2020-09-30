// Package pager Paging tool
//  pager.New(ctx, driver).SetIndex(c.entity.TableName()).Find(c.entity).Result()
package pager

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Where filter
type Where map[string]interface{}

// Sort sort
type Sort int

const (
	// Desc desc
	Desc Sort = -1
	// Asc asc
	Asc = 1
)

// SortInfo sort info
type SortInfo struct {
	Key  string
	Sort Sort
}

// FilterKey Custom name of the parameter carried by url
type FilterKey struct {
	// page parameter name
	Page string
	// rows parameter name
	Rows string
	// sorts parameter name
	Sorts string
	// range parameter name
	Range string
}

// Driver Data query driver, Pagination will pass the parsed parameters to it.
// Theoretically, any driver that implements this interface can call this package to achieve the general paging function.
type Driver interface {
	// Set filter conditions
	Where(kv Where)
	// Set range query conditions
	Range(r Range)
	// Set the limit per page
	Limit(limit int)
	// Set the number of skips
	Skip(skip int)
	// Index Index name, which can be a table name, or, for example, the index name of es, something that identifies a specific collection of resources
	Index(index string)
	// 排序
	Sort(kv []SortInfo)
	// Query operation, driver, it is recommended to perform real data scanning in this method to save performance
	Find(data interface{})
	// If the data structure passed by the Find method needs to be analyzed in the implementation code of the specific driver,
	// this method will be called and passed in the data reflection type. For example, in mongo,
	// the construction of the query language has strict requirements on the type. Here,
	// the reflection Type of the data type will be passed in to drive the driver to analyze the record data type and convert the filtered data passed in by Where.
	SetTyp(typ reflect.Type)
	// Calculate the total number of qualified data items according to all filter conditions
	Count() int64
}

// Result Result Object
type Result struct {
	Data   interface{} `json:"data" xml:"data"`       // data slice
	NextID interface{} `json:"next_id" xml:"next_id"` // next page id
	PrevID interface{} `json:"prev_id" xml:"prev_id"` // prev page id
	Count  int64       `json:"count" xml:"count"`     // number of qualified data items according to all filter conditions
	Rows   int         `json:"rows" xml:"rows"`       // Display the number of entries per page
}

// Pagination page tool
type Pagination struct {
	ctx             *http.Request
	defaultWhere    Where
	defaultLimit    int
	index           string
	driver          Driver
	dataTyp         reflect.Type
	paginationField string
	result          *Result

	filterKey FilterKey

	// Field mapping. If the key passed in is not a key, with the database, write this. Key is the parameter passed in by GET. The key, value is the database field name.
	fieldMapping map[string]string

	// The value set here will be removed from the where parameter, and value is the same value as the data table column
	disabledField []string
}

// New instance
func New(request *http.Request, driver Driver) *Pagination {
	return &Pagination{
		ctx:          request,
		driver:       driver,
		defaultLimit: 12,
		fieldMapping: make(map[string]string),
		filterKey: FilterKey{
			Page:  "page",
			Rows:  "rows",
			Sorts: "sorts",
			Range: "range",
		},
	}
}

// SetFilterKey custom paging related query parameters
func (pagination *Pagination) SetFilterKey(key FilterKey) *Pagination {
	pagination.filterKey = key
	return pagination
}

// AddDisableField add fields that are prohibited from participating in search filtering
func (pagination *Pagination) AddDisableField(fields ...string) *Pagination {
	pagination.disabledField = append(pagination.disabledField, fields...)
	return pagination
}

// AddFilterKeyMapping filter condition mapping. If the key carried by url is different from the field name of the database, set it here
func (pagination *Pagination) AddFilterKeyMapping(key, column string) *Pagination {
	pagination.fieldMapping[key] = column
	return pagination
}

// Result get result
func (pagination *Pagination) Result() *Result {
	return pagination.result
}

// SetPaginationField when this value is set, the next_id in the returned data will be equal to the field value of the last element in slice,
// and prev_id will be equal to the field value of the first element in slice
func (pagination *Pagination) SetPaginationField(field string) *Pagination {
	pagination.paginationField = field
	return pagination
}

// SetIndex set the collection or table name in the query driver, for example, the database is the table name
func (pagination *Pagination) SetIndex(index string) *Pagination {
	pagination.index = index
	return pagination
}

// Where set default query parameters, such as returning only the user's own browsing log after the user logs in,
// which is valid throughout the life cycle of the object.
func (pagination *Pagination) Where(kv Where) *Pagination {
	pagination.defaultWhere = kv
	return pagination
}

// Limit set the default number of data items returned per page
func (pagination *Pagination) Limit(limit int) *Pagination {
	pagination.defaultLimit = limit
	return pagination
}

// Sort specify sort
func (pagination *Pagination) Sort(kv []SortInfo) *Pagination {
	pagination.driver.Sort(kv)
	return pagination
}

// Range specify range query parameters
func (pagination *Pagination) Range(r Range) *Pagination {
	pagination.driver.Range(r)
	return pagination
}

// Find use driver to query data
//  structure the structure of the data, without the need for a pointer
func (pagination *Pagination) Find(structure interface{}) *Pagination {
	limit := pagination.ParsingLimit()

	pagination.dataTyp = reflect.TypeOf(structure)
	pagination.driver.SetTyp(pagination.dataTyp)
	pagination.driver.Limit(limit)
	pagination.driver.Sort(ParseSorts(pagination.ctx, pagination.fieldMapping))
	pagination.driver.Index(pagination.index)
	pagination.driver.Skip(limit * ParseSkip(pagination.ctx))
	pagination.driver.Range(pagination.ParseRange())

	var filter = pagination.mergeWhere(pagination.ParsingQuery())
	for _, column := range pagination.disabledField {
		delete(filter, column)
	}
	pagination.driver.Where(filter)

	data := newSlice(pagination.dataTyp)
	pagination.driver.Find(data.Interface())
	pagination.result = &Result{
		Data:  data.Interface(),
		Count: pagination.driver.Count(),
		Rows:  limit,
	}
	if data.Elem().Len() > 0 && pagination.paginationField != "" {
		pagination.result.NextID = data.Elem().Index(data.Elem().Len() - 1).FieldByName(pagination.paginationField).Interface()
		pagination.result.PrevID = data.Elem().Index(0).FieldByName(pagination.paginationField).Interface()
	}

	return pagination
}

// ParseSkip Number of rows skipped
func ParseSkip(request *http.Request) int {
	var page, err = strconv.Atoi(request.URL.Query().Get("page"))
	if page == 0 || err != nil {
		return 0
	}
	return page - 1
}

// ParseSorts Parse the sort field, and the passed-in rule is  sorts=-filed1,+field2,field3
//  "-"  The "-" sign identifies the descending order.
//  "+" the "+" or unsigned ascending order
func ParseSorts(request *http.Request, filterMapping map[string]string) []SortInfo {
	var sortMap = make([]SortInfo, 0)
	query := request.URL.Query().Get("sorts")
	if query == "" {
		return sortMap
	}
	sorts := strings.Split(query, ",")
	for _, sort := range sorts {
		if strings.HasPrefix(sort, "-") {
			var column = strings.TrimPrefix(sort, "-")
			if v, ok := filterMapping[column]; ok {
				column = v
			}
			sortMap = append(sortMap, SortInfo{
				Key:  column,
				Sort: Desc,
			})
		} else {
			var column = strings.TrimPrefix(sort, "+")
			if v, ok := filterMapping[column]; ok {
				column = v
			}
			sortMap = append(sortMap, SortInfo{
				Key:  column,
				Sort: Asc,
			})
		}
	}
	return sortMap
}

// ParsingLimit Parse the number of entries displayed per page from the http request
func (pagination *Pagination) ParsingLimit() int {
	var rows = pagination.ctx.URL.Query().Get(pagination.filterKey.Rows)
	limit, err := strconv.Atoi(rows)
	if err != nil || limit == 0 {
		return pagination.defaultLimit
	}
	return limit
}

// ParseRange range
func (pagination *Pagination) ParseRange() Range {
	rangeFilter := make(Range)
	rangeFilter.Parse(pagination.ctx)

	for key, m := range rangeFilter {
		if column, ok := pagination.fieldMapping[string(key)]; ok {
			rangeFilter[RangeKey(column)] = m
			delete(rangeFilter, key)
		}
	}

	return rangeFilter
}

// ParsingQuery Parse the filter condition from the http request
func (pagination *Pagination) ParsingQuery() Where {
	where := make(Where)
	query := pagination.ctx.URL.Query()
	for key, val := range query {
		if v, ok := pagination.fieldMapping[key]; ok {
			key = v
		}

		if len(val) == 1 {
			if val[0] != "" {
				where[key] = val[0]
			}
		}
		if len(val) > 1 {
			where[key] = val
		}
	}
	return where
}

func (pagination *Pagination) mergeWhere(where Where) Where {
	for k, v := range pagination.defaultWhere {
		if k == pagination.filterKey.Rows || k == pagination.filterKey.Sorts || k == pagination.filterKey.Page || k == pagination.filterKey.Range {
			continue
		}
		where[k] = v
	}
	delete(where, pagination.filterKey.Rows)
	delete(where, pagination.filterKey.Sorts)
	delete(where, pagination.filterKey.Page)
	delete(where, pagination.filterKey.Range)
	return where
}

func newSlice(typ reflect.Type) reflect.Value {
	newInstance := reflect.MakeSlice(reflect.SliceOf(typ), 0, 0)
	items := reflect.New(newInstance.Type())
	items.Elem().Set(newInstance)
	return items
}
