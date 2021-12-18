package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	// set up codegen
	opts := codegen.Options{
		GenerateTypes: true,
	}

	// v1.0
	resp, err := http.Get("https://graphexplorerapi.azurewebsites.net/openapi?operationIds=*&openApiVersion=3&graphVersion=beta&format=json")
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}

	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(prefixup(string(data))))

	code, err := codegen.Generate(swagger, "msgraph", opts)
	err = ioutil.WriteFile("msgraph/types.go", []byte(postfixup(code)), 0644)
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}

	// beta
	resp, err = http.Get("https://graphexplorerapi.azurewebsites.net/openapi?operationIds=*&openApiVersion=3&graphVersion=beta&format=json")
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}

	loader = openapi3.NewLoader()
	swagger, err = loader.LoadFromData([]byte(prefixup(string(data))))

	code, err = codegen.Generate(swagger, "msgraphbeta", opts)
	err = ioutil.WriteFile("msgraphbeta/types.go", []byte(postfixup(code)), 0644)
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}
}

var regexpFixup = regexp.MustCompile(`(?m)"(any|one)Of": \[\n\s+{\n\s+"\$ref": "([^"]+)"\n\s+}\n\s+\]`)

func prefixup(input string) string {
	result := strings.ReplaceAll(input, `"format": "decimal"`, `"format": "double"`)
	result = strings.ReplaceAll(result, ` NaN`, ` "NaN C'mon Microsoft"`)
	result = regexpFixup.ReplaceAllString(result, `"$$ref": "$2"`)
	return result
}

var regexpPostFixup = regexp.MustCompile(`([ \t]+[^\s]+)(.+json:"\S*_)`)

func postfixup(input string) string {
	return regexpPostFixup.ReplaceAllString(input, `${1}_${2}`)
}
