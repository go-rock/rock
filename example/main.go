package main

import (
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
	{
		api.Get("/home", ApiIndex)
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
