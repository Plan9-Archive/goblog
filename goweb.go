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
	
type BlogPost struct {
	Year	int
	Month	int
	Day		int
	Post	int
	Type	int

	Title	string
	Body	string
}


// Yuck yuck yuck
func ListPosts(year, month string) (res string) {
	res = month + "\n"
	res = res + "<ul>\n"
	f, err := os.Open(root + blogdir + year + "/" + month)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer f.Close()
	fi, err := f.Readdir(0)
	if err != nil {
		fmt.Print(err)
		return
	}
	// Step through the list of days
	for _, info := range fi {
		if info.IsDirectory() {
			res = res + "<li>" + info.Name + "\n"
			res = res + "<ul>\n"
			day, err := os.Open(root + blogdir + year + "/" + month + "/" + info.Name)
			if err != nil {
				fmt.Print(err)
				return
			}
			defer day.Close()
			posts, err := day.Readdir(0)
			if err != nil {
				fmt.Print(err)
				return
			}
			// Step through the posts under this day
			for _, post := range posts {
				p, err := os.Open(root + blogdir + year + "/" + month + "/" + info.Name + "/" + post.Name)
				if err != nil {
					fmt.Print(err)
					return
				}
				defer p.Close()
				read := bufio.NewReader(p)
				title, _, err := read.ReadLine()
				if err != nil {
					res = res + "<li>" + string(title) + "\n"
				} else {
					return
				}
			}
			res = res + "</ul>\n"
		}
	}
	res = res + "</ul>\n"
	return res
}

func GenYear(year string) (res string) {
	res = year + "\n"
	res = res + "<ul>\n"
	f, err := os.Open(root + blogdir + year)
	if err != nil {
		fmt.Print(err)
	}
	defer f.Close()
	fi, err := f.Readdir(0)
	if err != nil {
		fmt.Print(err)
	}
	for _, info := range fi {
		if info.IsDirectory() {
			res = res + "<li>" + ListPosts(year, info.Name)
		}
	}
	res = res + "</ul>\n"
	return res
}

func (b *BlogPost) GenArchivePage() {
	res := "<ul>\n"
	f, _ := os.Open(root + blogdir)
	defer f.Close()
	fi, _ := f.Readdir(0)
	for _, info := range fi {
		if info.IsDirectory() {
			res = res + "<li>" + GenYear(info.Name)
		}
	}
	res = res + "</ul>\n"
	b.Body = res
}

func NewBlogPost(path string) (b *BlogPost) {
	b = new(BlogPost)

	splitpath := strings.Split(path, "/")
	if path == "" {
		splitpath = []string{}
	}
	fmt.Printf("%#v splits to %#v\n", path, splitpath)

	switch len(splitpath) {
		case 4:
			b.Post, _ = strconv.Atoi(splitpath[3])
			fallthrough
		case 3:
			b.Day, _ = strconv.Atoi(splitpath[2])
			fallthrough
		case 2:
			b.Month, _ = strconv.Atoi(splitpath[1])
			fallthrough
		case 1:
			b.Year, _ = strconv.Atoi(splitpath[0])
			break;
	}

	b.Type = len(splitpath)

	return b
}

func BlogServer(w http.ResponseWriter, req *http.Request) {

	path := req.URL.Path[len(blogdir):]
	base := root + blogdir

	bp := NewBlogPost(path)

	bp.Title = "foobar"

	t := template.Must(template.ParseFile("/usr/john/goweb/blog.html"))

	switch bp.Type {
	case Post:
		tmp, _ := ioutil.ReadFile(base + path)
		bp.Body = string(tmp)
		break
	case Year:
		bp.Body = GenYear(strconv.Itoa(bp.Year))
		break
	case Day:
	case Month:
		bp.Body = ListPosts(strconv.Itoa(bp.Year), strconv.Itoa(bp.Month))
		break
	default:
		bp.GenArchivePage()
		break
	}
	t.Execute(w, bp)
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
