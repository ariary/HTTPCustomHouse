package response

import "net/http"

type Response struct {
	Status  int
	Headers http.Header
	Body    []byte
}
