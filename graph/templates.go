// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package graph

import "text/template"

const (
	dotTemplateSrc = `digraph {
	graph[fontname="Go"];
	node[shape=box,fontname="Go"];
	{{range .Nodes}}
	"{{.Name}}" [URL="?node={{urlquery .Name}}"{{if gt .Multiplicity 1}},shape=box3d{{end}}];
	{{- end}}
	{{range .Channels}}
	"{{.Name}}" [xlabel="{{.Name}}",URL="?channel={{urlquery .Name}}",shape=point,fontname="Go Mono"];
	{{- end}}
	{{range $n := .Nodes -}}
		{{range $.DeclaredChannels .ChannelsRead}}
			{{if eq .Type "error"}}
	"{{.Name}}" -> "{{$n.Name}}" [URL="?channel={{urlquery .Name}}",color="red"];
	{rank=same "{{.Name}}" "{{$n.Name}}"}
			{{else}}
	"{{.Name}}" -> "{{$n.Name}}" [URL="?channel={{urlquery .Name}}"];
			{{- end}}
		{{- end}}
		{{- range $.DeclaredChannels .ChannelsWritten}}
		    {{if eq .Type "error"}}
	"{{$n.Name}}" -> "{{.Name}}" [URL="?channel={{urlquery .Name}}",color="red"];
	{rank=same "{{.Name}}" "{{$n.Name}}"}
			{{else}}
	"{{$n.Name}}" -> "{{.Name}}" [URL="?channel={{urlquery .Name}}"];
			{{- end}}	
		{{- end}}
	{{- end}}
}`

	goTemplateSrc = `{{if .IsCommand}}
// The {{.PackageName}} command was automatically generated by Shenzhen Go.
package main
{{else}}
// Package {{.PackageName}} was automatically generated by Shenzhen Go.
package {{.PackageName}} {{if ne .PackagePath .PackageName}} // import "{{.PackagePath}}"{{end}}
{{end}}

import (
	{{range .AllImports}}
	{{.}}
	{{- end}}
)

var (
	{{- range .Channels}}
	{{.Name}} = make(chan {{.Type}}, {{.Cap}})
	{{- end}}
)

{{if .IsCommand}}
func main() {
{{else}}
// Run executes all the goroutines associated with the graph that generated 
// this package, and waits for any that were marked as "wait for this to 
// finish" to finish before returning.
func Run() {
{{end}}
	var wg sync.WaitGroup
	{{range .Nodes}}
	
	// {{.Name}}
	{{if .Wait -}}
	wg.Add(1)
	{{- end}}

	go func() {
		{{if .Wait -}}
		defer wg.Done()
		{{end}}
		{{.ImplHead}}
		{{if eq .Multiplicity 1 -}}
		func(instanceNumber, multiplicity int) {
			{{.ImplBody}}
		}(0, 1)
		{{- else -}}
		var multWG sync.WaitGroup
		multWG.Add({{.Multiplicity}})
		for n:=0; n<{{.Multiplicity}}; n++ {
			go func(instanceNumber, multiplicity int) {
				defer multWG.Done()
				{{.ImplBody}}
			}(n, {{.Multiplicity}})
		}
		multWG.Wait()
		{{- end}}
		{{.ImplTail}}
	}()
	{{- end}}

	// Wait for the end
	wg.Wait()
}`

	goRunnerTemplateSrc = `package main

	import "{{.PackagePath}}"

	func main() {
		{{.PackageName}}.Run()
	}
`
)

var (
	dotTemplate      = template.Must(template.New("dot").Parse(dotTemplateSrc))
	goTemplate       = template.Must(template.New("golang").Parse(goTemplateSrc))
	goRunnerTemplate = template.Must(template.New("golang-runner").Parse(goRunnerTemplateSrc))
)
