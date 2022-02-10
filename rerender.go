package rock

import (
	"strings"
)

func (c *Ctx) HTML(name string, viewData ...interface{}) {
	c.SetHeader("Content-Type", "text/html")
	if err := c.renderView(name, viewData...); err != nil {
		c.String(500, err.Error())
	}
}

func (ctx *Ctx) ViewEngine(engine ViewEngine) {
	key := ctx.app.config.GetViewEngineContextKey()
	ctx.values.Set(key, engine)
}

func ensureTemplateName(s string, v ViewEngine) string {
	if s == "" {
		return s
	}

	s = strings.TrimPrefix(s, "/")

	if ext := v.Ext(); ext != "" {
		if !strings.HasSuffix(s, ext) {
			return s + ext
		}
	}

	return s
}

func (ctx *Ctx) renderView(filename string, optionalViewModel ...interface{}) error {
	cfg := ctx.app.config

	var bindingData interface{}
	if len(optionalViewModel) > 0 /* Don't do it: can break a lot of servers: && optionalViewModel[0] != nil */ {
		// a nil can override the existing data or model sent by `ViewData`.
		bindingData = optionalViewModel[0]
	} else {
		bindingData = ctx.values.Get(cfg.GetViewDataContextKey())
	}

	if key := cfg.GetViewEngineContextKey(); key != "" {
		if engineV := ctx.values.Get(key); engineV != nil {
			if engine, ok := engineV.(ViewEngine); ok {
				// filename := ensureTemplateName(filename, engine)
				return engine.ExecuteWriter(ctx, filename, bindingData)
			}
		}
	}
	return ctx.app.View(ctx, filename, bindingData)
}
