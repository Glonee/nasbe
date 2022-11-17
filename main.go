package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type SinglePageFs struct {
	http.FileSystem
}

func (fs SinglePageFs) Open(name string) (http.File, error) {
	file, err := fs.FileSystem.Open(name)
	if err != nil {
		file, err = fs.FileSystem.Open("index.html")
	}
	return file, err
}

type MyFileHandler struct {
	http.Handler
}

func (h MyFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("download") == "true" {
		w.Header().Set("Content-Disposition", "attachment")
	}
	h.Handler.ServeHTTP(w, r)
}
func main() {
	http.Handle("/", http.FileServer(SinglePageFs{http.Dir("build")}))
	http.Handle("/files/", http.StripPrefix("/files/", MyFileHandler{http.FileServer(http.Dir("files"))}))
	http.HandleFunc("/listFiles/", listFiles)
	http.HandleFunc("/fileInfo/", fileInfo)
	http.ListenAndServe(":8080", nil)
}
func listFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	type Data struct {
		Name  string `json:"name"`
		IsDir bool   `json:"isDir"`
		Size  int64  `json:"size"`
	}
	type Resp struct {
		Exist bool   `json:"exist"`
		Data  []Data `json:"data"`
	}
	var resp Resp
	encoder := json.NewEncoder(w)
	path := r.URL.Path[len("/listFiles"):]
	entries, err := os.ReadDir("files" + path)
	if err != nil {
		resp.Exist = false
		encoder.Encode(resp)
		return
	}
	resp.Exist = true
	resp.Data = make([]Data, len(entries))
	for i := range entries {
		info, _ := entries[i].Info()
		resp.Data[i].IsDir = info.IsDir()
		resp.Data[i].Name = info.Name()
		resp.Data[i].Size = info.Size()
	}
	encoder.Encode(resp)
}
func fileInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; utf-8")
	type Data struct {
		Exist bool  `json:"exist"`
		Size  int64 `json:"size"`
	}
	var data Data
	encoder := json.NewEncoder(w)
	path := r.URL.Path[len("/fileInfo"):]
	info, err := os.Stat("files" + path)
	if err != nil || info.IsDir() {
		data.Exist = false
		encoder.Encode(data)
		return
	}
	data.Exist = true
	data.Size = info.Size()
	encoder.Encode(data)
}
