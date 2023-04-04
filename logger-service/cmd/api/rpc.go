package main

import (
	"context"
	"logger/data"
)

type RpcServer struct {
}

type RpcPayload struct {
	Name string
	Data string
}

func (rpcSv *RpcServer) Log(pl RpcPayload, resp *string) error {
	clt := client.Database("logs").Collection("logs")

	if _, err := clt.InsertOne(context.TODO(), data.LogEntry{
		Name: pl.Name,
		Data: pl.Data,
	}); err != nil {
		println("Error write to Mongo: ", err)
		return err
	}

	*resp = "Processed payload via Rpc: " + pl.Name

	return nil
}
