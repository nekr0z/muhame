package handlers

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/nekr0z/muhame/internal/storage"
)

// RootHandleFunc returns the handler for the / endpoint.
func RootHandleFunc(st storage.Storage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mm, err := listAllMetrics(r.Context(), st)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")

		_, err = fmt.Fprint(w, begin)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
			return
		}

		for _, met := range mm {
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

type displayedMetric struct {
	name  string
	t     string
	value string
}

func listAllMetrics(ctx context.Context, st storage.Storage) ([]displayedMetric, error) {
	names, mm, err := st.List(ctx)
	if err != nil {
		return nil, err
	}

	mms := make([]displayedMetric, 0, len(names))
	for i, name := range names {
		mms = append(mms, displayedMetric{
			name:  name,
			t:     mm[i].Type(),
			value: mm[i].String(),
		})
	}

	slices.SortFunc(mms, func(a, b displayedMetric) int {
		if a.t != b.t {
			// gauges before counters
			return -cmp.Compare(a.t, b.t)
		}
		return cmp.Compare(a.name, b.name)
	})

	return mms, nil
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
