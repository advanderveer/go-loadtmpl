package loadtmpl

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

//Loader will fetch templates by name from an http.FileSystem. It provides template inheritance by
//search the loaded template for a specially crafted comment. Loading of templates with the same
//name are cached and not loaded calls
type Loader struct {
	fs    http.FileSystem
	funcs template.FuncMap
	cache map[string]*template.Template
}

//New loader will be setup
func New(fs http.FileSystem, funcs template.FuncMap) *Loader {
	return &Loader{
		fs:    fs,
		funcs: funcs,
		cache: map[string]*template.Template{},
	}
}

//Load a template and cache it.
func (l *Loader) Load(name string) (t *template.Template, err error) {
	if t, ok := l.cache[name]; ok {
		return t, nil
	}

	l.cache[name], err = l.load(name)
	if err != nil {
		return nil, err
	}

	return l.cache[name], nil
}

func (l *Loader) load(name string) (t *template.Template, err error) {
	f, err := l.fs.Open(name)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	text := string(data)
	parent := MatchParentComment(data)
	if parent != "" {
		t, err = l.load(parent)
		if err != nil {
			return nil, fmt.Errorf("failed to load: %s", err)
		}

		text = fmt.Sprintf(`{{ define "%s"}}%s{{end}}`, name, text)
	} else {
		t = template.New(name)
		t = t.Funcs(l.funcs)
	}

	t, err = t.Parse(text)
	if err != nil {
		return nil, err
	}

	return t, nil
}

//MatchParentComment will search data for the comment that defines the parent it extends. If
//match is empty is means no parent is defined
func MatchParentComment(data []byte) string {
	groups := regExpExtendLine.FindSubmatch(data)
	if len(groups) != 2 {
		return ""
	}

	return string(groups[1])
}

var regExpExtendLine = regexp.MustCompile(`^\W*{{/\*\W*extends "([^"]*)"\W*\*/}}`)
