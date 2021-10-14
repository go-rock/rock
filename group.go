package rock

func (app *App) Group(prefix string) *Router {
	return &Router{
		app:    app,
		prefix: prefix,
		trie:   app.router.trie,
	}
}
