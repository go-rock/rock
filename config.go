package rock

type Configuration struct {
	// Defaults to "rock.view.engine".
	ViewEngineContextKey string `ini:"view_engine_context_key" json:"viewEngineContextKey,omitempty" yaml:"ViewEngineContextKey" toml:"ViewEngineContextKey"`
	// ViewLayoutContextKey is the context's values key
	// responsible to store and retrieve(string) the current view layout.
	// A middleware can modify its associated value to change
	// the layout that `ctx.View` will use to render a template.
	//
	// Defaults to "rock.view.layout".
	ViewLayoutContextKey string `ini:"view_layout_context_key" json:"viewLayoutContextKey,omitempty" yaml:"ViewLayoutContextKey" toml:"ViewLayoutContextKey"`
	// ViewDataContextKey is the context's values key
	// responsible to store and retrieve(interface{}) the current view binding data.
	// A middleware can modify its associated value to change
	// the template's data on-fly.
	//
	// Defaults to "rock.view.data".
	ViewDataContextKey string `ini:"view_data_context_key" json:"viewDataContextKey,omitempty" yaml:"ViewDataContextKey" toml:"ViewDataContextKey"`
	// FallbackViewContextKey is the context's values key
	// responsible to store the view fallback information.
	//
	// Defaults to "rock.view.fallback".
	FallbackViewContextKey string `ini:"fallback_view_context_key" json:"fallbackViewContextKey,omitempty" yaml:"FallbackViewContextKey" toml:"FallbackViewContextKey"`
}

func DefaultConfiguration() Configuration {
	return Configuration{
		ViewEngineContextKey:   "rock.view.engine",
		ViewLayoutContextKey:   "rock.view.layout",
		ViewDataContextKey:     "rock.view.data",
		FallbackViewContextKey: "rock.view.fallback",
	}
}

// GetViewDataContextKey returns the ViewDataContextKey field.
func (c *Configuration) GetViewDataContextKey() string {
	return c.ViewDataContextKey
}

// GetViewDataContextKey returns the ViewDataContextKey field.
func (c *Configuration) GetViewEngineContextKey() string {
	return c.ViewEngineContextKey
}
