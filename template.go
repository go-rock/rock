package rock

// //
// type PgEngine struct {
// 	name string
// }

// func NewPgEngine(name string) *PgEngine {
// 	return &PgEngine{name}
// }

// func (e *PgEngine) Name() string {
// 	return e.name
// }

// func (e *PgEngine) Render(w io.Writer, tmplName, layoutName string, data interface{}) error {
// 	w.Write([]byte("pg engine"))
// 	return nil
// }

// // ExecuteWriter renders a template on "w".
// func (s *PgEngine) ExecuteWriter(w io.Writer, tmplName, layoutName string, data interface{}) error {
// 	if layoutName == NoLayout {
// 		layoutName = ""
// 	}

// 	pp.Println("pg render from")

// 	return s.Render(w, tmplName, layoutName, data)
// }
