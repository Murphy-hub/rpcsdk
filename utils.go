package rpcsdk

import (
	"github.com/Murphy-hub/rpcsdk/errors"
	"io"
	"net/http"
)

func InArray(s string, slice []string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}
	return false
}

func requestError(response *http.Response) error {
	switch response.StatusCode {
	case 401:
		return errors.Unauthorized("Unauthorized")
	case 403:
		return errors.Forbidden("No permission")
	case 404:
		body, _ := io.ReadAll(response.Body)
		return errors.NotFound("not found").WithBody(body)
	default:
		body, _ := io.ReadAll(response.Body)
		return errors.InternalServerError("Request exception").WithBody(body)
	}
}
