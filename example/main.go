package main

import (
	"log"
	"time"

	"github.com/doabit/rock"
)

func main() {
	app := rock.New()
	app.Get("/", Home)
	app.Get("/posts/:id", Post)

	admin := app.Group("/admin")
	{
		admin.Get("/login", AdminLogin)
	}

	api := app.Group("/api")
	// api.Use(Logger())
	{
		api.Get("/home", ApiIndex)
		v1 := api.Group("v1")
		{
			v1.Get("/home", ApiIndex)
		}
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
	c.JSON(200, rock.H{"msg": "admin login"})
}

// Api

func ApiIndex(c rock.Context) {
	c.JSON(200, rock.H{"msg": "api v1 index"})
}

// middleware

// func onlyForApi() rock.HandlerFunc {
// 	return func(c rock.Context) {
// 		// Start timer
// 		t := time.Now()
// 		// if a server error occurred
// 		c.Fail(500, "Internal Server Error")
// 		// Calculate resolution time
// 		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
// 	}
// }

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
