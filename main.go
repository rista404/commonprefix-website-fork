package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

const (
	servePort = "8100"
	tmplDir   = "./templates"
	buildDir  = "./public"

	layoutTmplName   = "layout.html"
	indexTmplName    = "index.html"
	teamTmplName     = "team.html"
	projectTmplName  = "project.html"
	researchTmplName = "research.html"
)

const description string = "Common Prefix is a small team of scientists and software engineers offering blockchain science consulting services."

var layoutPath = filepath.Join(tmplDir, layoutTmplName)
var homeTmpl = template.Must(template.ParseFiles(layoutPath, filepath.Join(tmplDir, indexTmplName)))
var teamTmpl = template.Must(template.ParseFiles(layoutPath, filepath.Join(tmplDir, teamTmplName)))
var projectTmpl = template.Must(template.ParseFiles(layoutPath, filepath.Join(tmplDir, projectTmplName)))
var researchTmpl = template.Must(template.ParseFiles(layoutPath, filepath.Join(tmplDir, researchTmplName)))

// Data structures

type TeamMember struct {
	Handle         string
	Name           string
	Specialization string
	Desc           template.HTML
	Image          string
}

type Finding struct {
	Url  string
	Name string
}

func (f *Finding) Ext() string {
	bits := strings.Split(f.Url, ".")
	return bits[len(bits)-1]
}

type Project struct {
	Handle   string
	Name     string
	Body     template.HTML
	Image    template.HTML
	Findings []Finding
	Team     []TeamMember
}

type ResearchPaper struct {
	Handle     string
	Name       string
	Conference string
	Authors    string
	Url        string
	Tags       []Tag
}

type Research struct {
	ResearchPapers []ResearchPaper
	TagToColor     map[Tag]string
}

type Page struct {
	SmallContainer bool
	Title          string
	Description    string
	Members        []TeamMember
	Projects       []Project
	Research       Research
}

type ProjectPage struct {
	SmallContainer bool
	Title          string
	Description    string
	Project        Project
	NextProject    Project
}

type ResearchPage struct {
	Title      string
	TagToColor map[Tag]string
}

var team = []TeamMember{}

func htmlToFormattedString(s template.HTML) string {
	bmp := bluemonday.StripTagsPolicy()
	replacer := strings.NewReplacer("\n", " ", "\t", "")
	return replacer.Replace(strings.TrimSpace(bmp.Sanitize(string(s))))
}

func build() {
	//
	// Build index page
	//
	index := filepath.Join(buildDir, indexTmplName)
	// Remove the old version
	os.Remove(index)
	// Create new file
	f, err := os.Create(index)
	if err != nil {
		log.Fatalf("can't create %s", indexTmplName)
	}
	homeTmpl.ExecuteTemplate(f, "base", Page{SmallContainer: true, Title: "", Description: description, Members: team, Projects: Projects})
	f.Close()
	fmt.Printf("🏠  %s sucessfully generated.\n", indexTmplName)

	//
	// Build team page
	//
	teamPage := filepath.Join(buildDir, teamTmplName)
	// Remove the old version
	os.Remove(teamPage)
	// Create new file
	f, err = os.Create(teamPage)
	if err != nil {
		log.Fatalf("can't create %s", teamTmplName)
	}
	teamTmpl.ExecuteTemplate(f, "base", Page{Title: " — Team", Members: team, Description: description, Projects: Projects})
	f.Close()
	fmt.Printf("👫  %s sucessfully generated.\n", teamTmplName)

	//
	// Build projects directory
	//
	psf := filepath.Join(buildDir, "projects")
	_ = os.Mkdir(psf, os.ModePerm)

	// Build project pages
	for idx, p := range Projects {
		nextP := Projects[(idx+1)%len(Projects)]
		pf := filepath.Join(buildDir, "projects", p.Handle+".html")
		// Remove the old version
		os.Remove(pf)
		// Create new file
		f, err = os.Create(pf)
		if err != nil {
			log.Fatal("can't create projects/" + p.Handle + ".html")
		}

		projectTmpl.ExecuteTemplate(f, "base", ProjectPage{SmallContainer: true, Title: " — " + p.Name, Description: htmlToFormattedString(p.Body), Project: p, NextProject: nextP})
		f.Close()
		fmt.Printf("📖  projects/%s.html sucessfully generated.\n", p.Handle)
	}

	//
	// Build research page
	//
	researchPage := filepath.Join(buildDir, researchTmplName)
	// Remove the old version
	os.Remove(researchPage)
	// Create new file
	f, err = os.Create(researchPage)
	if err != nil {
		log.Fatalf("can't create %s", researchTmplName)
	}
	// for _, r := range Research {
	// 	Authors = ``
	// 	for _, a := range r.Authors {
	// 		if Slice.Contains(team, a) {
	// 			Authors += `<a href="/team#` + a.Handle + `>` + a.Name + `</a>, `
	// 		}
	// 	}
	// }
	researchTmpl.ExecuteTemplate(f, "base", Page{Title: " — Research", Description: description, Research: Research{ResearchPapers: ResearchPapers, TagToColor: TagToColor}})
	f.Close()
	fmt.Printf("👫  %s sucessfully generated.\n", researchTmplName)
}

func main() {
	// build a list of all members
	for _, m := range Members {
		team = append(team, m)
	}
	// sort team members
	sort.Slice(team, func(i, j int) bool {
		lastname := func(n string) string {
			n = strings.TrimPrefix(n, "Prof. ")
			n = strings.TrimPrefix(n, "Dr. ")
			n = strings.Split(n, " ")[1]
			return n
		}
		n1 := lastname(team[i].Name)
		n2 := lastname(team[j].Name)
		return n1 < n2
	})

	serve := flag.Bool("serve", false, "serve mode")
	port := flag.String("p", servePort, "port to serve on")
	flag.Parse()

	build()

	if *serve == true {
		fs := http.FileServer(http.Dir(buildDir))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Clean urls
			if strings.HasPrefix(r.URL.Path, "/projects") || strings.HasPrefix(r.URL.Path, "/team") {
				if !strings.HasSuffix(r.URL.Path, ".html") {
					r.URL.Path = r.URL.Path + ".html"
				}
			}

			fs.ServeHTTP(w, r)
		})
		fmt.Printf("🧞‍♂️  Serving on http://localhost:%s\n", *port)
		log.Fatal(http.ListenAndServe(":"+*port, nil))
	}
}
