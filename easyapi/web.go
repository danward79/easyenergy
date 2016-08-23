package easyapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/danward79/easyenergy/db"
)

//EasyClient ...
type EasyClient struct {
	user         string
	pass         string
	lastInterval string
	loggedIn     bool
	lastAction   time.Time
	client       http.Client
	db           *postgres.DBSession
}

//Jar ..
type Jar struct {
	sync.Mutex
	cookies map[string][]*http.Cookie
}

//SetCookies ...
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Lock()
	if _, ok := jar.cookies[u.Host]; ok {
		for _, c := range cookies {
			jar.cookies[u.Host] = append(jar.cookies[u.Host], c)
		}
	} else {
		jar.cookies[u.Host] = cookies
	}
	jar.Unlock()
}

//Cookies ...
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

//NewJarClient ...
func NewJarClient() *http.Client {
	return &http.Client{
		Jar: &Jar{
			cookies: make(map[string][]*http.Cookie),
		},
	}
}

//NewClient create a new easy energy client
func NewClient(user, pass, dbUser, dbPass, dbPath, dbName string) *EasyClient {

	client := NewJarClient()

	db, err := postgres.New(dbUser, dbPass, dbPath, dbName)
	if err != nil {
		log.Fatal(err)
	}

	ec := EasyClient{
		user:   user,
		pass:   pass,
		client: *client,
		db:     db,
	}

	return &ec
}

//GetCookie stores a session cookie
func (c *EasyClient) GetCookie() error {

	req, err := http.NewRequest("GET", "https://energyeasy.ue.com.au/login/index", nil)
	if err != nil {
		return fmt.Errorf("Error occured with forming cookie request. %v", err)
	}
	setHeaders(*req)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error occured with obtaining cookie. %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Error occured with obtaining cookie. %v, %v", err, resp.Status)
	}

	return nil
}

//Login does a login.
func (c *EasyClient) Login() error {

	form := url.Values{
		"login_email":    {c.user},
		"login_password": {c.pass},
		"submit":         {"Login"},
	}

	req, _ := http.NewRequest("POST", "https://energyeasy.ue.com.au/login_security_check", bytes.NewBufferString(form.Encode()))
	setHeaders(*req)
	req.Header.Set("Origin", "https://energyeasy.ue.com.au")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Referer", "https://energyeasy.ue.com.au/login/index")
	req.Header.Set("Content-Length", strconv.Itoa(len(form.Encode())))

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error occured with Authentication. %v", err)
	}
	defer resp.Body.Close()

	if c.loginSuccess() {
		return nil
	}

	return fmt.Errorf("Error occured with Authentication. %v, %v", err, resp.Status)
}

//loginSuccess used to check for successful login. As redirects don't return a status 200. A hack!
func (c *EasyClient) loginSuccess() bool {
	URL := "https://energyeasy.ue.com.au/electricityView/index"

	req, _ := http.NewRequest("GET", URL, nil)
	setHeaders(*req)

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true
	}
	return false
}

// Logout ...
func (c *EasyClient) Logout() error {

	req, _ := http.NewRequest("GET", "https://energyeasy.ue.com.au/login_security_check", nil)
	setHeaders(*req)
	req.Header.Set("Referer", "https://energyeasy.ue.com.au/electricityView/index")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error occured with Authentication. %v", err)
	}
	defer resp.Body.Close()

	return nil
}

//QueryDay returns a given day of Data, takes the day required.
//"https://energyeasy.ue.com.au/electricityView/period/day/1?_=1468026371738"
func (c *EasyClient) QueryDay(day int) (QueryResult, error) {

	var result QueryResult
	unixTime := time.Now().Unix()

	queryURL := fmt.Sprintf("https://energyeasy.ue.com.au/electricityView/period/day/%d?_=%d", day, unixTime)

	req, _ := http.NewRequest("GET", queryURL, nil)
	setHeaders(*req)

	resp, err := c.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {

		for {
			err := json.NewDecoder(resp.Body).Decode(&result)
			if err == io.EOF {
				break
			} else if err != nil {
				return result, err
			}
		}

		c.lastInterval = result.LatestInterval
		return result, nil
	}
	return result, errors.New("Error occured with day query.")
}

// Poll struct
type poll struct {
	Poll bool `json:"poll"`
}

//PollUpdatesAvailable ...
func (c *EasyClient) PollUpdatesAvailable() (bool, error) {

	// NOTE: below returns {"poll":false} if nothing available. true if new data available.
	url := fmt.Sprintf("https://energyeasy.ue.com.au/electricityView/latestData?lastKnownInterval=%s&_=%d", c.lastInterval, time.Now().Unix())

	req, _ := http.NewRequest("GET", url, nil)
	setHeaders(*req)
	req.Header.Set("Referer", "https://energyeasy.ue.com.au/electricityView/index")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("Error occured with checking poll updates. %v", err)
	}
	defer resp.Body.Close()

	p := &poll{}
	if resp.StatusCode == 200 {

		err := json.NewDecoder(resp.Body).Decode(&p)
		if err != nil {
			return false, err
		}

		if !p.Poll {
			return false, nil
		}

	}

	// NOTE: If above returns true, then web app polls below, every 5 seconds until true Then returns to view. Don't think this is necessary for this.
	//https://energyeasy.ue.com.au/electricityView/isElectricityDataUpdated?lastKnownInterval=2016-08-14:34
	return p.Poll, nil
}

//setHeaders sets general headers for all queries
func setHeaders(r http.Request) {
	r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36")
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Accept-Encoding", "gzip, deflate, sdch, br")
	r.Header.Set("Accept-Language", "en-US,en;q=0.8")
	r.Header.Set("Upgrade-Insecure-Requests", "1")
}
