package main

import (
	"github.com/kataras/iris"
	"github.com/gorilla/securecookie"
	"github.com/valyala/fasthttp"
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
)

var hashKey = []byte("very-secret")
var blockKey = []byte("0123456789123456")
var secure = securecookie.New(hashKey, blockKey)

var cookieName = "user"

var loginURL = "https://login.example.com/cas/login?service=%s"
var validateURL = "https://login.example.com/cas/validate?service=%s&ticket=%s"
var serviceURL = "http://localhost:8080/"

type IrisCas struct {
}

func GetUser(c *iris.Context) string {
	payload := c.GetCookie(cookieName)
	var decoded_value string
	secure.Decode(cookieName, payload, &decoded_value)
	return decoded_value
}

func SetUser(c *iris.Context, username string) {
	var cookie fasthttp.Cookie
	cookie.SetKey(cookieName)
	encoded_value, err := secure.Encode(cookieName, username)
	if err != nil {
		c.Log("Error encoding cookie %v", err)
	}	
	cookie.SetValue(encoded_value)
	//cookie.SetPath(basePath)
	c.SetCookie(&cookie)
}

func DeleteUser(c *iris.Context) {
	var cookie fasthttp.Cookie
	cookie.SetKey(cookieName)
	cookie.SetValue("")
	//cookie.SetPath(basePath)
	c.SetCookie(&cookie)
}

func (i *IrisCas) Serve(c *iris.Context) {
	user := GetUser(c)
	if user == "" {
		ticket := c.URLParam("ticket")
		c.Log("Ticket %s", ticket)
		if ticket != "" {
			res, err := http.Get(fmt.Sprintf(validateURL, serviceURL, ticket))
			if err != nil {
				c.Log(err.Error())
				return
			}

			casresponse, err := ioutil.ReadAll(res.Body)
			if err != nil {
				c.Log(err.Error())
				return
			}
			res.Body.Close()
			spli := strings.Split(string(casresponse), "\n")
			c.Log("First line %s", spli[0])
			if spli[0] == "yes" {
				c.Log("user validated %s", spli[1])
				SetUser(c, spli[1])
			} else {
				c.Log("Redirecting to login")
				c.Redirect(fmt.Printf(loginURL, serviceURL), 303)
				return
			}
		} else {
			c.Log("Redirecting to login no ticket")
			c.Redirect(fmt.Printf(loginURL, serviceURL), 303)
			return
		}
	}
	c.Next()
}
	

func main() {

	iris.Use(&IrisCas{})

	iris.Get("/", func (c *iris.Context) {
		c.Write("Hello World")
	})

	iris.Get("/logout", func (c *iris.Context) {
		DeleteUser(c)
		c.Redirect("https://login.imim.cloud/cas/logout?service=http://localhost:8080/", 303)
	})

	iris.Listen(":8080")
}
