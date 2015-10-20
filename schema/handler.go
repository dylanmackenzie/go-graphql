package schema

import (
	"io"
	"log"
	"net/http"
	"strings"

	"dylanmackenzie.com/graphql/ast"
)

type RequestInfo struct {
	Document  io.Reader
	Variables string
	Operation string
}

// query:  A string GraphQL document to be executed.
//
// variables: The runtime values to use for any GraphQL query variables
// as a JSON object.
//
// operationName: If the provided query contains multiple named
// operations, this specifies which operation should be executed.  If
// not provided, a 400 error will be returned if the query contains
// multiple named operations.
func (sch *Schema) Handler() http.Handler {
	sch.Finalize()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET requests can't have bodies, so parse the query parameters
		// for information about the GraphQL request
		info := RequestInfo{}
		q := r.URL.Query()
		switch r.Method {
		case "GET":
			doc := q.Get("query")
			if doc == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("No GraphQL query present"))
				return
			}

			info.Document = strings.NewReader(doc)

		case "POST":
			info.Document = r.Body

		default:
			w.Header().Add("Allow", "GET, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		info.Variables = q.Get("variables")
		info.Operation = q.Get("operationName")

		doc, err := ast.FromReader(info.Document)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		ctx, err := Execute(sch, &doc, info.Operation)
		if err != nil {
			log.Printf("%#v", ctx.Errors)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.Response.MarshalJSON()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(res)
	})
}

func Handler() http.Handler {
	return def.Handler()
}
