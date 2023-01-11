package service

import (
	"context"
	"encoding/json"
	"errors"
	pb "github.com/BingguWang/hystrix-study/grpc_test/server/proto"
	"log"
	"math/rand"
)

type ServiceImpl struct {
	pb.UnimplementedScoreServiceServer
}

func (*ServiceImpl) AddScoreByUserID(ctx context.Context, in *pb.AddScoreByUserIDReq) (*pb.AddScoreByUserIDResp, error) {
	log.Println(ToJsonString(in))

	// 随机设置错误
	if rand.Int()%2 == 0 {
		return nil, errors.New("call AddScoreByUserID failed")
	}

	return &pb.AddScoreByUserIDResp{UserID: in.UserID}, nil
}

func ToJsonString(v interface{}) string {
	marshal, _ := json.Marshal(v)
	return string(marshal)
}
