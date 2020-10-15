package nativehttpclient

import (
	"bytes"
	"html/template"
)

func unescape(s string) template.HTML {
	return template.HTML(s)
}

func StringTemplate(body string, data interface{}) string {
	tmpl, _ := template.New("StringTemplate").Funcs(template.FuncMap{"Unescape": unescape}).Parse(body)

	var result bytes.Buffer
	_ = tmpl.Execute(&result, data)

	return result.String()
}
