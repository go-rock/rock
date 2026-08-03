package rock

import "testing"

func TestDefaultConfiguration(t *testing.T) {
	c := DefaultConfiguration()

	if c.GetViewDataContextKey() != "rock.view.data" {
		t.Errorf("ViewDataContextKey 应为 rock.view.data, got %q", c.GetViewDataContextKey())
	}
	if c.GetViewEngineContextKey() != "rock.view.engine" {
		t.Errorf("ViewEngineContextKey 应为 rock.view.engine, got %q", c.GetViewEngineContextKey())
	}
	if c.TrustProxyHeaders {
		t.Error("默认不应信任代理头")
	}
}

func TestConfigurationTrustProxy(t *testing.T) {
	c := DefaultConfiguration()
	c.TrustProxyHeaders = true
	app := New()
	app.config = &c
	app.SetTrustProxy(true)
	if !app.ConfigurationReadOnly().TrustProxyHeaders {
		t.Error("SetTrustProxy(true) 应开启")
	}
}
