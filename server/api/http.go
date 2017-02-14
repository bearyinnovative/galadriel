package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Req struct {
	w http.ResponseWriter
	r *http.Request
}

func (r Req) requireMethod(method string) bool {
	if r.r.Method != method {
		r.w.WriteHeader(404)
		return false
	}

	return true
}

func (r Req) responseJSON(data json.Marshaler) {
	marshaledData, err := data.MarshalJSON()
	if err != nil {
		panic(err)
	}

	_, err = r.w.Write(marshaledData)
	if err != nil {
		panic(err)
	}
}

func (r Req) responseErr(err error) {
	r.responseJSON(errResp{err})
}

func (r Req) created(entity json.Marshaler) {
	r.w.WriteHeader(201)
	r.responseJSON(entity)
}

func (r Req) badRequest(err error) {
	r.w.WriteHeader(400)
	r.responseErr(err)
}

func (r Req) notFound(err error) {
	r.w.WriteHeader(404)
	r.responseErr(err)
}

func (r Req) gone(err error) {
	r.w.WriteHeader(410)
	r.responseErr(err)
}

func (r Req) serverError(err error) {
	r.w.WriteHeader(500)
	r.responseErr(err)
}

func (r Req) queryValue(key string) (string, bool) {
	values := r.r.URL.Query()

	if v, exists := values[key]; exists {
		return v[0], true
	} else {
		return "", false
	}
}

func (r Req) queryValueInt(key string) (int, bool) {
	if s, exists := r.queryValue(key); exists {
		if value, err := strconv.ParseInt(s, 10, 32); err == nil {
			return int(value), true
		}
	}
	return 0, false
}

type errResp struct {
	err error
}

func (e errResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"error": e.err.Error()})
}

type mapResp map[string]interface{}

func (m mapResp) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(m))
}
