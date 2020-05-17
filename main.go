package main

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"time"
)

// Web-Deploy API Endpoint
var webDeployAPI string = "http://192.168.1.100/api"
var buttonArray map[string][]string = make(map[string][]string)

func main() {

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
		internalRoutes.GET("/", compete)
		internalRoutes.GET("/compete", compete)
		internalRoutes.GET("/compete/:title", competition)
		internalRoutes.GET("/compete/:title/ws", func(c *gin.Context) {
			competitionEndpoint(c)
		})
		internalRoutes.GET("/learn", learn)
		internalRoutes.GET("/learn/:vapp", lesson)
		internalRoutes.GET("/deploy", deploy)
		internalRoutes.GET("/deploy/ws", deployWS)
		internalRoutes.GET("/leaderboard", leaderboard)
		internalRoutes.GET("/about", func(c *gin.Context) {
			c.HTML(http.StatusOK, "about.html", gin.H{"user": getUser(c)})
		})
	}

	apiRoutes := routes.Group("/api", gin.BasicAuth(gin.Accounts{
        // lol
		"admin": "password",
	}))
	{
		apiRoutes.POST("/compete", createCompetition)
		apiRoutes.POST("/learn", createLesson)
	}
	r.Run()
}

///////////////////
// GET Endpoints //
///////////////////

func compete(c *gin.Context) {
	competitions := getEvents(01)
	c.HTML(http.StatusOK, "compete.html", gin.H{"competitions": competitions, "user": getUser(c)})
}

func competition(c *gin.Context) {
	title := c.Param("title")
	if validateCompetition(c, title) {
    	competition, _ := getEvent("title", title)
    	c.HTML(http.StatusOK, "competition.html", gin.H{"competition": competition, "user": getUser(c)})
    }
}

func validateCompetition(c *gin.Context, title string) bool {
	if !validateName(title) {
    	competitions := getEvents(01)
    	c.HTML(http.StatusOK, "compete.html", gin.H{"error": "Invalid competition name", "competitions": competitions, "user": getUser(c)})
		return false
	}
	competition, _ := getEvent("title", title)
	if competition.Vapp == "" {
		competitions := getEvents(01)
		c.HTML(http.StatusOK, "compete.html", gin.H{"error": "Sorry, that competition doesn't exist.", "competitions": competitions, "user": getUser(c)})
		return false
	}
    return true
}

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

func leaderboard(c *gin.Context) {
	c.HTML(http.StatusOK, "leaderboard.html", gin.H{"leaderboard": getLeaderboard(), "user": getUser(c)})
}

////////////////////
// POST Endpoints //
////////////////////

func createLesson(c *gin.Context) {
	c.Request.ParseForm()
	title := c.Request.Form.Get("title")
	vapp := c.Request.Form.Get("vapp")
	if title == "" || vapp == "" {
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	addEvent(Event{
		Type:   10,
		Title:  title,
		Vapp:   vapp,
		Field1: "description",
		Field2: "youtube",
		Field3: "pdf",
	})
	c.JSON(http.StatusOK, nil)
}

func createCompetition(c *gin.Context) {
	c.Request.ParseForm()
	start := time.Now().String() // should parse datetime optionally
	end := "2020-11-10 23:00:00 +0000 UTC"
	private := false

	title := c.Request.Form.Get("title")
	vapp := c.Request.Form.Get("vapp")
	kind := c.Request.Form.Get("kind")
	owner := c.Request.Form.Get("owner")
	if title == "" || vapp == "" || kind == "" || owner == "" {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	err := addEvent(Event{
		Type:   01,
		Kind:   kind,
		Title:  title,
		Vapp:   vapp,
		Field1: start,
		Field2: end,
        Field3: owner,
		Switch: private,
	})
    if err != nil {
    	c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
    }
	c.JSON(http.StatusOK, nil)
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
