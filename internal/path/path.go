package path

import (
	"fmt"
	"regexp"
	"strings"

	in "github.com/blackflagsoftware/go-swagify/internal"
	// srv "github.com/blackflagsoftware/go-swagify/internal/server"
	opr "github.com/blackflagsoftware/go-swagify/internal/operation"
	par "github.com/blackflagsoftware/go-swagify/internal/parameter"
	perr "github.com/blackflagsoftware/go-swagify/internal/parseerror"
)

type (
	Path struct {
		Summary     string             `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description string             `json:"description,omitempty" yaml:"description,omitempty"`
		Parameters  []par.ParameterRef `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		Get         *opr.Operation     `json:"get,omitempty" yaml:"get,omitempty"`
		Put         *opr.Operation     `json:"put,omitempty" yaml:"put,omitempty"`
		Post        *opr.Operation     `json:"post,omitempty" yaml:"post,omitempty"`
		Delete      *opr.Operation     `json:"delete,omitempty" yaml:"delete,omitempty"`
		Options     *opr.Operation     `json:"options,omitempty" yaml:"options,omitempty"`
		Head        *opr.Operation     `json:"head,omitempty" yaml:"head,omitempty"`
		Patch       *opr.Operation     `json:"patch,omitempty" yaml:"patch,omitempty"`
		Trace       *opr.Operation     `json:"trace,omitempty" yaml:"trace,omitempty"`
	}
)

/* go-swagify
@@path: <path url>
@@summary: (optional)
@@description: (optional)
@@parameters.ref: (optional) semicolon(;) list of ref parameter names
*/
func BuildPaths(comments in.SwagifyComment, operationBuilds map[string]opr.OperationBuild) map[string]Path {
	paths := make(map[string]Path)
	path := &Path{}
	for name, lines := range comments.Comments {
		err := parsePathLines(lines, path, operationBuilds[name])
		if err != nil {
			// will never be not nil
			continue
		}
		paths[name] = *path
	}
	return paths
}

func parsePathLines(lines []string, path *Path, operationBuilds opr.OperationBuild) error {
	// go through each line and do logic on
	reg := regexp.MustCompile("(?P<name>[a-zA-Z/.]+): *?(?P<value>.+)")
	for _, line := range lines {
		matches := reg.FindStringSubmatch(line)
		nameIdx := reg.SubexpIndex("name")
		valueIdx := reg.SubexpIndex("value")
		if len(matches) < 2 {
			perr.AddError(fmt.Sprintf("[Warning] @@path: bad format of line: %s", line))
			continue
		}
		value := strings.TrimSpace(matches[valueIdx])
		switch matches[nameIdx] {
		case "summary":
			path.Summary = value
		case "description":
			path.Description = value
		case "parameters.ref":
			split := strings.Split(value, ";")
			parameters := []par.ParameterRef{}
			for i := range split {
				parameters = append(parameters, par.ParameterRef{Ref: fmt.Sprintf("#/components/parameters/%s", split[i])})
			}
			path.Parameters = parameters
		default:
			perr.AddError(fmt.Sprintf("[Warning] @@path: invalid name option: %s", line))
		}
	}
	linkOperations(path, operationBuilds)
	return nil
}

func linkOperations(path *Path, operationBuilds opr.OperationBuild) {
	for k, v := range operationBuilds.Operations {
		switch k {
		case "get":
			path.Get = &v
		case "put":
			path.Put = &v
		case "post":
			path.Post = &v
		case "delete":
			path.Delete = &v
		case "options":
			path.Options = &v
		case "head":
			path.Head = &v
		case "patch":
			path.Patch = &v
		case "trace":
			path.Trace = &v
		default:
			perr.AddError(fmt.Sprintf("[Warning] @@path: invalid method: %s", k))
		}
	}
}
