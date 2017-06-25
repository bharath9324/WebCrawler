package main

import (
	"fmt"        
	"net/http"    
	"golang.org/x/net/html"
	"strings"    
	"crypto/tls"
	"time"
	"net/url"
	"sync"
)


var domainName string
var mutex sync.Mutex

func main() {
	
    linkQueue := make(chan string)         
    done := make(chan bool)
//---------------Write the domain in which you want to crawl --------------------
    var url string = "https://monzo.com/"
//-------------------------------------------------------------------------------
    domainName = url

	go crawl(linkQueue, done)
    linkQueue <- url 
	<-done
}


func crawl(linkQueue chan string, done chan bool) {
	visited := make(map[string]bool)
	for {
		select {
		case url := <-linkQueue:
			if _, yes := visited[url]; yes {
				continue
			} else {
				visited[url] = true
				go getLinks(url, linkQueue)
			}
		case <-time.After(4 * time.Second):
			fmt.Printf("Explored %d pages\n", len(visited))
			done <- true
		}
	}
}


func getLinks(url string , linkQueue chan string){
	
	transport := &http.Transport{
	    TLSClientConfig: &tls.Config{
	      InsecureSkipVerify: true,
	    },
  	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(url)                   
	if err != nil {
		return
	}
                            
 	page := html.NewTokenizer(resp.Body)                                         
    

 	mutex.Lock()
 	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("At:")
	fmt.Println(url)

	fmt.Println("-----------------------------------------------------------------")
	
    for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			break
		}
		token := page.Token()
	    if tokenType == html.StartTagToken && token.DataAtom.String() == "img" {
			for _, attr := range token.Attr {
				var cleanedLink string = cleanLink(attr.Val, url)
				if attr.Key == "src"{
					fmt.Print("Image: ")
					fmt.Println(cleanedLink)
				}
			}
		}

		if tokenType == html.StartTagToken && token.DataAtom.String() == "link" {
			for _, attr := range token.Attr {
				var cleanedLink string = cleanLink(attr.Val, url)
				if attr.Key == "href"{
					fmt.Print("Link: ")
					fmt.Println(cleanedLink)
				}
			}
		}

		if tokenType == html.StartTagToken && token.DataAtom.String() == "script" {
			for _, attr := range token.Attr {
				var cleanedLink string = cleanLink(attr.Val, url)
				if attr.Key == "src"{
					fmt.Print("Script: ")
					fmt.Println(cleanedLink)
				}
			}
		}

	    if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				var cleanedLink string = cleanLink(attr.Val, url) 
				fmt.Print("URL: ")
				fmt.Println(cleanedLink) 
				if attr.Key == "href"  && checkLink(cleanedLink) {
					linkQueue <- cleanedLink
				}
			}
		}
	}
	fmt.Println("-----------------------------------------------------------------")
	mutex.Unlock()

    	 
  	

	resp.Body.Close()
}

func cleanLink(link string, base string) string {
	if strings.Contains(link, "#") {
		var i int 
		for i=0; i<len(link); i++{
			if link[i] == '#' {
				break
			}
		}
		link = link[:i]
	}

	uri, err := url.Parse(link)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()

}

func checkLink(oneLink string) bool {
	if strings.HasPrefix(oneLink, domainName)  {
		return true
	}
	return false

}

