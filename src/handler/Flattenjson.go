package handler

import (
	"encoding/json"
	"io/ioutil"

	"github.com/uneidel/sapleonardo/helper"
)

type FlattenFileCmd struct {
	In  string `required`
	Out string `required`
}

func (cmd *FlattenFileCmd) Run() error {
	file, _ := helper.ReadJson(cmd.In)
	jo := map[string]interface{}{}
	json.Unmarshal(file, &jo)
	x := cmd.Flatten(jo)
	outfile, _ := json.Marshal(x)
	ioutil.WriteFile(cmd.Out, outfile, 0644)
	return nil
}
func (cmd *FlattenFileCmd) Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := cmd.Flatten(child)
			for nk, nv := range nm {
				o[k+"_"+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}
