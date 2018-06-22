package main

import (
	"os"
	"time"

	cmn "github.com/moxiaomomo/configcenter/common"
	proto "github.com/moxiaomomo/configcenter/proto"

	"github.com/moxiaomomo/configcenter/logger"
	"golang.org/x/net/context"
)

func testCreateConfig(client proto.ConfigClient, ctx context.Context, name, path string) {
	data := `{"host":"xxx","port":3306,"db":"vip","table":"vip_order"}`
	resp, err := client.Create(ctx, &proto.CreateRequest{
		ConfigSet: &proto.ConfigSet{
			Name:    name,
			Path:    path,
			Version: "1.0",
			Comment: "just4test",
			ChangeSet: &proto.ChangeSet{
				Timestamp: time.Now().Unix(),
				Data:      data,
				Checksum:  cmn.Md5(data),
				Source:    "rpc_client",
				Format:    "json",
			},
		},
	})
	if err != nil {
		logger.Errorf("[Client]create config error: %v\n", err)
	} else {
		logger.Infof("[Client]create config response: %d\n", resp.GetResp())
	}
}

func testUpdateConfig_DBUfile(client proto.ConfigClient, ctx context.Context, name, path string) {
	data := `{"host":"10.11.1.","port":3306,"db":"u115","user":"u115a","pwd":"Y1y1w5u115ApW"}`
	resp, err := client.Update(ctx, &proto.UpdateRequest{
		ConfigSet: &proto.ConfigSet{
			Name:    name,
			Path:    path,
			Version: "1.0",
			Comment: "to update database user_file auth info.",
			ChangeSet: &proto.ChangeSet{
				Timestamp: time.Now().Unix(),
				Data:      data,
				Checksum:  cmn.Md5(data),
				Source:    "rpc_client",
				Format:    "json",
			},
		},
	})
	if err != nil {
		logger.Errorf("[Client]create config error: %v\n", err)
	} else {
		logger.Infof("[Client]create config response: %d\n", resp.GetResp())
	}
}

func testReadConfig(client proto.ConfigClient, ctx context.Context, name, path string) {
	resp, err := client.Read(ctx, &proto.ReadRequest{
		Name: name,
		Path: "/path/to/notfound",
	})
	if err != nil {
		logger.Errorf("[Client]error: %v\n", err)
		os.Exit(1)
	}
	logger.Infof("[Client]info: resp::%d path:%s\n", resp.GetResp(), resp.GetConfigSet().GetPath())

	resp, err = client.Read(ctx, &proto.ReadRequest{
		Name: name,
		Path: path,
	})
	if err != nil {
		logger.Errorf("[Client]error: %v\n", err)
		os.Exit(1)
	}
	logger.Infof("[Client]info: resp::%d path:%s\n", resp.GetResp(), resp.GetConfigSet().GetPath())
}

func testUpdateConfig(client proto.ConfigClient, ctx context.Context, name, path string) {
	data := `{"host":"xxx.xxx","port":3306,"db":"vip2","table":"vip_order2"}`
	resp, err := client.Update(ctx, &proto.UpdateRequest{
		ConfigSet: &proto.ConfigSet{
			Name:    name,
			Path:    path,
			Version: "1.0",
			Comment: "just4test, to update",
			Status:  1,
			ChangeSet: &proto.ChangeSet{
				Timestamp: time.Now().Unix(),
				Data:      data,
				Checksum:  cmn.Md5(data),
				Source:    "rpc_client",
				Format:    "json",
			},
		},
	})
	if err != nil {
		logger.Errorf("[Client]update config error: %v\n", err)
	} else {
		logger.Infof("[Client]update config response: %d\n", resp.GetResp())
	}
}

func testDeleteConfig(client proto.ConfigClient, ctx context.Context, name, path string) {
	resp, err := client.Delete(ctx, &proto.DeleteRequest{
		ConfigSet: &proto.ConfigSet{
			Name:    name,
			Path:    path,
			Version: "1.0",
			Comment: "just4test, to delete",
		},
	})
	if err != nil {
		logger.Errorf("[Client]delete config error: %v\n", err)
	} else {
		logger.Infof("[Client]delete config response: %d\n", resp.GetResp())
	}
}

func testSearchConfig(client proto.ConfigClient, ctx context.Context, name, path string) {
	resp, err := client.Search(ctx, &proto.SearchRequest{
		Name: name,
		Path: path,
	})
	if err != nil {
		logger.Errorf("[Client]search error: %v\n", err)
		os.Exit(1)
	}
	logger.Infof("[Client]search info: resp::%d len:%d\n", resp.GetResp(), len(resp.GetConfigs()))
	if len(resp.GetConfigs()) > 0 {
		logger.Infof("[Client]search info: path:%s\n", resp.GetConfigs()[0].GetPath())
	}
}

func testWatchConfig(client proto.ConfigClient, ctx context.Context, name, path string) {
	resp, err := client.Watch(ctx, &proto.WatchRequest{
		Name:    name,
		Path:    path,
		Version: "1.0",
	})
	if err != nil {
		logger.Errorf("[Client]watch error: %v\n", err)
		os.Exit(1)
	}
	for {
		watchRsp := proto.WatchResponse{}
		if err := resp.RecvMsg(&watchRsp); err != nil {
			logger.Errorf("[Client]watch error: %v\n", err)
			break
		}
		logger.Infof("[Client]watch result: %v\n", watchRsp)
	}
}

func main() {
	// cl := proto.NewConfigClient("micro.frame.srv.config", client.DefaultClient)
	// ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	// testName := "CONFIG:DB"
	// testPath := "/dev/mysql/ufile"

	// //  testCreateConfig(cl, ctx, testName, testPath)
	// testReadConfig(cl, ctx, testName, testPath)
	//  testUpdateConfig(cl, ctx, testName, testPath)
	//	testCreateConfig(cl, ctx, testName, testPath)
	//testReadConfig(cl, ctx, testName, testPath)
	//	time.Sleep(1)
	//	testDeleteConfig(cl, ctx, testName, testPath)
	//	testReadConfig(cl, ctx, testName, testPath)

	//testUpdateConfig_DBUfile(cl, ctx, testName, testPath)
	// testSearchConfig(cl, ctx, testName, testPath)

	//testWatchConfig(cl, ctx, testName, testPath)
}
