package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/css"
	"golang.org/x/gahttp"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func Banner() {
	banner := `
    __    __            __   ___    __                    __             
   / /_  / /_____ ___  / /  /   |  / /_  _________  _____/ /_  ___  _____
  / __ \/ __/ __ \__ \/ /  / /| | / __ \/ ___/ __ \/ ___/ __ \/ _ \/ ___/
 / / / / /_/ / / / / / /  / ___ |/ /_/ (__  ) /_/ / /  / /_/ /  __/ /
/_/ /_/\__/_/ /_/ /_/_/  /_/  |_/_.___/____/\____/_/  /_.___/\___/_/
                                    Version:1.0 Author:zha0gongz1@影

Usage: html-Absorber <mode> [<args>]
Modes:
 tags <tag names>        Extract text contained in tags
 attribs <attrib names>  Extract attribute values
 comments                Extract comments
Option:
 -output                 Save the result to file

Examples:
 cat urls.txt | html-Absorber tags title [-output]
 find . -type f -name "*.html" | html-Absorber attribs src href [-output]
 cat urls.txt | html-Absorber comments [-output]
`
	print(banner)
}

func extractSelector(r io.Reader, selector string) ([]string, error) {

	out := []string{}

	sel, err := css.Compile(selector)
	if err != nil {
		return out, err
	}

	node, err := html.Parse(r)
	if err != nil {
		return out, err
	}

	// it's kind of tricky to actually know what to output
	// if the resulting tags contain more than just a text node
	for _, ele := range sel.Select(node) {
		if ele.FirstChild == nil {
			continue
		}
		out = append(out, ele.FirstChild.Data)
	}

	return out, nil
}

func extractComments(r io.Reader) []string {

	z := html.NewTokenizer(r)

	out := []string{}
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.CommentToken {
			d := strings.Replace(t.Data, "\n", " ", -1)
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			out = append(out, d)
		}

	}
	return out
}

func extractAttribs(r io.Reader, attribs []string) []string {
	z := html.NewTokenizer(r)

	out := []string{}

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		for _, a := range t.Attr {

			if a.Val == "" {
				continue
			}

			for _, attrib := range attribs {
				if attrib == a.Key {
					out = append(out, a.Val)
				}
			}
		}
	}
	return out
}

func extractTags(r io.Reader, tags []string) []string {
	z := html.NewTokenizer(r)

	out := []string{}

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break
		}

		t := z.Token()

		if t.Type == html.StartTagToken {

			for _, tag := range tags {
				if t.Data == tag {
					if z.Next() == html.TextToken {
						text := strings.TrimSpace(z.Token().Data)
						if text == "" {
							continue
						}
						out = append(out, text)
					}
				}
			}
		}
	}
	return out
}

type target struct {
	location string
	r        io.ReadCloser
}

func main() {

	flag.Parse()

	//check mode is valid
	mode := flag.Arg(0)
	if mode == "" {
		Banner()
		return
	}

	args := flag.Args()[1:]
	var temp string
	//var str1 = [1]string{"output"}
	//if flag.NArg()>2{
	//	outArr = flag.Args()[flag.NArg():]
	//}
	for _, i := range flag.Args()[flag.NArg()-1:] {  //遍历数组中所有元素追加成string
		temp += i
	}
	//outArr := flag.Args()[2:]
	//fmt.Println(flag.NArg())
	//fmt.Println(flag.Args()[2:])
	targets := make(chan *target)	//无缓冲channel
	var wg sync.WaitGroup	//协程同步
	wg.Add(1)

	go func() {

		for t := range targets {
			vals := []string{}
			switch mode {
			case "tags":
				vals = extractTags(t.r, args)
			case "attribs":
				vals = extractAttribs(t.r, args)
			case "comments":
				vals = extractComments(t.r)
			case "query":
				var err error
				vals, err = extractSelector(t.r, flag.Arg(1))
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to parse CSS selector: %s\n", err)
					break
				}

			default:
				fmt.Fprintf(os.Stderr, "unsupported mode '%s'\n", mode)
				break
			}
			//是否输出文件
			if temp != "-output"{
				for _, v := range vals {
					fmt.Println(v)
				}
			}else{
				fi, _ := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
				defer fi.Close()
				//结构体格式化输入指定内容
				_,err := fi.WriteString("[+]Success to fetch URL:"+t.location+"\n")
				if err != nil {
					return
				}

				for _, v := range vals {
					_,err = fi.WriteString(v+"\n")
					if err != nil {
						return
					}
				}
			}

			//for _, v := range vals {
			//			fmt.Println(v)
			//}
			// 完成后，需要关闭
			t.r.Close()
		}

		wg.Done()
	}()

	p := gahttp.NewPipeline()
	p.SetClient(gahttp.NewClient(gahttp.SkipVerify))
	p.SetConcurrency(20)

	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		// location 可以是文件名也可以是URL
		location := strings.TrimSpace(s.Text())

		// 如果是URL，需要gahttp请求获取
		nl := strings.ToLower(location)
		if strings.HasPrefix(nl, "http:") || strings.HasPrefix(nl, "https:") {
			p.Get(location, func(req *http.Request, resp *http.Response, err error) {
				if err != nil {
					time.Sleep(1*time.Second)
					fmt.Fprintf(os.Stderr,"\u001B[1;31;40m[-]failed to fetch URL: %s\u001B[0m\n", err)
					//fmt.Fprintf(os.Stderr,"[-]Failed to fetch URL: %s\n", err)

				}

				if resp != nil && resp.Body != nil {
					time.Sleep(5*time.Second)
					fmt.Fprintf(os.Stderr,"\u001B[0;40;32m[+]Success to fetch URL: %s\u001B[0m\n", nl)
					//fmt.Fprintf(os.Stderr,"[+]Success to fetch URL: %s\n", nl)

					targets <- &target{req.URL.String(), resp.Body}
				}
			})
			continue
		}

		//如果是本地html文件
		f, err := os.Open(location)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[-]Failed to open file: %s\n", err)
			continue
		}
		targets <- &target{location, f}

	}

	p.Done()
	p.Wait()

	close(targets)
	wg.Wait()
}
