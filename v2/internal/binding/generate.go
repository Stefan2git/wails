package binding

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/internal/fs"

	"github.com/leaanthony/slicer"
)

//go:embed assets/package.json
var packageJSON []byte

func (b *Bindings) GenerateBackendJS(targetfile string, isDevBindings bool) error {

	store := b.db.store
	var output bytes.Buffer

	output.WriteString(`// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
`)

	if isDevBindings {
		json, err := b.ToJSON()
		if err != nil {
			return err
		}
		output.WriteString("window.wailsbindings = " + json + ";")
		output.WriteString("\n")
	}

	output.WriteString(`const go = {`)
	output.WriteString("\n")

	var sortedPackageNames slicer.StringSlicer
	for packageName := range store {
		sortedPackageNames.Add(packageName)
	}
	sortedPackageNames.Sort()
	sortedPackageNames.Each(func(packageName string) {
		packages := store[packageName]
		output.WriteString(fmt.Sprintf("  \"%s\": {", packageName))
		output.WriteString("\n")
		var sortedStructNames slicer.StringSlicer
		for structName := range packages {
			sortedStructNames.Add(structName)
		}
		sortedStructNames.Sort()

		sortedStructNames.Each(func(structName string) {
			structs := packages[structName]
			output.WriteString(fmt.Sprintf("    \"%s\": {", structName))
			output.WriteString("\n")

			var sortedMethodNames slicer.StringSlicer
			for methodName := range structs {
				sortedMethodNames.Add(methodName)
			}
			sortedMethodNames.Sort()

			sortedMethodNames.Each(func(methodName string) {
				methodDetails := structs[methodName]
				output.WriteString("      /**\n")
				output.WriteString("       * " + methodName + "\n")
				var args slicer.StringSlicer
				for count, input := range methodDetails.Inputs {
					arg := fmt.Sprintf("arg%d", count+1)
					args.Add(arg)
					output.WriteString(fmt.Sprintf("       * @param {%s} %s - Go Type: %s\n", goTypeToJSDocType(input.TypeName), arg, input.TypeName))
				}
				returnType := "Promise"
				returnTypeDetails := ""
				if methodDetails.OutputCount() > 0 {
					firstType := goTypeToJSDocType(methodDetails.Outputs[0].TypeName)
					returnType += "<" + firstType
					if methodDetails.OutputCount() == 2 {
						secondType := goTypeToJSDocType(methodDetails.Outputs[1].TypeName)
						returnType += "|" + secondType
					}
					returnType += ">"
					returnTypeDetails = " - Go Type: " + methodDetails.Outputs[0].TypeName
				} else {
					returnType = "Promise<void>"
				}
				output.WriteString("       * @returns {" + returnType + "} " + returnTypeDetails + "\n")
				output.WriteString("       */\n")
				argsString := args.Join(", ")
				output.WriteString(fmt.Sprintf("      \"%s\": (%s) => {", methodName, argsString))
				output.WriteString("\n")
				output.WriteString(fmt.Sprintf("        return window.go.%s.%s.%s(%s);", packageName, structName, methodName, argsString))
				output.WriteString("\n")
				output.WriteString(fmt.Sprintf("      },"))
				output.WriteString("\n")

			})

			output.WriteString("    },\n")
		})

		output.WriteString("  },\n\n")
	})

	output.WriteString(`};
export default go;`)
	output.WriteString("\n")

	dir := filepath.Dir(targetfile)
	packageJsonFile := filepath.Join(dir, "package.json")
	if !fs.FileExists(packageJsonFile) {
		err := os.WriteFile(packageJsonFile, packageJSON, 0755)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(targetfile, output.Bytes(), 0755)
}

func goTypeToJSDocType(input string) string {
	switch true {
	case input == "string":
		return "string"
	case input == "error":
		return "Error"
	case
		strings.HasPrefix(input, "int"),
		strings.HasPrefix(input, "uint"),
		strings.HasPrefix(input, "float"):
		return "number"
	case input == "bool":
		return "boolean"
	case input == "[]byte":
		return "string"
	case strings.HasPrefix(input, "[]"):
		arrayType := goTypeToJSDocType(input[2:])
		return "Array.<" + arrayType + ">"
	default:
		if strings.ContainsRune(input, '.') {
			return strings.Split(input, ".")[1]
		}
		return "any"
	}
}
