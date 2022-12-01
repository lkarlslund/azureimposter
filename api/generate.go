package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/lkarlslund/oapi-codegen/pkg/codegen"
)

type generatorinfo struct {
	version   string
	cachefile string
	typename  string
	filename  string
}

func main() {
	versions := []generatorinfo{
		{
			version:   "v1.0",
			cachefile: "msgraphv10.json",
			typename:  "msgraph",
			filename:  "msgraph/types.go",
		},
		{
			version:   "beta",
			cachefile: "msgraphbeta.json",
			typename:  "msgraphbeta",
			filename:  "msgraphbeta/types.go",
		},
	}

	for _, version := range versions {
		err := generatetypes(version)
		if err != nil {
			fmt.Printf("Error generating types for version %v: %v", version, err)
		}
	}
}

func generatetypes(gi generatorinfo) error {
	// set up codegen
	opts := codegen.Configuration{
		PackageName: gi.typename,
		Generate: codegen.GenerateOptions{
			Models:       true,
			Client:       false,
			EmbeddedSpec: false,
		},
		// Compatibility: codegen.CompatibilityOptions{
		// 	OldAliasing: true,
		// },
	}

	var data []byte
	if f, err := os.Open(gi.cachefile); err == nil {
		log.Println("Loading cached schema from", gi.cachefile)
		data, err = ioutil.ReadAll(f)
		if err != nil {
			return err
		}
	} else {
		log.Println("Fetching schema from Graph Explorer API")
		resp, err := http.Get("https://graphexplorerapi.azurewebsites.net/openapi?operationIds=*&openApiVersion=3&graphVersion=" + gi.version + "&format=json")
		if err != nil {
			// handle error
			fmt.Printf("%s", err)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			return err
		}
		f, err := os.Create(gi.cachefile)
		if err == nil {
			f.Write(data)
			f.Close()
		}
	}

	loader := openapi3.NewLoader()

	data = []byte(prefixup(string(data)))

	log.Println("Loading swagger data")
	swagger, err := loader.LoadFromData(data)
	if err != nil {
		return err
	}

	// ctx := context.Background()
	// log.Println("Validating swagger data")
	// err = swagger.Validate(ctx, openapi3.DisableSchemaPatternValidation())
	// if err != nil {
	// 	return err
	// }

	log.Println("Generating code")
	code, err := codegen.Generate(swagger, opts)

	// code = postfixup(code)

	// Save it anyway
	err2 := ioutil.WriteFile(gi.filename, []byte(code), 0644)
	if err != nil {
		return err
	}
	log.Println("Done")
	return err2
}

var regexpFixup = regexp.MustCompile(`(?m)"(any|one)Of": \[\n\s+{\n\s+"\$ref": "([^"]+)"\n\s+}\n\s+\]`)

var regexpFixupSchema = regexp.MustCompile(`"\$ref": "#/components/schemas/([^"]+)"`)

func prefixup(input string) string {
	input = regexpFixupSchema.ReplaceAllString(input, `"\$ref": "#/components/schemas/schema.$1"`)
	input = strings.ReplaceAll(input, `"200": {`, `"Response200": {`)
	// replaced := map[string]struct{}{}
	// for _, match := range results {
	// 	replace := match[1]
	// 	if _, found := replaced[replace]; !found {
	// 		strings.ReplaceAll(input, `"#/components/schemas/`+replace+`"`, `"#/components/schemas/schema.`+replace+`"`)
	// 		replaced[replace] = struct{}{}
	// 	}
	// }
	// input = strings.ReplaceAll(input, `StringCollectionResponse`, `microsoft.graph.stringCollectionResponse`) // Not supported, ignore it
	// input = strings.ReplaceAll(input, `microsoft.graph.accessPackageAssignmentCollectionResponse`, `microsoft.graph.accessPackageAssignmentCollectionResponse2`) // Not supported, ignore it

	// input = strings.Replace(input, `"microsoft.graph.accessPackageAssignmentCollectionResponse": {`, `"microsoft.graph.accessPackageAssignmentCollectionResponse": {
	// "x-go-name": "Something",`, 1)

	// input = strings.ReplaceAll(input, `"format": "base64url"`, `"format": ""`) // Not supported, ignore it
	// input = strings.ReplaceAll(input, `"format": "int16"`, `"format": ""`)     // Not supported, ignore it

	input = strings.ReplaceAll(input, `"format": "decimal"`, `"format": "double"`)
	// result = strings.ReplaceAll(result, ` NaN`, ` "NaN C'mon Microsoft"`)
	// result = regexpFixup.ReplaceAllString(result, `"$$ref": "$2"`)
	return input
}

var regexpPostFixup = regexp.MustCompile(`([ \t]+[^\s]+)(.+json:"\S*_)`)

func postfixup(input string) string {
	return regexpPostFixup.ReplaceAllString(input, `${1}_${2}`)
}
