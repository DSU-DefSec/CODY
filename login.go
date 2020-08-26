package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var userkey = "user"

func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
	}
	c.Next()
}

func AdminRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
	} else {
		if !isAdmin(user.(string)) {
			c.Redirect(http.StatusSeeOther, "/")
			c.Abort()
		}
	}
	c.Next()
}

func isAdmin(userName string) bool {
	for _, admin := range codyConf.VCloudAdmin {
		if userName == admin {
			return true
		}
	}
	return false

}

// login is a handler that parses a form and checks for specific data
func login(c *gin.Context) {
	session := sessions.Default(c)
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Validate form input
	if strings.Trim(username, " ") == "" || strings.Trim(password, " ") == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Username or password can't be empty 🙄"})
		return
	}

	// FETCH FROM IALAB lol
	err := vcloudAuth(username, password)
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Incorrect username or password."})
		return
	}

	// Save the username in the session
	session.Set(userkey, username)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(http.StatusSeeOther, "/")
}

func getUser(c *gin.Context) string {
	session := sessions.Default(c)
	return fmt.Sprintf("%s", session.Get(userkey))
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	session.Delete(userkey)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}
