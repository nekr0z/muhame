package handlers

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
)

// RootHandleFunc returns the handler for the / endpoint.
func RootHandleFunc(st MetricsStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		names, mm, err := st.List()
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		type m struct {
			name  string
			t     string
			value string
		}

		mms := make([]m, 0, len(names))
		for i, name := range names {
			mms = append(mms, m{
				name:  name,
				t:     mm[i].Type(),
				value: mm[i].String(),
			})
		}

		slices.SortFunc(mms, func(a, b m) int {
			if a.t != b.t {
				return -cmp.Compare(a.t, b.t)
			}
			return cmp.Compare(a.name, b.name)
		})

		_, err = fmt.Fprint(w, begin)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		for _, met := range mms {
			link := fmt.Sprintf("/value/%s/%s", met.t, met.name)
			_, err = fmt.Fprintf(w, "<li><a href=\"%s\">%s (%s)</a>: %s</li>\n", link, met.name, met.t, met.value)
			if err != nil {
				http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
				return
			}
		}

		_, err = fmt.Fprint(w, end)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

const (
	begin = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Metrics</title>
</head>
<body>
	<h1>Metrics</h1>
	<ul>
`
	end = `</ul>
</body>
</html>
`
)
