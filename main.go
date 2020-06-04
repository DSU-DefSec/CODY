package main

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
    "flag"
    "fmt"
)

func init() {
	flag.StringVar(&webDeployAPI, "w", "", "WebDeploy Endpoint")
	flag.StringVar(&webDeployAPIPassword, "p", "", "WebDeploy Password")
	flag.Parse()
}

// Web-Deploy API Endpoint
var webDeployAPI string
var webDeployAPIPassword string
var buttonArray map[string][]string = make(map[string][]string)

func main() {

	if webDeployAPI == "" || webDeployAPIPassword == "" {
		fmt.Println("Missing parameters.")
		fmt.Println("Usage: ./cody -w endpoint -p password")
		return
	}
    fmt.Println("passsord is", webDeployAPIPassword)
	// reset Db (dev only reee)
	resetDB()

	// Initialize Gin router
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	r.Use(sessions.Sessions("cool_beans", sessions.NewCookieStore([]byte("secret"))))

	// Routes
	routes := r.Group("/")
	{
		routes.GET("/login", func(c *gin.Context) {
			session := sessions.Default(c)
			user := session.Get(userkey)
			if user == nil {
				c.HTML(http.StatusOK, "login.html", nil)
			} else {
				c.Redirect(http.StatusSeeOther, prevPath)
			}
			return
		})
		routes.POST("/login", login)
		routes.GET("/logout", logout)
	}

	internalRoutes := routes.Group("/")
	internalRoutes.Use(AuthRequired)
	{
		internalRoutes.GET("/", learn)
		internalRoutes.GET("/learn", learn)
		internalRoutes.GET("/learn/:vapp", lesson)
		internalRoutes.GET("/deploy", deploy)
		internalRoutes.GET("/deploy/ws", deployWS)
		internalRoutes.GET("/about", func(c *gin.Context) {
			c.HTML(http.StatusOK, "about.html", gin.H{"user": getUser(c)})
		})
	}

	apiRoutes := routes.Group("/api", gin.BasicAuth(gin.Accounts{
        // lol
		"admin": "password",
	}))
	{
		apiRoutes.POST("/learn", createLesson)
	}
	r.Run()
}

///////////////////
// GET Endpoints //
///////////////////

func learn(c *gin.Context) {
	lessons := getEvents(10)
	c.HTML(http.StatusOK, "learn.html", gin.H{"lessons": lessons, "user": getUser(c)})
}

func lesson(c *gin.Context) {
	vapp := c.Param("vapp")
	if !validateName(vapp) {
		lessons := getEvents(10)
		c.HTML(http.StatusOK, "learn.html", gin.H{"error": "Sorry, that lesson name isn't valid.", "lessons": lessons, "user": getUser(c)})
		return
	}
	lesson, _ := getEvent("vapp", vapp)
	if lesson.Vapp == "" {
		lessons := getEvents(10)
		c.HTML(http.StatusOK, "learn.html", gin.H{"error": "Sorry, that lesson doesn't exist.", "lessons": lessons, "user": getUser(c)})
		return
	}
	c.HTML(http.StatusOK, "lesson.html", gin.H{"lesson": lesson, "user": getUser(c)})
}

func deploy(c *gin.Context) {
	c.HTML(http.StatusOK, "deploy.html", gin.H{"user": getUser(c)})
}

////////////////////
// POST Endpoints //
////////////////////

func createLesson(c *gin.Context) {
	c.Request.ParseForm()
	title := c.Request.Form.Get("title")
	vapp := c.Request.Form.Get("vapp")
	description := c.Request.Form.Get("description")
	video := c.Request.Form.Get("video")
	if title == "" || vapp == "" {
		c.JSON(http.StatusBadRequest, "Bad request")
		return
	}
	err := addEvent(Event{
		Type:   10,
		Title:  title,
		Vapp:   vapp,
		Field1: description,
		Field2: video,
		Field3: "pdf", // placeholder
	})
    fmt.Println(err)
	c.JSON(http.StatusOK, "OK")
}


//////////////////////
// Helper functions //
//////////////////////

func getUser(c *gin.Context) Player {
	return Player{getUserName(c), 100} // arbitrary points
}

func validateName(name string) bool {
    inputValidation := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	return inputValidation.MatchString(name)
}
