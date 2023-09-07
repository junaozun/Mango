package main

import (
	"github.com/junaozun/mango/engin"
	"github.com/junaozun/mango/mgpool"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type User struct {
	Name string
}

func defineLog() engin.HandleFunc {
	return func(c *engin.Context) {
		log.Println("name is suxuefeng define func")
		c.Next()
	}
}

func defineLog2() engin.HandleFunc {
	return func(c *engin.Context) {
		log.Println("name is houwenwen define func")
		c.Next()
	}
}

func main() {
	r := engin.Default()
	r.Use(defineLog())
	r.Use(defineLog2())
	r.GET("/index", func(c *engin.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
		log.Println("http req")
	})
	r.GET("/index/su/suxuefeng", func(c *engin.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	r.GET("/user/name/hou", func(c *engin.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	r.GET("/redirect", func(c *engin.Context) {
		c.Redirect(http.StatusFound, "/htmlTemplate")
	})
	r.GET("/htmlTemplate", func(c *engin.Context) {
		u := &User{
			Name: "suxuefengAndHouwenwen",
		}
		c.HtmlTemplateGlob("login.html", u, "tpl/*.html")
	})
	r.GET("/downlaod", func(c *engin.Context) {
		c.File("tpl/test.xlsx")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *engin.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *engin.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *engin.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *engin.Context) {
			c.JSON(http.StatusOK, engin.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}
	v3 := r.Group("/v2")
	{
		v3.GET("/niss/:name", func(c *engin.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}
	p, _ := mgpool.NewPool(5)
	r.ANY("/pool", func(c *engin.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page33333333</h1>")
		p.Submit(func() {
			time.Sleep(5 * time.Second)
			log.Println("submit任务", rand.Int())
		})
	})
	r.Run(":9999")
}
