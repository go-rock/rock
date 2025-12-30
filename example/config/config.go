package config

import (
	"fmt"

	render "github.com/go-rock/rock-pongo2"

	"github.com/fsnotify/fsnotify"
	"github.com/go-rock/rock"
	"github.com/spf13/viper"
)

const (
	THEME_PATH = "./themes/views/"
)

var Config = viper.New()

func Setup(app *rock.App) {
	LoadConfig(app)
}

func LoadConfig(app *rock.App) {
	Config.SetConfigFile("./config.json")

	if err := Config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("没找到")
		} else {
			fmt.Println("找到, 出错")
		}
	}
	app.RegisterView(render.New(render.ViewConfig{
		ViewDir:   THEME_PATH + Config.GetString("theme") + "/",
		Extension: ".html",
	}))
	// app.GetHTMLRender().SetViewDir(THEME_PATH + Config.GetString("theme") + "/")
	Config.WatchConfig()
	Config.OnConfigChange(func(e fsnotify.Event) {
		// app.RegisterView(render.New(render.ViewConfig{
		// 	ViewDir:   THEME_PATH + Config.GetString("theme") + "/",
		// 	Extension: ".html",
		// }))
		app.GetView().Engine.SetViewDir(THEME_PATH + Config.GetString("theme") + "/")
		// app.GetHTMLRender().SetViewDir(THEME_PATH + Config.GetString("theme") + "/")
	})
}

func Installed() bool {
	return Config.GetBool("installed")
}

func SetConfig(key string, value interface{}) error {
	Config.Set(key, value)
	return Config.WriteConfig()
}
