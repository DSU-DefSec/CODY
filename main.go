package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/BurntSushi/toml"

	"github.com/google/uuid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type config struct {
	PlaylistId     []string
	VCloudUsername string
	VCloudPassword string
	VCloudAdmin    []string
	DeveloperKey   string
	DarkMode       bool
}

type lessonData struct {
	Id          string
	Title       string
	Description string
	VApp        string
	Video       string
	Slides      string
	PDF         string
}

const (
	ConfigFile = "cody.conf"
	vOrg       = "Defsec"
	vHref      = "https://vcloud.ialab.dsu.edu/api"
	vVDC       = "DefSec_Default"
	vInsecure  = false
)

var (
	codyConf       = &config{}
	globalPlaylist = []lessonData{}
)

func main() {
	configContent, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		log.Fatalln("error reading config file " + ConfigFile + ": " + err.Error())
	}
	if _, err := toml.Decode(string(configContent), &codyConf); err != nil {
		log.Fatalln("error decoding toml: " + err.Error())
	}
	if codyConf.VCloudUsername == "" || codyConf.VCloudPassword == "" || len(codyConf.VCloudAdmin) == 0 {
		fmt.Println("Missing options. Make sure you include vCloud settings and at least one playlistid.")
		return
	}

	// Initialize vCloud connection
	err = initVCD()
	if err != nil {
		panic(err)
	}

	//refreshPlaylistData()

	// Initialize Gin router
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")
	store := cookie.NewStore([]byte(uuid.New().String()))
	//store := cookie.NewStore([]byte("2314irj"))
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

	authRoutes := routes.Group("/")
	authRoutes.Use(AuthRequired)
	{
		authRoutes.GET("/", learn)
		authRoutes.GET("/learn", learn)
		authRoutes.GET("/learn/:id", lesson)
		authRoutes.GET("/deploy", deploy)
		authRoutes.GET("/deploy/ws", deployWS)
	}

	adminRoutes := routes.Group("/")
	adminRoutes.Use(AdminRequired)
	{
		adminRoutes.GET("/settings", settings)
		adminRoutes.GET("/refresh", refresh)
	}

	r.Run()
}

func findLesson(id string, lessons []lessonData) lessonData {
	var nilLesson lessonData
	if !validateName(id) {
		return nilLesson
	}
	for _, lesson := range lessons {
		if lesson.Id == id {
			return lesson
		}
	}
	return nilLesson
}

func learn(c *gin.Context) {
	c.HTML(http.StatusOK, "learn.html", pageData(c, gin.H{"lessons": globalPlaylist}))
}

func settings(c *gin.Context) {
	c.HTML(http.StatusOK, "settings.html", pageData(c, gin.H{"lessons": globalPlaylist}))
}

func refresh(c *gin.Context) {
	refreshPlaylistData()
	c.Redirect(http.StatusSeeOther, "/")
}

func refreshPlaylistData() {
	// No support for multiple playlists right now.
	globalPlaylist = retrievePlaylist(codyConf.PlaylistId[0])
}

func lesson(c *gin.Context) {
	id := c.Param("id")
	var lesson lessonData
	if lesson = findLesson(id, globalPlaylist); lesson.Id == "" {
		c.HTML(http.StatusOK, "learn.html", pageData(c, gin.H{"error": "Sorry, that lesson name isn't valid.", "lessons": globalPlaylist}))
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

//////////////////////
// Helper functions //
//////////////////////

func pageData(c *gin.Context, ginMap gin.H) gin.H {
	newGinMap := gin.H{}
	newGinMap["user"] = getUser(c)
	newGinMap["admin"] = isAdmin(newGinMap["user"].(string))
	newGinMap["dark"] = codyConf.DarkMode
	for key, value := range ginMap {
		newGinMap[key] = value
	}
	return newGinMap
}

func validateName(name string) bool {
	inputValidation := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	return inputValidation.MatchString(name)
}
