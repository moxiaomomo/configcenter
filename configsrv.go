package main

import (
	"github.com/micro/go-micro/errors"
	"github.com/moxiaomomo/configcenter/db"
	"github.com/moxiaomomo/configcenter/logger"
	proto "github.com/moxiaomomo/configcenter/proto"
	"github.com/moxiaomomo/configcenter/watch"
	"github.com/moxiaomomo/configcenter/web"

	"golang.org/x/net/context"

	"github.com/neverlee/microframe/config"
	"github.com/neverlee/microframe/pluginer"
	"github.com/neverlee/microframe/service"
)

type Config struct {
	pluginer.SrvPluginBase
}

const (
	ERR_REQ_OK            = 0
	ERR_REQ_INVALID_ID    = -1
	ERR_REQ_SERVER_FAILED = -2
	ERR_PUBMSG_FAILED     = -3

	WEB_UI_PORT = 8765
)

func (c *Config) Read(ctx context.Context, req *proto.ReadRequest, rsp *proto.ReadResponse) error {
	logger.Info("Receive Read config request...")
	if len(req.Name) <= 0 || len(req.Path) <= 0 {
		rsp.Resp = ERR_REQ_INVALID_ID
		return nil
	}

	config, err := db.Read(req.Name, req.Path, req.Version, req.DelInclude)
	if err != nil {
		rsp.Resp = ERR_REQ_SERVER_FAILED
		return nil
	}

	rsp.ConfigSet = config
	return nil
}

func (c *Config) Create(ctx context.Context, req *proto.CreateRequest, rsp *proto.CreateResponse) error {
	logger.Info("Receive Create config request...")
	if err := db.Create(req.ConfigSet); err != nil {
		logger.Error(err)
		rsp.Resp = ERR_REQ_SERVER_FAILED
		return errors.InternalServerError("micro.frame.srv.config.Create", err.Error())
	} else {
		rsp.Resp = ERR_REQ_OK

		config, err := db.Read(req.ConfigSet.Name, req.ConfigSet.Path, req.ConfigSet.Version, true)
		if err != nil {
			logger.Error(err)
			return err
		} else if config != nil {
			watch.Publish(ctx, &proto.WatchResponse{
				Name:      config.Name,
				Path:      config.Path,
				Version:   config.Version,
				Status:    config.Status,
				ChangeSet: config.ChangeSet,
			})
		}
	}
	return nil
}

func (c *Config) Update(ctx context.Context, req *proto.UpdateRequest, rsp *proto.UpdateResponse) error {
	logger.Info("Receive Update config request...")
	if err := db.Update(req.ConfigSet); err != nil {
		logger.Error(err)
		rsp.Resp = ERR_REQ_SERVER_FAILED
		return errors.InternalServerError("micro.frame.srv.config.Update", err.Error())
	} else {
		rsp.Resp = ERR_REQ_OK

		config, err := db.Read(req.ConfigSet.Name, req.ConfigSet.Path, req.ConfigSet.Version, true)
		if err != nil {
			logger.Error(err)
			return err
		} else if config != nil {
			watch.Publish(ctx, &proto.WatchResponse{
				Name:      config.Name,
				Path:      config.Path,
				Version:   config.Version,
				Status:    config.Status,
				ChangeSet: config.ChangeSet,
			})
		}
	}
	return nil
}

func (c *Config) Delete(ctx context.Context, req *proto.DeleteRequest, rsp *proto.DeleteResponse) error {
	logger.Info("Receive Delete config request...")
	if err := db.Delete(req.ConfigSet); err != nil {
		logger.Error(err)
		rsp.Resp = ERR_REQ_SERVER_FAILED
		return errors.InternalServerError("micro.frame.srv.config.Delete", err.Error())
	} else {
		rsp.Resp = ERR_REQ_OK

		config, err := db.Read(req.ConfigSet.Name, req.ConfigSet.Path, req.ConfigSet.Version, true)
		if err != nil {
			logger.Error(err)
			return err
		} else if config != nil {
			watch.Publish(ctx, &proto.WatchResponse{
				Name:      config.Name,
				Path:      config.Path,
				Version:   config.Version,
				Status:    config.Status,
				ChangeSet: config.ChangeSet,
			})
		}
	}
	return nil
}

func (c *Config) Search(ctx context.Context, req *proto.SearchRequest, rsp *proto.SearchResponse) error {
	logger.Info("Receive Search config request...")
	res, err := db.Search(req.Name, req.Path, req.Version,
		req.Status, req.Limit, req.Offset)
	if err != nil {
		logger.Error(err)
		rsp.Resp = ERR_REQ_SERVER_FAILED
		return errors.InternalServerError("micro.frame.srv.config.Search", err.Error())
	} else {
		rsp.Resp = ERR_REQ_OK
		rsp.Configs = res
	}
	return nil
}

func (c *Config) AuditLog(ctx context.Context, req *proto.AuditLogRequest, rsp *proto.AuditLogResponse) error {
	logger.Info("Receive Auditlog config request...")

	res, err := db.AuditLog(req.From, req.To, req.Limit, req.Offset, req.Reverse)

	if err != nil {
		logger.Error(err)
		rsp.Resp = ERR_REQ_SERVER_FAILED
		return errors.InternalServerError("micro.frame.srv.config.AuditLog", err.Error())
	} else {
		rsp.Resp = ERR_REQ_OK
		rsp.Changes = res
	}
	return nil
}

func (c *Config) Watch(ctx context.Context, req *proto.WatchRequest, stream proto.Config_WatchStream) error {
	logger.Info("Receive Watch request...")

	if len(req.Name) == 0 || len(req.Path) == 0 {
		stream.Close()
		return errors.BadRequest("micro.frame.srv.config.Watch", "invalid name&path")
	}

	watchId := watch.ToWatchId(req.Name, req.Path, req.Version)
	watch, err := watch.Watch(watchId)
	if err != nil {
		stream.Close()
		return errors.InternalServerError("micro.frame.srv.config.Watch", err.Error())
	}
	defer watch.Stop()

	for {
		ch, err := watch.Next()
		if err != nil {
			stream.Close()
			return errors.InternalServerError("micro.frame.srv.config.Watch", err.Error())
		}

		if err := stream.Send(ch); err != nil {
			stream.Close()
			return errors.InternalServerError("micro.frame.srv.config.Watch", err.Error())
		}
	}
}

func NewPlugin(pconf *config.RawYaml) (pluginer.SrvPluginer, error) {
	s := &Config{
		SrvPluginBase: pluginer.SrvPluginBase{
			Phase: pluginer.ContentPhase,
		},
	}
	return s, nil
}

func (s *Config) Init(srv *service.Service) error {
	proto.RegisterConfigHandler(srv.Server, s)
	db.Init()
	web.AsyncStart(WEB_UI_PORT)
	// subscribe config watch topic, to continully receive config updated
	srv.Server.Subscribe(srv.Server.NewSubscriber(watch.WatchTopic, watch.Watcher))
	return nil
}

func (s *Config) Stop(srv *service.Service) error {
	logger.Info("To close mysql connection...")
	db.Uninit()
	return nil
}

func main() {
}
