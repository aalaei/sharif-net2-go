package main

import (
	"fmt"
	// "io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"regexp"
	"strconv"
	"context"
  	"net"

	ptime "github.com/yaa110/go-persian-calendar"
	"github.com/PuerkitoBio/goquery"
)
var net2_url string="https://net2.sharif.edu/%s"

var bw_url string="https://bw.ictc.sharif.edu/login"

// net_headers := {
//     'Referer': 'https://net.sharif.edu/',
//     'Host': 'net.sharif.edu',
//     'Accept-Language' : 'en-US,en;q=0.9,fa;q=0.8',
//     'User-Agent':'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36',
//     'X-Requested-With':'XMLHttpRequest'
// }
type credentials struct {
	username string
	password  string
}
func PostForm2(httpclient http.Client, url string, data url.Values) (resp *http.Response, err error) {
	var contentType string = "application/x-www-form-urlencoded"
	var method string ="POST"
	data_io :=strings.NewReader(data.Encode())
    res0, err := httpclient.Get(url)
	if err != nil {
        return nil, fmt.Errorf("got error %s", err.Error())
    }
	if res0.StatusCode != 200 {
		return nil, fmt.Errorf("got error %s", err.Error())
	}
	var new_cookie=strings.Split(strings.Split(res0.Header.Get("Set-Cookie"), ";")[0], "=")
	req, err := http.NewRequest(method, url, data_io)
    if err != nil {
        return nil, fmt.Errorf("got error %s", err.Error())
    }
    // req.Header.Set("user-agent", "golang application")
    req.Header.Set("Content-Type", contentType)
	req.AddCookie(&http.Cookie{Name: new_cookie[0], Value: new_cookie[1]})
	req.Header.Add("Host", "bw.ictc.sharif.edu")	
    response, err := httpclient.Do(req)
    if err != nil {
        return nil,fmt.Errorf("got error %s", err.Error())
    }
    // defer response.Body.Close()
    return response,nil
}

func net2_login(httpclient http.Client, cr credentials){
	net2_status_url:=fmt.Sprintf(net2_url, "status")
    res, err:=httpclient.Get(net2_status_url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s\n", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalf("Parse HTML error: %s\n", err)
	}
	
	page_title:=doc.Find("title").Text()
	if page_title=="logout"{
		fmt.Println("Already Connected!")
	} else {
		net2_login_url:=fmt.Sprintf(net2_url, "login")
		cr := url.Values{
			"username": {cr.username},
			"password": {cr.password},
		}
		res2, err:=httpclient.PostForm(net2_login_url, url.Values(cr))
		if err != nil {
			log.Fatal(err)
		}
		defer res2.Body.Close()
		if res2.StatusCode != 200 {
			fmt.Println("Not able to login!")
			log.Fatalf("status code error: %d %s\n", res2.StatusCode, res2.Status)
		}
		// Load the HTML document
		doc2, err := goquery.NewDocumentFromReader(res2.Body)
		if err != nil {
			log.Fatalf("Parse HTML error: %s\n", err)
		}
		
		page2_title:=doc2.Find("title").Text()
		if strings.Contains(page2_title,"mikrotik"){
			fmt.Println("Done :)")
		}else{
			fmt.Println("Incorrect password")
		}
		
	}
}
func check_bw(httpclient http.Client, cr credentials){
	c:=url.Values{
		"normal_username": {cr.username}, 
		"normal_password": {cr.password},
	}
	res_bw,err:=PostForm2(httpclient, bw_url,c)
    // res_bw,err:=http.Post(bw_url,"application/x-www-form-urlencoded",strings.NewReader(c2))
	// res_bw,err:=http.PostForm(bw_url,c)
	if err != nil {
		log.Fatal(err)
	}

	defer res_bw.Body.Close()
	if res_bw.StatusCode != 200 {
		fmt.Println("Not connected!")
		log.Fatalf("status code error: %d %s\n", res_bw.StatusCode, res_bw.Status)
	}
	// body, err := io.ReadAll(res_bw.Body)
	if err != nil {
		log.Fatalf("Parse HTML error: %s\n", err)
	}
	// fmt.Println(string(body))
	// Load the HTML document
	doc2, err := goquery.NewDocumentFromReader(res_bw.Body)
	if err != nil {
		log.Fatalf("Parse HTML error: %s\n", err)
	}
	var remaining string=""
	doc2.Find("script").Each(func(i int, s *goquery.Selection){
		script := s.Text()

		re := regexp.MustCompile(`باقی مانده\', value: [0-9]+\.[0-9]*`)
		match:=re.FindString(script)
		match=strings.Replace(match,"باقی مانده', value: ","",1)
		if match!=""{
			remaining=match
		}
		// fmt.Println("------------------------------------------------")
		// fmt.Print("باقی مانده")
		// fmt.Println(s.Text())
	})
	remain_float, err := strconv.ParseFloat(remaining, 64)
	if err != nil {
		log.Fatal(err)
	}
	pt:= ptime.New(time.Now())
	current_day_of_month:=pt.Day()
	var remaining_days int
	if current_day_of_month <= 10{
            remaining_days=(10-current_day_of_month)
	}else{
            remaining_days=10+pt.RMonthDay()
	}
    fmt.Printf("You have %v GB remaining data for %d days(%.2f per day).\n", remain_float, remaining_days, remain_float/float64(remaining_days))	
}
func logout(httpclient http.Client){
	net2_logout_url:=fmt.Sprintf(net2_url, "logout")
    r, err:=http.Get(net2_logout_url)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
        fmt.Println("Not connected!")
		fmt.Println("Not able to logout!")
	}else{
        fmt.Println("Successfully disconnected :)")
	}
}


func main() {
	var (
		dnsResolverIP        = "81.31.160.34:53" // Google DNS resolver.
		dnsResolverProto     = "udp"        // Protocol to use for the DNS resolver
		dnsResolverTimeoutMs = 5000         // Timeout (ms) for the DNS resolver (optional)
	)
	
	dialer := &net.Dialer{
	Resolver: &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
		}
		return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
		},
	},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
	return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
	httpClient := &http.Client{}

	var cr credentials=credentials{
		username: "alaei",
		password: "ecEgolSharif1@",
	}
	if(len(os.Args)<2){
		net2_login(*httpClient,cr)
		return
	}else if(len(os.Args)==2){
		if(os.Args[1]=="c"){
			check_bw(*httpClient,cr)
			return
		}else if(os.Args[1]=="d"){
			logout(*httpClient)
			return
		}
	}
	
	// check_bw(cr)
	// logout()
	fmt.Println(cr)
}
