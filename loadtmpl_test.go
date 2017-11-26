package loadtmpl

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateLoading(t *testing.T) {
	for _, c := range []struct {
		Name      string
		Load      string
		Data      interface{}
		Files     map[string]string
		ExpOutput string
		NoCache   bool
	}{
		{
			Name:      "one extend",
			Load:      "/b.html",
			Data:      nil,
			ExpOutput: `a(b)`,
			Files: map[string]string{
				"a.html": `a({{block "a" .}}a{{end}})`,
				"b.html": `{{/* extends "a.html" */}}b({{block "a" .}}b{{end}})`,
			},
		},
		{
			Name:      "three level extend",
			Load:      "/c.html",
			Data:      nil,
			ExpOutput: `a(b(c))`,
			Files: map[string]string{
				"a.html": `a({{block "a" .}}a{{end}})`,
				"b.html": `{{/*extends "a.html"*/}}{{block "a" .}}b({{block "b" .}}b{{end}}){{end}}`,
				"c.html": `{{/*extends "b.html"*/}}{{block "b" .}}c{{end}}`,
			},
		},
		{
			Name:      "three level extend, no cache",
			Load:      "/c.html",
			Data:      nil,
			NoCache:   true,
			ExpOutput: `a(b(c))`,
			Files: map[string]string{
				"a.html": `a({{block "a" .}}a{{end}})`,
				"b.html": `{{/*extends "a.html"*/}}{{block "a" .}}b({{block "b" .}}b{{end}}){{end}}`,
				"c.html": `{{/*extends "b.html"*/}}{{block "b" .}}c{{end}}`,
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			dir, _ := ioutil.TempDir("", "loadtmpl_")
			defer os.RemoveAll(dir)

			for name, content := range c.Files {
				ioutil.WriteFile(filepath.Join(dir, name), []byte(content), 0777)
			}

			fs := http.Dir(dir)
			l := New(fs, nil)
			if c.NoCache {
				l.NoCache = true
			}

			tmpl, err := l.Load(c.Load)
			if err != nil {
				t.Fatal(err)
			}

			tmpl, err = l.Load(c.Load) //from cache
			if err != nil {
				t.Fatal(err)
			}

			buf := bytes.NewBuffer(nil)
			err = tmpl.Execute(buf, c.Data)
			if err != nil {
				t.Fatal(err)
			}

			if buf.String() != c.ExpOutput {
				t.Fatalf("expected output '%s', got: '%s'", c.ExpOutput, buf.String())
			}
		})
	}
}

func TestExtendMatch(t *testing.T) {
	for _, c := range []struct {
		Data      []byte
		ExpParent string
	}{
		{Data: []byte(`{{/* extends "layout.html" */}}`), ExpParent: "layout.html"},
		{Data: []byte(`{{/*extends "layout.html" */}}`), ExpParent: "layout.html"},
		{Data: []byte(`{{/*extends "layout.html"*/}}`), ExpParent: "layout.html"},
		{Data: []byte("\n" + `{{/*extends "layout.html"*/}}`), ExpParent: "layout.html"},
		{Data: []byte(`{{/*extends "layout.html"*/}}{{/*extends "layout.html"*/}}`), ExpParent: "layout.html"},
		{Data: []byte(`{{/* extends "" */}}`), ExpParent: ""},
		{Data: []byte(`{{/* extends"layout.html" */}}`), ExpParent: ""},
		{Data: []byte(`{{/* extends "layo"ut.html" */}}`), ExpParent: ""},
		{Data: []byte(`{{/* ` + "\n" + ` extends "layout.html"` + "\n" + ` */}}`), ExpParent: "layout.html"},
		{Data: []byte(`{{/* extends` + "\n" + `"layout.html" */}}`), ExpParent: ""},
		{Data: []byte(`n{{/* extends "layout.html" */}}`), ExpParent: ""},
		{Data: []byte(` {{/* extends "layout.html" */}}`), ExpParent: "layout.html"},
	} {
		t.Run(string(c.Data), func(t *testing.T) {
			parent := MatchParentComment(c.Data)
			if parent != c.ExpParent {
				t.Fatalf("expected parent '%s', got: '%s'", c.ExpParent, parent)
			}
		})
	}

}
