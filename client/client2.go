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
	"strings"
	"time"
)

// 初始化一个cookie存储对象
// something-very-secret应该是一个你自己的密匙，只要不被别人知道就行
var store2 = sessions.NewCookieStore([]byte("something-very-secret"))

func main() {
	//获得配置对象
	Yaml := configs.InitConfig()
	initialize.Init(Yaml)

	r := gin.Default()

	// 注册中间件
	r.Use(MiddleWare2())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello test_client_2")
	})
	r.Any("/login",login2)
	r.POST("/submit",submit2)
	r.GET("/admin",TokenMiddleWare2(),admin2)
	r.GET("/logout",logout2)
	r.GET("/sso/refreshToken",refreshToken2)

	//加载模板文件目录
	r.LoadHTMLGlob("client/*")

	//监听端口默认为9094
	r.Run(":9094")
}

func login2(c *gin.Context) {
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
		val.Add("redirect_uri", "http://localhost:9094/home")

		body := strings.NewReader(val.Encode())
		req, err := http.NewRequest(http.MethodPost, "http://localhost:9096/token", body)
		//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Content-Type","application/x-www-form-urlencoded")
		//req.BasicAuth()
		req.SetBasicAuth("test_client_2", "test_secret_2")
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
		session, err := store2.Get(r, "session-name")
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
		//c.HTML(http.StatusOK, "login.html", gin.H{"name": "测试应用2", "address": "www.100txy.com"})
		code :=c.Query("code")
		if code == "" {
			log.Println("code获取失败！")
			c.Redirect(302, "/")
		}
		//获取sso统一登录令牌
		val := url.Values{}
		val.Add("grant_type", "authorization_code")
		val.Add("code", code) // Set Add 都可以
		val.Add("scope", "all") // Set Add 都可以
		val.Add("redirect_uri", "http://localhost:9094/login")

		body := strings.NewReader(val.Encode())
		req, err := http.NewRequest(http.MethodPost, "http://localhost:9096/token", body)
		//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Content-Type","application/x-www-form-urlencoded")
		//req.BasicAuth()
		req.SetBasicAuth("test_client_2", "test_secret_2")
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
		//设置共享session
		initialize.RedisClient.Set("test_client_2", string(bs),-1)
		session, err := store2.Get(r, "sso2-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 在session中存储值
		session.Values["token"] = m["access_token"]
		// 保存更改
		session.Save(r, w)

		c.Redirect(302, "/admin")
	}
}

func submit2(c *gin.Context) {
	name := c.DefaultQuery("name", "lily")
	c.String(200, fmt.Sprintf("hello %s\n", name))
}

//主页
func admin2(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", gin.H{"title": "首页-雷小天社区", "address": "www.100txy.com"})
}

//退出
func logout2(c *gin.Context) {
	w := c.Writer
	r := c.Request
	session, err := store2.Get(r, "sso2-session-name")
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
	//c.Redirect(302, "/home")
	c.Redirect(302, "http://localhost:9096/logout?redirect_uri=http%3a%2f%2flocalhost%3a9096%2fauthorize%3fclient_id%3dtest_client_2%26response_type%3dcode%26scope%3dall%26state%3dxyz%26redirect_uri%3dhttp%3a%2f%2flocalhost%3a9094%2flogin")
}


// 定义全局中间
func MiddleWare2() gin.HandlerFunc {
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
func TokenMiddleWare2() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		w := c.Writer
		fmt.Println("局部中间件开始执行了")
		session, err := store2.Get(r, "sso2-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		token := session.Values["token"]
		refresh_token := session.Values["refresh_token"]
		fmt.Println("token：", token)
		fmt.Println("refresh_token：", refresh_token)
		if token != nil {
			tokenMap := checkToken2(c, token.(string))
			fmt.Println("tokenMap", tokenMap)
		}else {
			c.Redirect(302, "http://localhost:9096/authorize?client_id=test_client_2&response_type=code&scope=all&state=xyz&redirect_uri=http://localhost:9094/login")
		}
	}
}


//效验sso分发的access_token有效性
func checkToken2(c *gin.Context, oauth_token string) interface{} {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:9096/test", nil)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Type","application/x-www-form-urlencoded")
	//req.BasicAuth()
	req.Header.Set("Authorization","Bearer "+oauth_token)
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
		w := c.Writer
		r := c.Request
		session, err := store2.Get(r, "sso2-session-name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		token := oauth_token
		fmt.Println("checkToken token:",token)
		// 删除
		// 将session的最大存储时间设置为小于零的数即为删除
		session.Options.MaxAge = -1
		session.Save(r, w)
		fmt.Println("checkToken2 token:",token)
		c.Redirect(302,"http://localhost:9096/logout?redirect_uri=http%3a%2f%2flocalhost%3a9096%2fauthorize%3fclient_id%3dtest_client_2%26response_type%3dcode%26scope%3dall%26state%3dxyz%26redirect_uri%3dhttp%3a%2f%2flocalhost%3a9094%2flogin")
		//logout2(c)
	}
	fmt.Println("map:", m)//map: map[client_id:test_client_3 domain:http://localhost:8080 expires_in:7199 scope:all user_id:admin]
	return m
}

//刷新token
func refreshToken2(c *gin.Context) {
	//获取恭喜session
	sessionStr,err := initialize.RedisClient.Get("test_client_2").Result()
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
	req.SetBasicAuth("test_client_2", "test_secret_2")
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