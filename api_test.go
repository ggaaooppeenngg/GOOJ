package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPI(t *testing.T) {
	inTest = true
	r := NewEngine()
	var ret struct {
		Id int `json:"id"`
	}
	{
		req, err := http.NewRequest("POST", "/problem", strings.NewReader(`{"title": "a+b","timeLimit": 1000, "memoryLimit": 1000, "description": "a plus b", "input": "1\n1 2\n","output": "3\n"}`))
		if err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != 200 {
			t.Fatalf("POST /problem failed, response: %s\n", res.Body.Bytes())
		}
		err = json.Unmarshal(res.Body.Bytes(), &ret)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		req, err := http.NewRequest("GET", fmt.Sprintf("/problem/%d", ret.Id), nil)
		if err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != 200 {
			t.Fatalf("GET /problem/%d failed, response: %s\n", ret.Id, res.Body.Bytes())
		}
	}
	{
		req, err := http.NewRequest("POST", "/code", strings.NewReader(`{"source":"#include <stdio.h>\r\nint main()\r\n{\r\n       \tint n;\r\n       \tint x,y;\r\n       \tscanf(\"%d\",&n);\r\n       \tfor(int i = 0; i < n; i++)\r\n       \t{\r\n       \t\tscanf(\"%d %d\",&x,&y);\r\n       \t\tprintf(\"%d\\n\",x+y);\r\n       \t}\r\n}","language":"c","problemId":6}`))
		if err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != 200 {
			t.Fatalf("POST /code failed, response: %s\n", res.Body.Bytes())
		}
		err = json.Unmarshal(res.Body.Bytes(), &ret)
		if err != nil {
			t.Fatal(err)
		}

	}
	{
		req, err := http.NewRequest("GET", fmt.Sprintf("/code/%d", ret.Id), nil)
		if err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != 200 {
			t.Fatalf("GET /code/%d failed, response: %s\n", ret.Id, res.Body.Bytes())
		}

	}

}
