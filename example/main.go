package main

import (
	"log"
	"time"

	"github.com/doabit/rock"
	"github.com/doabit/rock/example/config"
	"github.com/doabit/rock/example/render"
)

func main() {
	app := rock.New()
	app.Use(Logger())
	app.HTMLRender(render.Default())
	config.Setup(app)
	// app.LoadHTMLGlob("templates/*")
	app.Static("/assets", "./static")
	app.Get("/", Home)
	app.Get("/posts/:id", Post)
	api := app.Group("/api")
	api.Use(onlyForApi())
	{
		api.Get("/home", ApiIndex)
		v1 := api.Group("/v1")
		{
			v1.Get("/home", ApiIndex)
		}
	}

	admin := app.Group("/admin")
	admin.Use(auth())
	// app.GetHTMLRender().SetViewDir("./tem/")
	// app.GetHTMLRender().SetViewDir("./template/")
	{
		app.GetHTMLRender().SetViewDir("./tem/")
		admin.Get("/login", AdminLogin)
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}
}

func Post(c rock.Context) {
	log.Printf("query from is %s %d", c.Query("from"), c.QueryInt("cid"))
	c.String(200, "post id is %s", c.Param("id"))
}

func Home(c rock.Context) {
	// c.JSON(200, rock.H{"msg": "ok"})
	c.HTML("home")
}

// admin
func AdminLogin(c rock.Context) {
	log.Println("admin auth action")
	// c.JSON(http.StatusOK, rock.H{"msg": "admin login"})
	c.HTML("admin/login")
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
		log.Println("auth before")
		c.Next()
		log.Println("auth after")
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
