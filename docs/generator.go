// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	inputFilename  string = "openapi.yaml"
	targetFileName string = "openapi.go"
)

var conv = map[string]interface{}{"conv": fmtByteSlice}

func fmtByteSlice(s []byte) string {
	builder := strings.Builder{}
	for _, v := range s {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}
	return builder.String()
}

func main() {
	b, err := ioutil.ReadFile(inputFilename)
	if err != nil {
		log.Fatalf("Error reading %s: %s", inputFilename, err)
	}

	tmpl, err := template.New("openapi.yaml.tmpl").Funcs(conv).ParseFiles("openapi.yaml.tmpl")
	if err != nil {
		log.Fatal("Error parsing template: ", err)
	}

	builder := &bytes.Buffer{}
	data := make(map[string]interface{})
	data["date"] = time.Now().UTC().Format(time.RFC3339)
	data["content"] = b
	if err = tmpl.Execute(builder, data); err != nil {
		log.Fatal("Error executing template: ", err)
	}

	f, err := os.Create(targetFileName)
	if err != nil {
		log.Fatal("Error creating target file:", err)
	}
	defer f.Close()

	out, err := format.Source(builder.Bytes())
	if err != nil {
		log.Fatal("Error formatting generated code", err)
	}

	if err = ioutil.WriteFile(targetFileName, out, os.ModePerm); err != nil {
		log.Fatal("Error writing output file", err)
	}
}
