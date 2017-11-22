# go-loadtmpl
Load standard library HTML templates from an `http.Filesystem` while adding inheritance using a special comment that indicates the template's parent.

## Example

First define your base template, e.g _layout.html_:
```html
<html>
  <body>
    <h1>Page {{block "page_title" .}}{{end}}</h1>
    <div>
      {{block "page" .}}{{end}}
    </div>
  </body>
</html>
```

Then, write the template that extends it and add the specially crafted comment at the top:

```html
{{/* extends "layout.html" */}}
{{block "page_title" .}}About{{end}}
{{block "page" .}}I load templates{{end}}
```

In you Go file, probably somewhere next to your web handler

```Go
fs := http.Dir(".") //provide access to the template files through the http.FileSystem interface.
l := loadtmpl.New(fs)
t, _ := l.Load("about.html") //load the template, its base template will automatically loaded

t, _ = l.Load("about.html") //second calls are retrieved from cache
```
