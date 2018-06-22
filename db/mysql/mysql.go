package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	cmn "github.com/moxiaomomo/configcenter/common"
	"github.com/moxiaomomo/configcenter/logger"
	proto "github.com/moxiaomomo/configcenter/proto"
)

const (
	MAX_IDLE_CONN    = 512
	MAX_CONN_TIMEOUT = 2 * time.Minute

	DB_HOST        = "10.220.16.133:3306"
	DB_USER        = "test"
	DB_PWD         = "123456"
	DB_NAME        = "config"
	TABLE_NAME     = "configs"
	TABLE_LOG_NAME = "configs_audit"
)

const (
	CFG_Q_FILEDS = `name, path, version, comment, createdAt, updatedAt, status, 
		changeset_timestamp, changeset_checksum, changeset_data, changeset_source, changeset_format`
	LOG_Q_FILEDS = `action, name, path, version, comment, createdAt, updatedAt, status, 
		changeset_timestamp, changeset_checksum, changeset_data, changeset_source, changeset_format`
)

var (
	changeQ = map[string]string{
		"read":   `SELECT _cfg_ from %s.%s WHERE name=? AND path=? AND version=? _status_ LIMIT 1`,
		"create": `INSERT INTO %s.%s (_cfg_) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"update": `UPDATE %s.%s SET comment=?, updatedAt=?, status=?, changeset_timestamp=?, changeset_checksum=?, 
			changeset_data=?, changeset_source=?, changeset_format=? WHERE name=? AND path=? AND version=? AND status>=0 LIMIT 1`,
		"delete":       `UPDATE %s.%s SET updatedAt=?, status=-1 WHERE name=? AND path=? AND version=? LIMIT 1`,
		"delforCreate": `DELETE FROM %s.%s WHERE name=? AND path=? AND version=? AND status=-1 LIMIT 1`,
		"search":       `SELECT _cfg_ FROM %s.%s WHERE status>=0 LIMIT ? OFFSET ?`,
		"searchParam":  `SELECT _cfg_ FROM %s.%s WHERE _param_ AND status>=0 LIMIT ? OFFSET ?`,
		"scan":         `SELECT _cfg_ FROM %s.%s LIMIT ? OFFSET ?`,
	}

	changeLogQ = map[string]string{
		"opLog":     `INSERT INTO %s.%s (_log_) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"searchLog": `SELECT _log_ FROM %s.%s WHERE name=? AND path=? LIMIT ? OFFSET ?`,
		"scanLog":   `SELECT _log_ FROM %s.%s _scan_ _orderby_ LIMIT ? OFFSET ?`,
	}

	DBUrl = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", DB_USER, DB_PWD, DB_HOST, DB_NAME)

	qSQLs   = map[string]string{}
	logSQLs = map[string]string{}
)

type MyConn struct {
	mysql *cmn.DbLayer
}

// Init mysql connection initiation
func (m *MyConn) Init() error {
	if m.mysql != nil && m.mysql.ConnTimeout > time.Now().Unix() {
		return nil
	}

	dblayer, err := cmn.NewDbLayer("mysql", DBUrl, func(db *sql.DB) {
		db.SetMaxIdleConns(MAX_IDLE_CONN)
		//		db.SetConnMaxLifetime(MAX_CONN_TIMEOUT)
	})
	if err != nil {
		logger.Errorf("Fail to init mysql-Myconn:%v", err)
		return err
	}
	m.mysql = dblayer

	for query, statement := range changeQ {
		statement = strings.Replace(statement, "_cfg_", CFG_Q_FILEDS, -1)
		qSQLs[query] = fmt.Sprintf(statement, DB_NAME, TABLE_NAME)
	}
	for query, statement := range changeLogQ {
		statement = strings.Replace(statement, "_log_", LOG_Q_FILEDS, -1)
		logSQLs[query] = fmt.Sprintf(statement, DB_NAME, TABLE_LOG_NAME)
	}
	logger.Info("Suc to init mysql connection.")
	return nil
}

// Uninit close mysql connection
func (m *MyConn) Uninit() error {
	if m.mysql != nil {
		m.mysql.Close()
	}
	return nil
}

// Create create tables
func (m *MyConn) Create(change *proto.ConfigSet) error {
	m.mysql.Exec(qSQLs["delforCreate"],
		change.Name,
		change.Path,
		change.Version,
	)

	_, err := m.mysql.Exec(qSQLs["create"],
		change.Name,
		change.Path,
		change.Version,
		change.Comment,
		change.CreatedAt,
		change.UpdatedAt,
		change.Status,
		change.ChangeSet.Timestamp,
		change.ChangeSet.Checksum,
		change.ChangeSet.Data,
		change.ChangeSet.Source,
		change.ChangeSet.Format,
	)
	if err != nil {
		return err
	}

	_, err = m.mysql.Exec(logSQLs["opLog"],
		"create",
		change.Name,
		change.Path,
		change.Version,
		change.Comment,
		change.CreatedAt,
		change.UpdatedAt,
		change.Status,
		change.ChangeSet.Timestamp,
		change.ChangeSet.Checksum,
		change.ChangeSet.Data,
		change.ChangeSet.Source,
		change.ChangeSet.Format,
	)
	if err != nil {
		logger.Errorf("Write configset_audit error: %+v\n", err)
	}
	return nil
}

// Update update mysql table
func (m *MyConn) Update(change *proto.ConfigSet) error {
	_, err := m.mysql.Exec(
		qSQLs["update"],
		change.Comment,
		change.UpdatedAt,
		change.Status,
		change.ChangeSet.Timestamp,
		change.ChangeSet.Checksum,
		change.ChangeSet.Data,
		change.ChangeSet.Source,
		change.ChangeSet.Format,
		change.Name,
		change.Path,
		change.Version,
	)
	if err != nil {
		return err
	}

	_, err = m.mysql.Exec(logSQLs["opLog"],
		"update",
		change.Name,
		change.Path,
		change.Version,
		change.Comment,
		change.CreatedAt,
		change.UpdatedAt,
		change.Status,
		change.ChangeSet.Timestamp,
		change.ChangeSet.Checksum,
		change.ChangeSet.Data,
		change.ChangeSet.Source,
		change.ChangeSet.Format,
	)
	if err != nil {
		logger.Errorf("Write configset_audit error: %+v\n", err)
	}
	return nil
}

// Delete delete some rows
func (m *MyConn) Delete(change *proto.ConfigSet) error {
	_, err := m.mysql.Exec(qSQLs["delete"], change.UpdatedAt, change.Name, change.Path, change.Version)
	if err != nil {
		return err
	}

	_, err = m.mysql.Exec(logSQLs["opLog"],
		"delete",
		change.Name,
		change.Path,
		change.Version,
		change.Comment,
		0,
		0,
		-1,
		"",
		"",
		"",
		"",
		"",
	)
	if err != nil {
		logger.Errorf("Write configset_audit error: %+v\n", err)
	}
	return nil
}

// Read query some rows
func (m *MyConn) Read(name, path, version string, delInclude bool) (*proto.ConfigSet, error) {
	if len(name) == 0 || len(path) == 0 {
		return nil, errors.New("Invalid config uniq_id")
	}
	if len(version) == 0 {
		version = "1.0"
	}

	var state = "AND status>=0"
	if delInclude {
		state = ""
	}
	var qSQL = strings.Replace(qSQLs["read"], "_status_", state, -1)

	change := &proto.ConfigSet{
		ChangeSet: &proto.ChangeSet{},
	}

	t := []string{name, path, version}
	params := make([]interface{}, len(t))
	for i, v := range t {
		params[i] = v
	}
	err := m.mysql.QueryRow(
		qSQL,
		params,
		&change.Name,
		&change.Path,
		&change.Version,
		&change.Comment,
		&change.CreatedAt,
		&change.UpdatedAt,
		&change.Status,
		&change.ChangeSet.Timestamp,
		&change.ChangeSet.Checksum,
		&change.ChangeSet.Data,
		&change.ChangeSet.Source,
		&change.ChangeSet.Format,
	)

	if err != nil {
		logger.Errorf("read_sqldb_err for ::name:%s path:%s version:%s::%+v\n",
			name, path, version, err)
		return nil, err
	}
	return change, nil
}

// Search query some rows
func (m *MyConn) Search(name, path, version string, limit, offset int64) (reslist []*proto.ConfigSet, err error) {
	var qStr []string
	var qParam []interface{}
	var params string

	if len(name) > 0 {
		qStr = append(qStr, " name=? ")
		qParam = append(qParam, name)
	}
	if len(path) > 0 {
		qStr = append(qStr, " path=? ")
		qParam = append(qParam, path)
	}
	if len(version) > 0 {
		qStr = append(qStr, " version=? ")
		qParam = append(qParam, version)
	}

	if len(qStr) <= 0 {
		return nil, errors.New("search params are invalid.")
	} else if len(qStr) == 1 {
		params = qStr[0]
	} else {
		params = strings.Join(qStr, " AND ")
	}
	var qSQL = strings.Replace(qSQLs["searchParam"], "_param_", params, -1)

	qParam = append(qParam, limit)
	qParam = append(qParam, offset)
	var dbRows []cmn.DbRow
	dbRows, err = m.mysql.All(qSQL, qParam...)
	if err != nil {
		return nil, err
	}

	reslist = []*proto.ConfigSet{}

	for _, row := range dbRows {
		cfg := &proto.ConfigSet{
			Name:      row["name"].(string),
			Path:      row["path"].(string),
			Version:   row["version"].(string),
			Comment:   row["comment"].(string),
			CreatedAt: row["createdAt"].(int64),
			UpdatedAt: row["updatedAt"].(int64),
			Status:    int32(row["status"].(int64)),
		}
		cfg.ChangeSet = &proto.ChangeSet{
			Timestamp: row["changeset_timestamp"].(int64),
			Checksum:  row["changeset_checksum"].(string),
			Data:      row["changeset_data"].(string),
			Source:    row["changeset_source"].(string),
			Format:    row["changeset_format"].(string),
		}
		reslist = append(reslist, cfg)
	}

	return reslist, nil
}

// AuditLog query audit logs
func (m *MyConn) AuditLog(fromts, tots, limit, offset int64, reverse bool) (reslist []*proto.ChangeLog, err error) {
	var params = ""
	var qParam []interface{}

	if fromts > 0 && tots >= fromts {
		params = "WHERE updatedAt>=? AND updatedAt<=?"
		qParam = append(qParam, fromts)
		qParam = append(qParam, tots)
	}
	qParam = append(qParam, limit)
	qParam = append(qParam, offset)

	var logSQL = strings.Replace(logSQLs["scanLog"], "_scan_", params, -1)

	var oderby = "ORDER BY updatedAt ASC"
	if reverse {
		oderby = "ORDER BY updatedAt DESC"
	}
	logSQL = strings.Replace(logSQL, "_orderby_", oderby, -1)

	var dbRows []cmn.DbRow
	dbRows, err = m.mysql.All(logSQL, qParam...)
	if err != nil {
		return nil, err
	}

	reslist = []*proto.ChangeLog{}

	for _, row := range dbRows {
		cfg := &proto.ConfigSet{

			Name:      row["name"].(string),
			Path:      row["path"].(string),
			Version:   row["version"].(string),
			Comment:   row["comment"].(string),
			CreatedAt: row["createdAt"].(int64),
			UpdatedAt: row["updatedAt"].(int64),
			Status:    int32(row["status"].(int64)),
		}
		cfg.ChangeSet = &proto.ChangeSet{
			Timestamp: row["changeset_timestamp"].(int64),
			Checksum:  row["changeset_checksum"].(string),
			Data:      row["changeset_data"].(string),
			Source:    row["changeset_source"].(string),
			Format:    row["changeset_format"].(string),
		}

		changlog := &proto.ChangeLog{
			Action:    row["action"].(string),
			ConfigSet: cfg,
		}
		reslist = append(reslist, changlog)
	}

	return reslist, nil
}
