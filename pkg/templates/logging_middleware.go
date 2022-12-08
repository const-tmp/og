package templates

var LoggingMiddleware = "{{ define \"args_printf_string\" }}\n{{- range . }}\n{{- if and (not (eq .Type.String \"context.Context\")) (not (eq .Type.String \"error\")) }}\n{{ .Name }}({{ .Type }}):   %+v\n{{- end }}\n{{- end }}\n{{- end }}\n\n{{- define \"args_printf\" }}\n{{- range . }}\n{{- if and (not (eq .Type.String \"context.Context\")) (not (eq .Type.String \"error\")) }}{{ .Name }}, {{ end -}}\n{{- end -}}\n{{ end -}}\n\npackage service\nimport (\n    \"context\"\n    \"log\"\n)\n\ntype loggingMiddleware struct {\n    l    *log.Logger\n    next {{ .Package }}.{{ .Name }}\n}\n\nfunc NewLoggingMiddleware(l *log.Logger) func(service {{ .Package }}.{{ .Name }}) {{ .Package }}.{{ .Name }} {\n    return func(service {{ .Package }}.{{ .Name }}) {{ .Package }}.{{ .Name }} {\n        return loggingMiddleware{l, service}\n    }\n}\n\n{{ range .Methods }}\nfunc (mw loggingMiddleware) {{ .Name }}{{ .Args }}{{ .Results }} {\n    mw.l.Printf(`calling {{ .Name }}:\n{{- template \"args_printf_string\" .Args }}`, {{- template \"args_printf\" .Args }})\n    {{ callArgs .Results.Args }} = mw.next.{{ .Name }}({{ callArgs .Args }})\n    if err != nil {\n        mw.l.Printf(\"{{ .Name }} error: %s\", err.Error())\n    } else {\n        mw.l.Printf(`{{ .Name }} result:\n{{- template \"args_printf_string\" .Results.Args }}`,\n{{- template \"args_printf\" .Results.Args }})\n    }\n\treturn {{ callArgs .Results.Args }}\n}\n{{ end }}\n"
