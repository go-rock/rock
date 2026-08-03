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

	// TrustProxyHeaders 控制 ClientIP 是否信任 X-Real-IP / X-Forwarded-For 头。
	// 仅当应用部署在可信反向代理（nginx/haproxy 等）之后时才应开启；
	// 直接暴露在公网时这些头可被客户端伪造，开启会允许伪造客户端 IP。
	//
	// 默认 false：只使用 RemoteAddr 作为客户端 IP。
	TrustProxyHeaders bool `ini:"trust_proxy_headers" json:"trustProxyHeaders,omitempty" yaml:"TrustProxyHeaders" toml:"TrustProxyHeaders"`
}

func DefaultConfiguration() Configuration {
	return Configuration{
		ViewEngineContextKey:   "rock.view.engine",
		ViewLayoutContextKey:   "rock.view.layout",
		ViewDataContextKey:     "rock.view.data",
		FallbackViewContextKey: "rock.view.fallback",
		TrustProxyHeaders:      false, // 默认不信任代理头，防止伪造客户端 IP
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
