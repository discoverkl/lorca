package lorca

import (
	"encoding/json"
	"fmt"
)

type Function struct {
	BindingName string `json:"bindingName"`
	Seq         int    `json:"seq"`

	ui UI
}

func (c *Function) Close() {
	c.ui.Eval(fmt.Sprintf(`window['%s']['functions'].delete(%d)`, c.BindingName, c.Seq))
}

func (c *Function) Call(args ...interface{}) error {
	raw, err := json.Marshal(args)
	if err != nil {
		return err
	}
	expr := fmt.Sprintf(`window['%s']['functions'].get(%d)(...%s)`, c.BindingName, c.Seq, string(raw))
	return c.ui.Eval(expr).Err()
}
