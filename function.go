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

func (c *Function) Call(args ...interface{}) Value {
	var jsArg string
	if args != nil {
		raw, err := json.Marshal(args)
		if err != nil {
			return &value{err: err}
		}
		jsArg = fmt.Sprintf(`...%s`, string(raw))
	}
	expr := fmt.Sprintf(`window['%s']['functions'].get(%d)(%s)`, c.BindingName, c.Seq, jsArg)
	return c.ui.Eval(expr)
}
