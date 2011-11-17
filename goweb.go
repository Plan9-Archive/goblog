package main

import (
	"http"
	"log"
	"strings"
	"strconv"
	"fmt"
	"os"
	"io/ioutil"
	"template"
	"bufio"
)

const (
	TopLevel = iota
	Year
	Month
	Day
	Post

	root = "/usr/john/www/"
	blogdir = "/b/"
)
	
type Request struct {
	Year	int
	Month	int
	Day		int
	Post	int
	Type	int
}

type Archive struct {
	Years	[]*ArchiveYear
}

type ArchiveYear struct {
	Year	string
	Posts	[]*BlogPost
}

type BlogPost struct {
	Path	string // e.g. "/b/2011/11/16/0"
	Title	string // e.g. "My First Post"
	Body	string // the file converted to HTML
}

func GenYear(year string) (res []*BlogPost) {
	f, err := os.Open(root + blogdir + year)
	if err != nil {
		fmt.Print(err)
	}
	defer f.Close()
	months, err := f.Readdir(0)
	if err != nil {
		fmt.Print(err)
	}
	for _, month := range months {
		if month.IsDirectory() {
			fmt.Printf("month = %v\n", month.Name)
			g, err := os.Open(root + blogdir + year + "/" + month.Name)
			if err != nil {
				fmt.Print(err)
				return
			}
			defer g.Close()
			days, err := g.Readdir(0)
			if err != nil {
				fmt.Print(err)
				return
			}
			// Step through the list of days
			for _, day := range days {
				if day.IsDirectory() {
					fmt.Printf("day = %v\n", day.Name)
					h, err := os.Open(root + blogdir + year + "/" + month.Name + "/" + day.Name)
					if err != nil {
						fmt.Print(err)
						return
					}
					defer h.Close()
					posts, err := h.Readdir(0)
					if err != nil {
						fmt.Print(err)
						return
					}
					// Step through the posts under this day
					for _, post := range posts {
						fmt.Printf("post = %v\n", post.Name)
						p, err := os.Open(root + blogdir + year + "/" + month.Name + "/" + day.Name + "/" + post.Name)
						if err != nil {
							fmt.Print(err)
							return
						}
						defer p.Close()
						read := bufio.NewReader(p)
						title, _, err := read.ReadLine()
						fmt.Printf("read title = %v, err = %v\n", title, err)
						if err == nil {
							fmt.Printf("appending a post with title\n")
							res = append(res, &BlogPost{blogdir + year + "/" + month.Name + "/" + day.Name + "/" + post.Name, string(title), ""})
							fmt.Printf("res is now %v\n", res)
						} else {
							fmt.Print(err)
						}
					}
				}
			}
		}
	}
	return res
}

func GenArchivePage() (res Archive) {
	var y *ArchiveYear
	f, _ := os.Open(root + blogdir)
	defer f.Close()
	fi, _ := f.Readdir(0)
	for _, info := range fi {
		if info.IsDirectory() {
			y = &ArchiveYear{info.Name, GenYear(info.Name)}
			res.Years = append(res.Years, y)
		}
	}
	return res
}

func NewRequest(path string) (r *Request) {
	r = new(Request)

	splitpath := strings.Split(path, "/")
	if path == "" {
		splitpath = []string{}
	}
	fmt.Printf("%#v splits to %#v\n", path, splitpath)

	switch len(splitpath) {
		case 4:
			r.Post, _ = strconv.Atoi(splitpath[3])
			fallthrough
		case 3:
			r.Day, _ = strconv.Atoi(splitpath[2])
			fallthrough
		case 2:
			r.Month, _ = strconv.Atoi(splitpath[1])
			fallthrough
		case 1:
			r.Year, _ = strconv.Atoi(splitpath[0])
			break;
	}

	r.Type = len(splitpath)

	return r
}

func BlogServer(w http.ResponseWriter, req *http.Request) {

	path := req.URL.Path[len(blogdir):]
	base := root + blogdir

	r := NewRequest(path)

	bp := new(BlogPost)

	switch r.Type {
	case Post:
		tmp, _ := ioutil.ReadFile(base + path)
		bp.Body = string(tmp)
		t := template.Must(template.ParseFile("/usr/john/goweb/page.html"))
		t.Execute(w, bp)
		break
	default:
		archive := GenArchivePage()
		fmt.Printf("posts = %v\n", (archive.Years[0].Posts))
		t := template.Must(template.ParseFile("/usr/john/goweb/archive.html"))
		t.Execute(w, archive)
		break
	}
}

func main() {
	os.Chdir(root)
	http.HandleFunc(blogdir, BlogServer)
	http.Handle("/", http.FileServer(http.Dir(root)))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.String())
	}
}
