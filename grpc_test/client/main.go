package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/BingguWang/hystrix-study/grpc_test/server/proto"
	"github.com/BingguWang/hystrix-study/grpc_test/server/service"
	"github.com/afex/hystrix-go/hystrix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
)

var (
	host = flag.String("host", "localhost", "")
	port = flag.String("port", "50051", "")
)

func main() {
	flag.Parsed()
	addr := net.JoinHostPort(*host, *port)

	// dial
	cc, err := grpc.DialContext(context.Background(), addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	// get client
	client := pb.NewScoreServiceClient(cc)

	// call service
	hystrix.ConfigureCommand("test", hystrix.CommandConfig{
		Timeout: 10, // 执行 command 的超时时间，单位ms,默认1000

		MaxConcurrentRequests: 100, // 最大并发量

		RequestVolumeThreshold: 5, // 一个统计窗口10秒内请求数量(貌似和10s没什么关系)，达到这个请求数量后才去判断是否要开启熔断，默认值是20

		SleepWindow: 50, //单位为毫秒 熔断器被打开后， SleepWindow 的时间就是指定过多久后去尝试服务是否可用了，默认是5000

		ErrorPercentThreshold: 20, // 错误百分比,请求数量大于等于 RequestVolumeThreshold 并且错误率到达这个百分比后就会启动熔断，默认是50
	})

	// 假设同时进行n个调用请求
	for i := 0; i < 50; i++ {
		hystrix.Do("test", func() error {
			rt, err := client.AddScoreByUserID(context.Background(), &pb.AddScoreByUserIDReq{UserID: 1})
			if err != nil {
				return err
			}
			fmt.Println(service.ToJsonString(rt))
			return nil
		}, func(err error) error {
			log.Println("fallback... err :", err.Error())
			return err
		})
	}
	/**
	可以发现，加了熔断器之后，熔断下， 调用方不会去调服务方，直接fail-fast返回error
	不过收到的还是error，所以我们需要考虑进行降级处理
	*/
}
