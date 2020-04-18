package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/contrib/sessions"
)
const (
	userkey = "user"
)
func main() {

    // Initialize
	r := gin.Default()
    r.LoadHTMLGlob("templates/*")
    r.Static("/assets", "./assets")
	r.Use(sessions.Sessions("mysession", sessions.NewCookieStore([]byte("secret"))))

    // Routes
	r.GET("/", func(c *gin.Context) {
    	c.Redirect(http.StatusSeeOther, "/login")})
	r.GET("/login", func(c *gin.Context) {
        c.HTML(http.StatusOK, "login.html", nil)})
	r.POST("/login", login)
	r.GET("/logout", logout)

	maize := r.Group("/maize")
	maize.Use(AuthRequired)
	{
		maize.GET("/", compete)
		maize.GET("/compete", compete)
		maize.GET("/upcoming", upcoming)
	}
	r.Run()
}

func compete(c *gin.Context) {
    c.HTML(http.StatusOK, "compete.html", nil)
}

func upcoming(c *gin.Context) {
    c.HTML(http.StatusOK, "upcoming.html", nil)
}
