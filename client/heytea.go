package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"oauth2/configs"
	"oauth2/initialize"
	"strconv"
	"strings"
	"time"
)

// 初始化一个cookie存储对象
// something-very-secret应该是一个你自己的密匙，只要不被别人知道就行
var store3 = sessions.NewCookieStore([]byte("something-very-secret"))

func main() {
	//获得配置对象
	Yaml := configs.InitConfig()
	initialize.Init(Yaml)

	r := gin.Default()

	// 注册中间件
	r.Use(MiddleWare3())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello 欢迎来到喜茶sso前台首页")
	})
	r.Any("/login",login3)
	r.POST("/submit",submit3)
	r.GET("/admin",TokenMiddleWare3(),admin3)
	r.GET("/logout",logout3)
	r.GET("/sso/refreshToken",refreshToken3)

	//加载模板文件目录
	r.LoadHTMLGlob("client/*")

	//监听端口默认为9093
	r.Run(":9093")
}

//独立登录授权
func login3(c *gin.Context) {
	w := c.Writer
	r := c.Request
	if c.Request.Method == "POST" {
		username := c.PostForm("username")
		password := c.PostForm("password")
		fmt.Println("username:",username)
		fmt.Println("password:",password)

		//获取sso统一登录令牌
		val := url.Values{}
		val.Add("grant_type", "client_credentials")
		val.Add("scope","all") // Set Add 都可以
		val.Add("redirect_uri", "http://localhost:9093/home")

		body := strings.NewReader(val.Encode())
		req, err := http.NewRequest(http.MethodPost, "http://localhost:9096/token", body)
		//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Content-Type","application/x-www-form-urlencoded")
		//req.BasicAuth()
		req.SetBasicAuth("test_client_1", "test_secret_1")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("resp str:",string(bs))
		m := make(map[string]interface{})
		err = json.Unmarshal(bs, &m)
		if err != nil {
			fmt.Println("Umarshal failed:", err)
			return
		}
		fmt.Println("m:", m)
		//设置token到session
		//　获取一个session对象，session-name是session的名字
		session, err := store3.Get(r, "session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 在session中存储值
		session.Values["token"] = m["access_token"]
		// 保存更改
		session.Save(r, w)

		c.Redirect(302, "/admin")
	}else {
		//独立登录页面
		//c.HTML(http.StatusOK, "login.html", gin.H{"name": "测试应用1", "address": "www.100txy.com"})
		access_token :=c.Query("access_token")
		if access_token == "" {
			log.Println("access_token获取失败！")
			c.Redirect(302, "/")
		}
		fmt.Println("登录token:",access_token)
		expires_in :=c.Query("expires_in")

		expiration, err := strconv.Atoi(expires_in)
		fmt.Println("expiration",expiration)
		//设置共享session
		initialize.RedisClient.Set("client_id", "aaaaaabbbbb", -1)
		initialize.RedisClient.Set("id_exampleOauthClient", access_token, -1)
		session, err := store3.Get(r, "sso-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 在session中存储值
		session.Values["token"] = access_token
		// 保存更改
		session.Save(r, w)

		c.Redirect(302, "/admin")
	}
}

func submit3(c *gin.Context) {
	name := c.DefaultQuery("name", "lily")
	c.String(200, fmt.Sprintf("hello %s\n", name))
}

//后台主页
func admin3(c *gin.Context) {
	// 取值
	req, _ := c.Get("request")
	fmt.Println("request:", req)//userId
	c.HTML(http.StatusOK, "admin.html", gin.H{"title": "heytea sso认证后台首页", "address": "client_id:exampleOauthClient","user_id":req})
}

//退出
func logout3(c *gin.Context) {
	w := c.Writer
	r := c.Request
	//session, err := store.Get(r, "session-name")
	session, err := store3.Get(r, "sso-session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token := session.Values["token"]
	fmt.Println("logout token:",token)
	// 删除
	// 将session的最大存储时间设置为小于零的数即为删除
	session.Options.MaxAge = -1
	session.Save(r, w)
	c.Redirect(302, "/")
	//c.Redirect(302, "http://localhost:9096/logout?redirect_uri=http%3a%2f%2flocalhost%3a9096%2fauthorize%3fclient_id%3dtest_client_1%26response_type%3dcode%26scope%3dall%26state%3dxyz%26redirect_uri%3dhttp%3a%2f%2flocalhost%3a9093%2flogin")
}


// 定义全局中间
func MiddleWare3() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		fmt.Println("中间件开始执行了")
		// 设置变量到Context的key中，可以通过Get()取
		c.Set("request", "中间件")
		status := c.Writer.Status()
		fmt.Println("中间件执行完毕", status)
		t2 := time.Since(t)
		fmt.Println("time:", t2)
	}
}

// 定义局部中间
func TokenMiddleWare3() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		w := c.Writer
		fmt.Println("局部中间件开始执行了")
		//session, err := store.Get(r, "session-name")//独立session
		session, err := store3.Get(r, "sso-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		token := session.Values["token"]
		fmt.Println("token：", token)
		if token != nil {
			fmt.Println("session clean")
			tokenMap := checkToken3(c, token.(string))
			fmt.Println("checkToken3 tokenMap2", tokenMap)
			userId := tokenMap.(map[string]interface{})["id"]
			c.Set("request", userId)
		}else {
			//c.Redirect(302, "/login")//独立登录
			c.Redirect(302, "https://dev-cas.heyteago.com/api/base-oauth2/oauth2.0/authorize?client_id=exampleOauthClient&response_type=token&scope=all&redirect_uri=http://sso.heytea.com/login")
		}
	}
}

func clenToken(c *gin.Context, oauth_token string) interface{} {
	//获取sso统一登录令牌
	val := url.Values{}
	val.Add("grant_type", "client_credentials")
	val.Add("scope","all") // Set Add 都可以
	val.Add("redirect_uri", "http://localhost:9093/home")

	body := strings.NewReader(val.Encode())
	req, err := http.NewRequest(http.MethodPost, "https://dev-cas.heyteago.com/api/base-oauth2/oauth2.0/profile?access_token=", body)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	//req.BasicAuth()
	req.SetBasicAuth("test_client_1", "test_secret_1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println("resp str:",string(bs))//invalid access token
	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		fmt.Println("Umarshal failed:", err)//Umarshal failed: invalid character 'i' looking for beginning of value
		//c.Redirect(302,"http://localhost:9096/authorize?client_id=test_client_2&response_type=code&scope=all&state=xyz&redirect_uri=http://localhost:9094/login")//authorization_code才需要到认证中心授权
		//删除session中token
		w := c.Writer
		r := c.Request
		session, err := store3.Get(r, "sso-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		//session.Values["token"] = access_token
		//token := oauth_token
		//fmt.Println("checkToken token:",token)
		// 删除
		// 将session的最大存储时间设置为小于零的数即为删除
		session.Options.MaxAge = -1
		session.Save(r, w)
		c.Redirect(302,"https://dev-cas.heyteago.com/api/base-oauth2/oauth2.0/authorize?client_id=exampleOauthClient&response_type=token&scope=all&redirect_uri=http://sso.heytea.com/login")
		//logout2(c)
	}
	fmt.Println("checkToken3 map:", m)//map: map[client_id:test_client_3 domain:http://localhost:8080 expires_in:7199 scope:all user_id:admin]
	return m
}

//效验sso分发的access_token有效性
func checkToken3(c *gin.Context, oauth_token string) interface{} {
	req, err := http.NewRequest(http.MethodGet, "https://dev-cas.heyteago.com/api/base-oauth2/oauth2.0/profile?access_token="+oauth_token, nil)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	//req.BasicAuth()
	//req.Header.Set("Authorization","Bearer "+oauth_token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println("resp str:",string(bs))//invalid access token
	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		fmt.Println("Umarshal failed:", err)//Umarshal failed: invalid character 'i' looking for beginning of value
		//c.Redirect(302,"http://localhost:9096/authorize?client_id=test_client_2&response_type=code&scope=all&state=xyz&redirect_uri=http://localhost:9094/login")//authorization_code才需要到认证中心授权
		//删除session中token
		w := c.Writer
		r := c.Request
		session, err := store3.Get(r, "sso-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		//session.Values["token"] = access_token
		//token := oauth_token
		//fmt.Println("checkToken token:",token)
		// 删除
		// 将session的最大存储时间设置为小于零的数即为删除
		session.Options.MaxAge = -1
		session.Save(r, w)
		c.Redirect(302,"https://dev-cas.heyteago.com/api/base-oauth2/oauth2.0/authorize?client_id=exampleOauthClient&response_type=token&scope=all&redirect_uri=http://sso.heytea.com/login")
		//logout2(c)
	}
	fmt.Println("checkToken3 map:", m)//map: map[client_id:test_client_3 domain:http://localhost:8080 expires_in:7199 scope:all user_id:admin]
	return m
}

//刷新token
func refreshToken3(c *gin.Context) {
	//获取恭喜session
	sessionStr,err := initialize.RedisClient.Get("test_client_1").Result()
	if err != nil {
		panic(err)
	}
	tokenMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(sessionStr), &tokenMap)
	if err != nil {
		fmt.Println("Umarshal failed:", err)
		return
	}
	fmt.Println("tokenMap:", tokenMap)
	refresh_token := tokenMap["refresh_token"]
	fmt.Println("refresh_token",refresh_token)

	//获取sso统一登录令牌
	val := url.Values{}
	val.Add("grant_type", "refresh_token")
	val.Add("refresh_token", refresh_token.(string))

	body := strings.NewReader(val.Encode())
	req, err := http.NewRequest(http.MethodPost, "http://localhost:9096/token", body)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	//req.BasicAuth()
	req.SetBasicAuth("test_client_1", "test_secret_1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("resp str:",string(bs))
	m := make(map[string]interface{})
	err = json.Unmarshal(bs, &m)
	if err != nil {
		fmt.Println("Umarshal failed:", err)
		return
	}
	fmt.Println("refresh m:", m)
}