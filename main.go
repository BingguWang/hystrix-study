package main

import (
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/gin-gonic/gin"
	"gopkg.in/resty.v1"
	"net/http"
	"time"
)

func server() {
	r := gin.Default()
	start := time.Now()
	r.GET("/a", func(ctx *gin.Context) {
		if time.Since(start) < 1000*time.Millisecond { // 前1s收到的请求都返回错误
			ctx.String(http.StatusInternalServerError, "fail")
			return
		}
		if time.Since(start).Milliseconds() > 9000 && time.Since(start).Milliseconds() < 10000 {
			ctx.String(http.StatusInternalServerError, "fail")
			return
		}
		if time.Since(start).Milliseconds() > 13000 && time.Since(start).Milliseconds() < 14000 {
			ctx.String(http.StatusInternalServerError, "fail")
			return
		}
		ctx.String(http.StatusOK, "ok")
	})
	if err := r.Run(":8080"); err != nil {
		fmt.Println(err.Error())
	}
}
func main() {
	go server()

	// 配置断路器,一个commandName就是一个熔断器,命名可以是一个域名也可以是一个具体的方法等
	hystrix.ConfigureCommand("test", hystrix.CommandConfig{
		Timeout: 10, // 执行 command 的超时时间，单位ms,默认1000

		MaxConcurrentRequests: 100, // 最大并发量

		RequestVolumeThreshold: 5, // 一个统计窗口10秒内请求数量(貌似和10s没什么关系)，达到这个请求数量后才去判断是否要开启熔断，默认值是20

		SleepWindow: 500, //单位为毫秒 熔断器被打开后， SleepWindow 的时间就是指定过多久后去尝试服务是否可用了，默认是5000

		// 错误百分比,请求数量大于等于 RequestVolumeThreshold 并且错误率到达这个百分比后就会启动熔断，默认是50
		ErrorPercentThreshold: 20,
	})

	// 模拟客户端请求, 每个请求间隔100ms
	for i := 0; i < 25; i++ {
		/*
			func Do(name string, run runFunc, fallback fallbackFunc) error
				run为业务逻辑函数，其结构为func() error
				fallback为失败回调函数，其结构为func(err error) error
		*/
		_ = hystrix.Do("test", func() error { // 业务逻辑
			resp, _ := resty.New().R().Get("http://localhost:8080/a") // 发起Get请求
			if resp.IsError() {
				return fmt.Errorf("err code: %s", resp.Status())
			}
			return nil
		}, func(err error) error { // 失败回调函数
			fmt.Println("fallback err: ", err)
			return err
		})
		//time.Sleep(100 * time.Millisecond)
		time.Sleep(1000 * time.Millisecond)
	}
}

/*
结果分析：
	RequestVolumeThreshold: 10
		请求数量要先到达到RequestVolumeThreshold && 窗口内的错误百分比达到 ErrorPercentThreshold

前2个请求错误，
再经过8个请求，达到了10个请求，满足RequestVolumeThreshold，2/10=20%也达到了ErrorPercentThreshold，开启熔断，不管服务是否可用，不放行任何的请求
然后就会看到后面的请求fallback err:  hystrix: circuit open
因为SleepWindow是500ms, 也就是5个请求的时间，所以只会看到5次，500ms后会去看服务是否可用，可用就关闭熔断

*/
