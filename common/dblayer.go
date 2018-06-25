package common

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	// go-sql-driver
	_ "github.com/go-sql-driver/mysql"
)

const (
	SQL_INSERT = "INSERT INTO %s (%s) VALUES (%s)"
	SQL_UPDATE = "UPDATE %s SET %s WHERE %s"
	SQL_DELETE = "DELETE FROM %s WHERE %s"
)

type DbRow map[string]interface{}

type DbRowSortByIntCt []DbRow

func (r DbRowSortByIntCt) Len() int           { return len(r) }
func (r DbRowSortByIntCt) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r DbRowSortByIntCt) Less(i, j int) bool { return ToInt64(r[i]["ct"]) < ToInt64(r[j]["ct"]) }

func (t DbRow) GetInt(key string) int {
	v, exists := t[key]
	if !exists || v == nil {
		return 0
	}

	switch v.(type) {
	case int:
		return v.(int)
	case int64:
		return int(v.(int64))
	default:
		panic("unsupported type for func GetInt")
	}
}

func (t DbRow) GetInt64(key string) int64 {
	value, exists := t[key]
	if !exists || value == nil {
		return 0
	}
	return value.(int64)
}

func (t DbRow) GetString(key string) string {
	value, exists := t[key]
	if !exists || value == nil {
		return ""
	}
	return value.(string)
}

type DbWrapper interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Prepare(query string) (*sql.Stmt, error)
}

type DbLayer struct {
	Db          interface{}
	ConnTimeout int64
}

func NewDbLayer(driver, dsn string, fn func(*sql.DB)) (*DbLayer, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if fn != nil {
		fn(db)
	}

	p := WrapDbLayer(db)
	return p, nil
}

func WrapDbLayer(db interface{}) *DbLayer {
	p := new(DbLayer)
	p.Db = db
	p.ConnTimeout = time.Now().Unix() + 60
	return p
}

func dbScan(rows *sql.Rows) DbRow {
	r := DbRow{}

	cols, _ := rows.Columns()
	c := len(cols)
	vals := make([]interface{}, c)
	valPtrs := make([]interface{}, c)

	for i := range cols {
		valPtrs[i] = &vals[i]
	}

	rows.Scan(valPtrs...)

	for i := range cols {
		if val, ok := vals[i].([]byte); ok {
			r[cols[i]] = string(val)
		} else {
			r[cols[i]] = vals[i]
		}
	}

	return r
}

func (p *DbLayer) Close() {
	if db, ok := p.Db.(*sql.DB); ok {
		db.Close()
	}
}

func (p *DbLayer) Transaction(fn func(*DbLayer) error) error {
	if db, ok := p.Db.(*sql.DB); ok {
		if tx, err := db.Begin(); err != nil {
			return err
		} else {
			if err = fn(WrapDbLayer(tx)); err != nil {
				tx.Rollback()
				return err
			} else {
				tx.Commit()
			}
		}
	}
	return nil
}

func (p *DbLayer) Exec(sql string, args ...interface{}) (res sql.Result, err error) {
	res, err = p.Db.(DbWrapper).Exec(sql, args...)
	return
}

func (p *DbLayer) Prepare(query string) (*sql.Stmt, error) {
	stmt, err := p.Db.(*sql.DB).Prepare(query)
	return stmt, err
}

func (p *DbLayer) Insert(table string, row DbRow) (int64, error) {
	var (
		fields []string
		values []string
		args   []interface{}
	)
	for field, value := range row {
		fields = append(fields, field)
		values = append(values, "?")
		args = append(args, value)
	}

	code := fmt.Sprintf(SQL_INSERT, table, strings.Join(fields, ", "), strings.Join(values, ", "))

	res, err := p.Db.(DbWrapper).Exec(code, args...)
	if err != nil {
		return 0, err
	}

	r, _ := res.LastInsertId()
	return r, nil
}

func (p *DbLayer) BatchInsert(table string, rows []DbRow, fields []string) (int64, error) {
	var (
		values  []string
		nfields []string
		args    []interface{}
	)
	for _, f := range fields {
		nfields = append(nfields, fmt.Sprintf("`%s`", f))
	}

	for _, r := range rows {
		var val []string
		for _, field := range fields {
			if v, ok := r[field]; ok {
				args = append(args, v)
			} else {
				args = append(args, "")
			}
			val = append(val, "?")
		}
		values = append(values, fmt.Sprintf("(%s)", strings.Join(val, ",")))
	}
	batch_value := strings.Join(values, ",")
	code := fmt.Sprintf("replace into %s (%s) values %s", table, strings.Join(nfields, ", "), batch_value)
	//ErrorLog(code)
	//ErrorLog("%v", args)
	res, err := p.Db.(DbWrapper).Exec(code, args...)
	if err != nil {
		return 0, err
	}

	r, _ := res.RowsAffected()
	return r, nil

}

func (p *DbLayer) Update(table string, row DbRow, condition string, args ...interface{}) (int64, error) {
	var (
		fields []string
		values []interface{}
	)
	for field, value := range row {
		fields = append(fields, fmt.Sprintf("%s = ?", field))
		values = append(values, value)
	}
	args = append(values, args...)

	code := fmt.Sprintf(SQL_UPDATE, table, strings.Join(fields, ", "), condition)

	res, err := p.Db.(DbWrapper).Exec(code, args...)
	if err != nil {
		return 0, err
	}

	r, _ := res.RowsAffected()
	return r, nil
}

func (p *DbLayer) Delete(table, condition string, args ...interface{}) (int64, error) {
	code := fmt.Sprintf(SQL_DELETE, table, condition)

	res, err := p.Db.(DbWrapper).Exec(code, args...)
	if err != nil {
		return 0, err
	}

	r, _ := res.RowsAffected()
	return r, nil
}

func (p *DbLayer) One(code string, args ...interface{}) (DbRow, error) {
	rows, err := p.Db.(DbWrapper).Query(code, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()

	return dbScan(rows), nil
}

func (p *DbLayer) All(code string, args ...interface{}) ([]DbRow, error) {
	rows, err := p.Db.(DbWrapper).Query(code, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	r := make([]DbRow, 0)

	for rows.Next() {
		r = append(r, dbScan(rows))
	}

	return r, nil
}

func (p *DbLayer) Scalar(code string, args ...interface{}) (interface{}, error) {
	rows, err := p.Db.(DbWrapper).Query(code, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()

	var r interface{}
	if err = rows.Scan(&r); err != nil {
		return nil, err
	}

	if val, ok := r.([]byte); ok {
		return string(val), nil
	}

	return r, nil
}

func (p *DbLayer) QueryRow(qsql string, fields []interface{}, args ...interface{}) error {
	stmt, err := p.Prepare(qsql)
	if err != nil {
		fmt.Printf("queryrow-prepare-err: %+v\n", err)
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(fields...).Scan(args...)
	if err == sql.ErrNoRows {
		return errors.New("not found")
	}
	return err
}

type DbRowSortByCT []DbRow

func (self DbRowSortByCT) Len() int {
	return len(self)
}

func (self DbRowSortByCT) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self DbRowSortByCT) Less(i, j int) bool {
	cti := ""
	irow := self[i]
	if ct_val, ok := irow["ct"]; ok {
		cti = ct_val.(string)
	}
	ctj := ""
	jrow := self[j]
	if ct_val, ok := jrow["ct"]; ok {
		ctj = ct_val.(string)
	}
	return cti < ctj
}

func SortByCT(rows []DbRow) {
	sort.Sort(sort.Reverse(DbRowSortByCT(rows)))
}

type MapSortByCT []map[string]interface{}

func (self MapSortByCT) Len() int {
	return len(self)
}

func (self MapSortByCT) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self MapSortByCT) Less(i, j int) bool {
	cti := ""
	irow := self[i]
	if ct_val, ok := irow["ct"]; ok {
		cti = ct_val.(string)
	}
	ctj := ""
	jrow := self[j]
	if ct_val, ok := jrow["ct"]; ok {
		ctj = ct_val.(string)
	}
	return cti < ctj
}

func SortByCTMap(rows []map[string]interface{}) {
	sort.Sort(sort.Reverse(MapSortByCT(rows)))
}
