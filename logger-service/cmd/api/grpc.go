package main

import (
	"fmt"
	"log"
	"logger/data"
	"logger/logs"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
}

func (app *Config) gRpcListen() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Error listen to gRpc: %v", err)
	}

	sv := grpc.NewServer()

	logs.RegisterLogServiceServer(sv, &LogServer{})

	if err = sv.Serve(ln); err != nil {
		log.Fatalf("Error serve gRpc server: %v", err)
	}
}

func (logSv *LogServer) WriteLog(ctx context.Context, req *logs.LogReq) (*logs.LogResp, error) {
	logEntr := req.GetLogEntry()

	if err := data.Insert(data.LogEntry{
		Name: logEntr.Name,
		Data: logEntr.Data,
	}); err != nil {
		return &logs.LogResp{
			Result: "failed",
		}, err
	}

	return &logs.LogResp{
		Result: "done",
	}, nil
}
