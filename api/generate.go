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

type generatorinfo struct {
	version  string
	typename string
	filename string
}

func main() {
	versions := []generatorinfo{
		{
			version:  "1.0",
			typename: "msgraph",
			filename: "msgraph/types.go",
		},
		{
			version:  "beta",
			typename: "msgraphbeta",
			filename: "msgraphbeta/types.go",
		},
	}

	for _, version := range versions {
		err := generatetypes(version.version, version.typename, version.filename)
		if err != nil {
			fmt.Printf("Error generating types for version %v: %v", version, err)
		}
	}
}

func generatetypes(version, typename, filename string) error {
	// set up codegen
	opts := codegen.Options{
		GenerateTypes: true,
	}

	resp, err := http.Get("https://graphexplorerapi.azurewebsites.net/openapi?operationIds=*&openApiVersion=3&graphVersion=" + version + "&format=json")
	if err != nil {
		// handle error
		fmt.Printf("%s", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return err
	}

	loader := openapi3.NewLoader()
	swagger, err := loader.LoadFromData([]byte(prefixup(string(data))))

	code, err := codegen.Generate(swagger, typename, opts)
	err = ioutil.WriteFile(filename, []byte(postfixup(code)), 0644)
	return err
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
