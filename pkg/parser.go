package pkg

import (
	"context"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	getter "github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type Layer struct {
	Type    string `hcl:"type,label"`
	Name    string `hcl:"name,label"`
	Content string `hcl:"content,optional"`
	X       int    `hcl:"x,optional"`
	Y       int    `hcl:"y,optional"`
	Width   int    `hcl:"width,optional"`
	Height  int    `hcl:"height,optional"`
	Size    int    `hcl:"size,optional"`
	Font    string `hcl:"font,optional"`
	Color   *Color `hcl:"color,block"`
}

type Color struct {
	Type  string    `hcl:"type,label"`
	Value string    `hcl:"value,optional"`
	Start *Position `hcl:"start,block"`
	End   *Position `hcl:"end,block"`
	Stops []Stop    `hcl:"stop,block"`
}

type Position struct {
	X int `hcl:"x"`
	Y int `hcl:"y"`
}

type Stop struct {
	Position float32 `hcl:"position"`
	Value    string  `hcl:"value"`
}

type Output struct {
	Type   string `hcl:"type,label"`
	Width  int    `hcl:"width"`
	Height int    `hcl:"height"`
}

type Template struct {
	Output    Output     `hcl:"output,block"`
	Layers    []Layer    `hcl:"layer,block"`
	Variables []Variable `hcl:"variable,block"`
}

type Variable struct {
	Name    string `hcl:"name,label"`
	Type    string `hcl:"type"`
	Default string `hcl:"default,optional"`
}

var FileFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "path",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		source := args[0].AsString()

		pwd, err := os.Getwd()
		if err != nil {
			return cty.NullVal(cty.String), err
		}

		dir, err := ioutil.TempDir("", "assets")
		if err != nil {
			return cty.NullVal(cty.String), err
		}

		defer os.RemoveAll(dir)

		opts := []getter.ClientOption{}
		client := &getter.Client{
			Ctx:     context.Background(),
			Src:     source,
			Dst:     dir,
			Pwd:     pwd,
			Mode:    getter.ClientModeAny,
			Options: opts,
		}

		err = client.Get()
		if err != nil {
			return cty.NullVal(cty.String), err
		}

		data, err := ioutil.ReadFile(filepath.Join(dir, filepath.Base(source)))
		if err != nil {
			return cty.NullVal(cty.String), err
		}

		out := base64.StdEncoding.EncodeToString(data)

		return cty.StringVal(out), nil
	},
})

func DownloadTemplate(source string) (*[]byte, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dir, err := ioutil.TempDir("", "templates")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(dir)

	opts := []getter.ClientOption{}
	client := &getter.Client{
		Ctx:     context.Background(),
		Src:     source,
		Dst:     dir,
		Pwd:     pwd,
		Mode:    getter.ClientModeAny,
		Options: opts,
	}

	err = client.Get()
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, filepath.Base(source)))
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func ParseTemplate(source string, parameters map[string]string) (*Template, error) {
	template, err := DownloadTemplate(source)
	if err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()

	// Parse HCL files.
	f, parseDiags := parser.ParseHCL(*template, filepath.Base(source))
	if parseDiags.HasErrors() {
		for _, diag := range parseDiags {
			log.Println("parse error")
			log.Println(diag.Error())
		}
		return nil, errors.New("parse error")
	}

	// Parse variables.
	variables := map[string]cty.Value{}
	body := f.Body.(*hclsyntax.Body)
	for _, b := range body.Blocks {
		switch b.Type {
		case "variable":
			v := Variable{
				Name: b.Labels[0],
			}
			variableDiags := gohcl.DecodeBody(b.Body, nil, &v)
			if variableDiags.HasErrors() {
				for _, diag := range variableDiags {
					log.Println("variable error")
					log.Println(diag.Error())
				}
				return nil, errors.New("variable error")
			}

			// Currently only parsing strings.
			// I do not foresee any other types to be needed anytime soon.
			switch v.Type {
			case "string":
				variables[v.Name] = cty.StringVal(v.Default)
			}
		}
	}

	// Set external variables.
	for k, v := range parameters {
		variables[k] = cty.StringVal(v)
	}

	// Set variables and functions on context.
	ctx := &hcl.EvalContext{
		Variables: variables,
		Functions: map[string]function.Function{
			"file": FileFunc,
		},
	}

	// Decode template.
	var s Template
	decodeDiags := gohcl.DecodeBody(f.Body, ctx, &s)
	if decodeDiags.HasErrors() {
		for _, diag := range decodeDiags {
			log.Println("decode error")
			log.Println(diag.Error())
		}
		return nil, errors.New("decode error")
	}

	return &s, nil
}
