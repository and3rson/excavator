//go:build ignore

package main

import (
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

var packetsTemplate *template.Template = template.Must(template.New("").Parse(`// Auto-generated code. Do not edit!

package main

type PacketIDMap = map[int32]string
type DirectionMap = map[Direction]PacketIDMap
type StateMap = map[GameState]DirectionMap

var PacketNames StateMap = StateMap{
	{{- range $stateKey, $value := . }}
	{{ $stateKey }}: {
		{{- range $directionKey, $value := $value }}
		{{ $directionKey }}: {
			{{- range $id, $name := $value }}
			0x{{ printf "%02s" $id }}: "{{ $name }}",
			{{- end }}
		},
		{{- end }}
	},
	{{- end }}
}
{{ range $stateKey, $value := . }}
{{- range $directionKey, $value := $value }}
{{- range $id, $name := $value }}
const {{ if eq $directionKey "ServerBound" }}O{{ else }}I{{ end }}{{ $stateKey }}{{ $name }} = 0x{{ printf "%02s" $id }}
{{- end }}
{{- end }}
{{- end }}
`))

func main() {
	var source []byte
	var err error
	var outFile *os.File
	if source, err = os.ReadFile("packets.yaml"); err != nil {
		panic(err)
	}
	data := map[string]map[string]map[string]string{}
	if err = yaml.Unmarshal(source, data); err != nil {
		panic(err)
	}
	for state, directions := range data {
		stateName := strings.TrimPrefix(state, "GameState")
		for direction, ids := range directions {
			directionName := "In"
			// Since packets are taken from server, in = serverbound & out = clientbound. We flip this.
			if direction == "ClientBound" {
				directionName = "Out"
			}
			for id, name := range ids {
				data[state][direction][id] = strings.TrimPrefix(name, "Packet"+stateName+directionName)
			}
		}
	}
	if outFile, err = os.OpenFile("packets.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o644); err != nil {
		panic(err)
	}
	packetsTemplate.Execute(outFile, data)
}
