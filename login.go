package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Crediential struct {
	Login    string `json:"Login"`
	Password string `json:"Password"`
}

const LoginUrl = "https://f1-api.jeb.nom.za/authenticate"

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var p Crediential

	if login != "" && password != "" {
		p.Login = login
		p.Password = password
	} else {
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	jsonStr, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", LoginUrl, bytes.NewBuffer(jsonStr))
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	client := &http.Client{}
	resp, _ := client.Do(req)

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	_ = resp.Body.Close()
}
