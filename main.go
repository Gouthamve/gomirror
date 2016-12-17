package main

import "github.com/labstack/echo"

func main() {
	e := echo.New()
	e.Static("/public", "public")
	e.File("/", "public/index.html")
	e.Logger.Fatal(e.Start(":1323"))
}
