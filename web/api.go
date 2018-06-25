package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/moxiaomomo/configcenter/logger"
	proto "github.com/moxiaomomo/configcenter/proto"
	hdl "github.com/moxiaomomo/configcenter/web/handler"
	"google.golang.org/grpc"
)

func NewRouter() http.Handler {
	handler := mux.NewRouter()
	handler.HandleFunc("/", hdl.Index)
	handler.HandleFunc("/index", hdl.Index)

	handler.HandleFunc("/audit/log", hdl.AuditLog)

	handler.HandleFunc("/config/create", hdl.ConfigCreate)

	handler.HandleFunc("/config/index", hdl.ConfigIndx)
	handler.HandleFunc("/config/query", hdl.ConfigQuery)
	handler.HandleFunc("/config/edit", hdl.ConfigEdit)
	handler.HandleFunc("/config/update", hdl.ConfigUpdate)

	handler.HandleFunc("/config/search", hdl.ConfigSearch)
	return handler
}

// Start start server
func Start(lbhost string, curhost string) error {
	logString := fmt.Sprintf("config-web start and listen on %s", curhost)
	logger.Info(logString)

	conn, err := grpc.Dial(lbhost, grpc.WithInsecure())
	if err != nil {
		logger.Error("failed to start server as rpc-client initiation failed.")
	}
	hdl.Init(
		"exts/configsrv/web/templates",
		proto.NewConfigClient(conn),
	)
	return http.ListenAndServe(curhost, NewRouter())
}

// AsyncStart aync-start server
func AsyncStart(lbhost string, curhost string) {
	go Start(lbhost, curhost)
}
