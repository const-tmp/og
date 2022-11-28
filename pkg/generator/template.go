package generator

type Interface interface {
	Generate() error
}

//type Template struct {
//	template    *template.Template
//	extractor   any
//	transformer any
//	loader      load.Dot
//	writer      io.Writer
//}
//
//func (t Template) Generate() error {
//
//	return nil
//}
