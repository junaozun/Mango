package main

import (
	"github.com/go-redis/redis"
	"github.com/junaozun/mango/engine"
	"github.com/junaozun/mango/mgpool"
	"github.com/junaozun/mango/tokenLimit"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type User struct {
	Name string
}

func ApiLimitHandler() engine.HandleFunc {
	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	mgr := tokenLimit.NewTokenLimiterMgr(client)
	// 桶的大小为100，一秒钟放入放入100个令牌到桶中，即限流一秒请求限制在100个
	return func(c *engine.Context) {
		l := mgr.GetOrCreateTokenLimiter(10, 10, c.Path)
		if l.Allow() {
			c.Next()
			return
		}
		c.String(http.StatusTooManyRequests, "限流了!")
	}
}

func main() {
	r := engine.Default()
	r.Use(ApiLimitHandler())
	r.GET("/index", func(c *engine.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
		log.Println("http req")
	})
	r.GET("/index/su/suxuefeng", func(c *engine.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	r.GET("/user/name/hou", func(c *engine.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	r.GET("/redirect", func(c *engine.Context) {
		c.Redirect(http.StatusFound, "/htmlTemplate")
	})
	r.GET("/htmlTemplate", func(c *engine.Context) {
		u := &User{
			Name: "suxuefengAndHouwenwen",
		}
		c.HtmlTemplateGlob("login.html", u, "tpl/*.html")
	})
	r.GET("/downlaod", func(c *engine.Context) {
		c.File("tpl/test.xlsx")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *engine.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *engine.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *engine.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *engine.Context) {
			c.JSON(http.StatusOK, engine.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}
	v3 := r.Group("/v2")
	{
		v3.GET("/niss/:name", func(c *engine.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}
	p, _ := mgpool.NewPool(5)
	r.ANY("/pool", func(c *engine.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page33333333</h1>")
		p.Submit(func() {
			time.Sleep(5 * time.Second)
			log.Println("submit任务", rand.Int())
		})
	})
	apiLimit := r.Group("api/limit")
	{
		apiLimit.ANY("/test1", func(c *engine.Context) {
			c.String(http.StatusOK, "api limit success")
		})
	}

	r.Run(":9999")
}
