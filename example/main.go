package main

import (
	"log"
	"net/http"
	"time"

	"github.com/doabit/rock"
)

func main() {
	app := rock.New()
	app.Use(Logger())

	app.Get("/", Home)
	app.Get("/posts/:id", Post)

	api := app.Group("/api")
	api.Use(onlyForApi())
	{
		api.Get("/home", ApiIndex)
		v1 := api.Group("v1")
		{
			v1.Get("/home", ApiIndex)
		}
	}

	admin := app.Group("/admin")
	admin.Use(auth())
	{
		admin.Get("/login", AdminLogin)
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}

func Post(c rock.Context) {
	c.String(200, "post id is %s", c.Param("id"))
}

func Home(c rock.Context) {
	c.JSON(200, rock.H{"msg": "ok"})
}

// admin
func AdminLogin(c rock.Context) {
	log.Println("admin auth action")
	c.JSON(http.StatusOK, rock.H{"msg": "admin login"})
}

// Api
func ApiIndex(c rock.Context) {
	c.JSON(200, rock.H{"msg": "api v1 index"})
}

// middlewares
func onlyForApi() rock.HandlerFunc {
	return func(c rock.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("Api only code [%d] %s in %v for group api", c.StatusCode(), c.Request().RequestURI, time.Since(t))
	}
}

func auth() rock.HandlerFunc {
	return func(c rock.Context) {
		// t := time.Now()
		log.Println("auth before")
		c.Next()
		log.Println("auth after")
		// c.Fail(500, "Internal Server Error")
	}
}

func Logger() rock.HandlerFunc {
	return func(c rock.Context) {
		// Start timer
		t := time.Now()
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode(), c.Request().RequestURI, time.Since(t))
	}
}
