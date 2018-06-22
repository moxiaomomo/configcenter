package db

import (
	"errors"
	//	"fmt"
	"time"

	mydb "github.com/moxiaomomo/configcenter/db/mysql"
	proto "github.com/moxiaomomo/configcenter/proto"
)

type DB interface {
	Init() error
	Uninit() error
	Create(*proto.ConfigSet) error
	Delete(*proto.ConfigSet) error
	Update(*proto.ConfigSet) error
	Read(name, path, version string, delInclude bool) (*proto.ConfigSet, error)
	Search(name, path, version string, limit, offset int64) ([]*proto.ConfigSet, error)
	AuditLog(from, to, limit, offset int64, reverse bool) ([]*proto.ChangeLog, error)
}

var (
	db            DB = &mydb.MyConn{}
	ErrorNotFound    = errors.New("not found")
)

func Register(backend DB) {
	db = backend
}

func Init() error {
	return db.Init()
}

func Uninit() error {
	return db.Uninit()
}

func Create(ch *proto.ConfigSet) error {
	ch.CreatedAt = time.Now().Unix()
	ch.UpdatedAt = ch.CreatedAt
	ch.Status = 1
	if len(ch.ChangeSet.Format) == 0 {
		ch.ChangeSet.Format = "json"
	}
	if len(ch.Version) == 0 {
		ch.Version = "1.0"
	}
	return db.Create(ch)
}

func Delete(ch *proto.ConfigSet) error {
	ch.UpdatedAt = time.Now().Unix()
	ch.Status = -1
	return db.Delete(ch)
}

func Update(ch *proto.ConfigSet) error {
	ch.UpdatedAt = time.Now().Unix()
	if ch.Status < -1 || ch.Status > 1 {
		return errors.New("param status valus is invalid.")
	}
	if len(ch.ChangeSet.Format) == 0 {
		ch.ChangeSet.Format = "json"
	}
	return db.Update(ch)
}

func Read(name, path, version string, delInclude bool) (*proto.ConfigSet, error) {
	return db.Read(name, path, version, delInclude)
}

func Search(name, path, version string, status int32, limit, offset int64) ([]*proto.ConfigSet, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return db.Search(name, path, version, limit, offset)
}

func AuditLog(fromts, tots, limit, offset int64, reverse bool) ([]*proto.ChangeLog, error) {
	if limit <= 0 {
		limit = 1
	}
	if offset <= 0 {
		offset = 0
	}
	return db.AuditLog(fromts, tots, limit, offset, reverse)
}
