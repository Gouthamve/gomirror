package main

import (
	"fmt"
	"os"

	"github.com/gouthamve/gomirror/handlers"
	"github.com/gouthamve/gomirror/util"
	"github.com/labstack/echo"
)

func main() {
	fmt.Println(os.TempDir())
	e := echo.New()
	e.Static("/public", "public")
	e.File("/", "public/index.html")
	e.POST("/index", handlers.IndexFace)
	e.POST("/detect", handlers.DetectFace)
	//util.DeleteCollection("nhack1")
	//util.DeleteCollection("nhack2")
	//util.CreateCollection("nhack1")
	go util.TwitterStream("nhack")
	defer util.DBClose()
	e.Logger.Fatal(e.Start(":1323"))
}
