package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/blackflagsoftware/go-swagify/config"
	in "github.com/blackflagsoftware/go-swagify/internal"
	ope "github.com/blackflagsoftware/go-swagify/internal/openapi"
	opr "github.com/blackflagsoftware/go-swagify/internal/operation"
	par "github.com/blackflagsoftware/go-swagify/internal/parameter"
	perr "github.com/blackflagsoftware/go-swagify/internal/parseerror"
	pat "github.com/blackflagsoftware/go-swagify/internal/path"
	req "github.com/blackflagsoftware/go-swagify/internal/requestBody"
	res "github.com/blackflagsoftware/go-swagify/internal/response"
	sch "github.com/blackflagsoftware/go-swagify/internal/schema"
	ser "github.com/blackflagsoftware/go-swagify/internal/server"
	"gopkg.in/yaml.v2"
)

func main() {
	var inputPath string
	var outputPath string
	flag.StringVar(&inputPath, "inputPath", "", "Working directory, omit to run in current directory")
	flag.StringVar(&outputPath, "outputPath", "", "outputPath file name with path, omit to save in current path with swagger.yaml|json")
	flag.StringVar(&config.OutputFormat, "outputFormat", "yaml", "yaml | json: outputPath file type, default of yaml if omitted")
	flag.StringVar(&config.AppOutputFormat, "appOutputFormat", "json", "your app's output format, default of json if omitted")
	flag.StringVar(&config.AltFieldFormat, "altFieldFormat", "snakeCase", "snakeCase | kebabCase | camelCase | pascalCase | lowerCase | upperCase: used as alternate field formatting")
	flag.Parse()
	if inputPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err.Error())
		}
		inputPath = wd
	}
	if outputPath == "" {
		outputPath = path.Join(inputPath, "swagger."+config.OutputFormat)
	}
	// parse all comments and put them in a map by type
	comments := in.ParseDirForComments(inputPath)
	swagifyComments := in.ParseSwagifyComment(comments)
	// temp output
	for k, v := range swagifyComments.Types {
		fmt.Println(k, " => ", v)
	}
	// parse for all known structs
	myStructs := in.ParseDirForStructs(inputPath, swagifyComments.Types["struct"])

	// create a new openApi struct to add everything to
	open := ope.BuildOpenApi(swagifyComments.Types["openapi"])

	// process servers
	servers := ser.BuildServers(swagifyComments.Types["server"])
	// add to openapi any servers that belong to the top level
	open.Servers = servers["openapi"]

	// build schemas & parameters
	schemas := sch.BuildSchemaStruct(myStructs)
	sch.BuildSchema(swagifyComments.Types["schema"], schemas)
	parameters := par.BuildParameters(swagifyComments.Types["parameter"])

	// build the request body section
	requestBodies := req.BuildRequestBody(swagifyComments.Types["requestBody"])

	// build the response section
	responses := res.BuildResponse(swagifyComments.Types["response"])

	// build the components section
	open.Components = ope.Component{Parameters: parameters, Schemas: schemas, Responses: responses, RequestBodies: requestBodies}

	// operations
	operations := opr.BuildOperations(swagifyComments.Types["operation"])

	// paths
	open.Paths = pat.BuildPaths(swagifyComments.Types["path"], operations)

	var outByte []byte
	var err error
	if config.OutputFormat == "yaml" {
		outByte, err = yaml.Marshal(open)
		if err != nil {
			fmt.Println("Marshal", err)
			return
		}
	}
	if config.OutputFormat == "json" {
		outByte, err = json.MarshalIndent(open, "", "  ")
		if err != nil {
			fmt.Println("Marshal", err)
			return
		}
	}
	errWrite := os.WriteFile(outputPath, outByte, 0644)
	if errWrite != nil {
		fmt.Println("write", err)
	}
	perr.PrintErrors()
}
