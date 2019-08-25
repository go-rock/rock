package rock

import "github.com/go-rock/sessions"

func (c *Ctx) Session() sessions.Session {
	return c.Request().Context().Value(sessions.DefaultKey).(sessions.Session)
}

func (c *Ctx) Flashes() interface{} {
	flashes := c.Session().Flashes()
	c.Session().Save()
	return flashes
}

func (c *Ctx) AddFlash(value interface{}, vars ...string) {
	c.Session().AddFlash(value)
	c.Session().Save()
}

func (c *Ctx) SaveSession() {
	c.Session().Save()
}

func (c *Ctx) DeleteSession(key interface{}) {
	c.Session().Delete(key)
}

func (c *Ctx) ClearSession() {
	c.Session().Clear()
}
