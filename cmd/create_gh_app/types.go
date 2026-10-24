package main

import "net/http"

type Json = map[string]any
type HandlerFunc = func(w http.ResponseWriter, r *http.Request) error
