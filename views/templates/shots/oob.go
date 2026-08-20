package shots

import "github.com/a-h/templ"

func rowOOBAttrs(oobMode string) templ.Attributes {
	attrs := templ.Attributes{}
	switch oobMode {
	case "insert":
		attrs["hx-swap-oob"] = "beforeend:#shots-tbody"
	case "replace":
		attrs["hx-swap-oob"] = "true"
	}
	return attrs
}
