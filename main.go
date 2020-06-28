package main

import (
    "flag"
    "fmt"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

func init() {
	flag.StringVar(&vCloudUsername, "u", "", "Username for the vCloud account")
	flag.StringVar(&vCloudPassword, "p", "", "Password for the vCloud account")
	flag.StringVar(&vCloudAdmin, "a", "", "Username for the C.O.D.Y. admin")
	flag.Parse()
}

const (
	vOrg      = "Defsec"
	vHref     = "https://vcloud.ialab.dsu.edu/api"
	vVDC      = "DefSec_Default"
	vInsecure = false
)

// Web-Deploy API Endpoint
var vCloudUsername string
var vCloudPassword string
var vCloudAdmin string
var buttonArray map[string][]string = make(map[string][]string)

func main() {

	if vCloudUsername == "" || vCloudPassword == "" || vCloudAdmin == "" {
		fmt.Println("Missing parameters.")
		fmt.Println("Usage: ./cody -u username -p password -a admin")
		return
	}

	// Initialize vCloud connection
	initVCD()

	// reset Db (dev only reee)
	resetDB()

	// Initialize Gin router
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	store := cookie.NewStore([]byte(uuid.New().String()))
	r.Use(sessions.Sessions("codySession", store))

	// Routes
	routes := r.Group("/")
	{
		routes.GET("/login", func(c *gin.Context) {
			session := sessions.Default(c)
			user := session.Get(userkey)
			if user == nil {
				c.HTML(http.StatusOK, "login.html", nil)
			} else {
				c.Redirect(http.StatusSeeOther, "/")
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
			c.HTML(http.StatusOK, "about.html", pageData(c, gin.H{}))
		})
		internalRoutes.GET("/create", displayLesson)
		internalRoutes.POST("/create", createLesson)
	}

	r.Run()
}

///////////////////
// GET Endpoints //
///////////////////

func learn(c *gin.Context) {
	lessons := getEvents(10)
	c.HTML(http.StatusOK, "learn.html", pageData(c, gin.H{"lessons": lessons}))
}

func lesson(c *gin.Context) {
	vapp := c.Param("vapp")
	if !validateName(vapp) {
		lessons := getEvents(10)
		c.HTML(http.StatusOK, "learn.html", pageData(c, gin.H{"error": "Sorry, that lesson name isn't valid.", "lessons": lessons}))
		return
	}
	lesson, _ := getEvent("vapp", vapp)
	if lesson.Vapp == "" {
		lessons := getEvents(10)
		c.HTML(http.StatusOK, "learn.html", pageData(c, gin.H{"error": "Sorry, that lesson doesn't exist.", "lessons": lessons}))
		return
	}
	c.HTML(http.StatusOK, "lesson.html", pageData(c, gin.H{"lesson": lesson}))
}

func deploy(c *gin.Context) {
	c.HTML(http.StatusOK, "deploy.html", pageData(c, gin.H{"user": getUser(c)}))
}

func displayLesson(c *gin.Context) {
	c.HTML(http.StatusOK, "create.html", pageData(c, gin.H{"user": getUser(c)}))
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
	c.Redirect(http.StatusSeeOther, "/learn")
}


//////////////////////
// Helper functions //
//////////////////////

func pageData(c *gin.Context, ginMap gin.H) gin.H {
	newGinMap := gin.H{}
	newGinMap["user"] = getUser(c)
	newGinMap["admin"] = vCloudAdmin
	newGinMap["isAdmin"] = (newGinMap["user"] == newGinMap["admin"])
	for key, value := range ginMap {
		newGinMap[key] = value
	}
	return newGinMap
}

func validateName(name string) bool {
    inputValidation := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	return inputValidation.MatchString(name)
}
