package tokenLimit

import (
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
	"testing"
	"time"
)

func TestTokenLimit(t *testing.T) {
	const (
		total = 100000
		rate  = 10
		burst = 10
	)
	//storeDev := redis.NewClient(&redis.Options{
	//	Addr:     "10.10.1.58:6379",
	//	Password: "o9i8u7y6",
	//})
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	mgr := NewTokenLimiterMgr(client)
	l := mgr.GetOrCreateTokenLimiter(rate, burst, "sdkDentalSdkMedicalRecordAdd")
	var allowed int
	for i := 0; i < total; i++ {
		time.Sleep(time.Millisecond * 1000)
		if l.Allow() {
			allowed++
			fmt.Printf("---------------allowed---------------:%v\n", allowed)
		} else {
			fmt.Printf("limit:%v\n", allowed)
		}
	}

}

func GetApi() {
	api := "http://localhost:9999/api/limit/test1"
	res, err := http.Get(api)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("get api fail\n")
	}
}

func Benchmark_Main(b *testing.B) {
	//for i := 0; i < b.N; i++ {
	for {
		time.Sleep(1 * time.Millisecond)
		GetApi()
	}
	//}
}

func Benchmark_Main2(b *testing.B) {
	//for i := 0; i < b.N; i++ {
	for {
		time.Sleep(1 * time.Second)
		GetApi2()
	}
	//}
}

func GetApi2() {
	api := "http://localhost:9999/index"
	res, err := http.Get(api)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("get api2 fail\n")
	} else {
		fmt.Printf("get api2 success\n")
	}
}
