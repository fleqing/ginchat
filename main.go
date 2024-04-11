package main

import (
	"ginchat/router"

	"github.com/spf13/viper"
)

func main() {
	r := router.Router()
	r.Run(viper.GetString("port.server")) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
