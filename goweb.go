package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/template"
)

const (
	TopLevel = iota
	Year
	Month
	Day
	Post
)

var (
	configpath = flag.String("config", "/lib/goweb.config", "Path to configuration file")
	nobody = flag.Bool("nobody", false, "Switch to user nobody")
	config     Config
)

type Config struct {
	Root       string
	Blogdir    string
	Shortname  string // disqus shortname
	Subdomains []Subdomain
}

type Subdomain struct {
	Domains []string
	Path   string
}

type Request struct {
	Year  int
	Month int
	Day   int
	Post  int
	Type  int
}

type Archive struct {
	Years     []*ArchiveYear
	Shortname string // the Disqus shortname for the site
}

type ArchiveYear struct {
	Year  string
	Posts bplist
}

type BlogPost struct {
	Path  string // e.g. "/b/2011/11/16/0"
	Title string // e.g. "My First Post"
	Body  string // the file converted to HTML
	Date  string

	Shortname string // the Disqus shortname of your site.
}

type bplist []*BlogPost

func (bp bplist) Len() int {
	return len(bp)
}

// This is all jenky, because I want to sort backwards. Evil, I know.
func (bp bplist) Less(i, j int) bool {
	si := strings.Split(bp[i].Date, "/")
	sj := strings.Split(bp[j].Date, "/")
	imon, _ := strconv.Atoi(si[0])
	jmon, _ := strconv.Atoi(sj[0])
	iday, _ := strconv.Atoi(si[1])
	jday, _ := strconv.Atoi(sj[1])

	if imon > jmon {
		return true
	}
	if imon == jmon && iday > jday {
		return true
	}
	return false
}

func (bp bplist) Swap(i, j int) {
	bp[i], bp[j] = bp[j], bp[i]
}

func init() {
	flag.Parse()
}

func GenYear(year string) (res bplist) {
	months, err := ioutil.ReadDir(config.Root + config.Blogdir + year)
	if err != nil {
		fmt.Print(err)
	}
	for _, month := range months {
		if month.IsDir() {
			days, err := ioutil.ReadDir(config.Root + config.Blogdir + year + "/" + month.Name())
			if err != nil {
				fmt.Print(err)
				return
			}
			// Step through the list of days
			for _, day := range days {
				if day.IsDir() {
					posts, err := ioutil.ReadDir(config.Root + config.Blogdir + year + "/" + month.Name() + "/" + day.Name())
					if err != nil {
						fmt.Print(err)
						return
					}
					// Step through the posts under this day
					for _, post := range posts {
						p, err := os.Open(config.Root + config.Blogdir + year + "/" + month.Name() + "/" + day.Name() + "/" + post.Name())
						if err != nil {
							fmt.Print(err)
							return
						}
						defer p.Close()
						read := bufio.NewReader(p)
						title, _, err := read.ReadLine()
						if err == nil {
							res = append(res, &BlogPost{config.Blogdir + year + "/" + month.Name() + "/" + day.Name() + "/" + post.Name(), string(title), "", month.Name() + "/" + day.Name(), ""})
						} else {
							fmt.Print(err)
						}
					}
				}
			}
		}
	}
	sort.Sort(res)
	return res
}

func GenArchivePage() (res Archive) {
	var y *ArchiveYear
	fi, err := ioutil.ReadDir(config.Root + config.Blogdir)
	if err != nil {
		return res
	}
	for _, info := range fi {
		if info.Mode().IsDir() {
			y = &ArchiveYear{info.Name(), GenYear(info.Name())}
			res.Years = append([]*ArchiveYear{y}, res.Years...)
		}
	}
	res.Shortname = config.Shortname
	return res
}

// I'm not actually sure most of this is required, but it may
// come in handy at some point.
func NewRequest(path string) (r *Request) {
	r = new(Request)

	splitpath := strings.Split(path, "/")
	if path == "" {
		splitpath = []string{}
	}

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
		break
	}

	r.Type = len(splitpath)

	return r
}

func BlogServer(w http.ResponseWriter, req *http.Request) {

	path := req.URL.Path[len(config.Blogdir):]
	base := config.Root + config.Blogdir

	r := NewRequest(path)

	bp := new(BlogPost)

	// Two choices: Either specify a full post,
	// or get sent to the archive page.
	switch r.Type {
	case Post:
		tmp, _ := ioutil.ReadFile(base + path)
		bp.Body = string(blackfriday.MarkdownCommon(tmp))
		bp.Date = strconv.Itoa(r.Year) + "/" + strconv.Itoa(r.Month) + "/" + strconv.Itoa(r.Day)
		bp.Shortname = config.Shortname
		p, err := os.Open(base + path)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer p.Close()
		read := bufio.NewReader(p)
		title, _, err := read.ReadLine()
		bp.Title = string(title)
		t := template.Must(template.ParseFiles(base + "/page.html"))
		t.Execute(w, bp)
	default:
		archive := GenArchivePage()
		t := template.Must(template.ParseFiles(base + "/archive.html"))
		t.Execute(w, archive)
	}
}

func ReadConfig(path string) (c Config) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
		panic("Couldn't read config file")
	}

	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Print(err)
		panic("Couldn't parse json")
	}
	return
}

func main() {
	u, err := user.Lookup("nobody")
        if err != nil {
                log.Fatalln(err)
        }
        uid, err := strconv.Atoi(u.Uid)
        if err != nil {
                log.Fatalln(err)
        }

	config = ReadConfig(*configpath)
	http.HandleFunc(config.Blogdir, BlogServer)
	http.Handle("/", http.FileServer(http.Dir(config.Root)))
	for _, s := range config.Subdomains {
		for _, sd := range s.Domains {
			http.Handle(sd, http.FileServer(http.Dir(config.Root+s.Path)))
		}
	}

	l, err := net.Listen("tcp", ":http")
	if err != nil {
                log.Fatalln(err)
        }

	if *nobody {
		err = syscall.Setuid(uid)
		if err != nil {
			log.Fatalln(err)
		}
	}
		
	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
