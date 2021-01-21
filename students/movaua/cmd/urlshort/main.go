package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"../../pkg/storage"
	"../../pkg/urlshort"
)

func main() {
	jsonFilename := flag.String("json", "pathsToURLs.json", "a JSON file")
	yamlFilename := flag.String("yaml", "pathURLs.yaml", "a YAML file")
	dbFilename := flag.String("db", "db", "a BoltDB file")
	flag.Parse()

	mux := defaultMux()

	jsonHandler, close := createJSONHandler(*jsonFilename, mux)
	defer close()

	yamlHandler, close := createYAMLHandler(*yamlFilename, jsonHandler)
	defer close()

	dbHandler, close := createDBHandler(*dbFilename, yamlHandler)
	defer close()

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", dbHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func createJSONHandler(filename string, fallback http.Handler) (http.HandlerFunc, func()) {
	file, err := os.Open(filename)
	checkError("could not open JSON file", err)
	close := func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}

	handler, err := urlshort.JSONHandler(file, fallback)
	checkError("could not create JSONHandler", err)

	return handler, close
}

func createYAMLHandler(filename string, fallback http.Handler) (http.HandlerFunc, func()) {
	file, err := os.Open(filename)
	checkError("could not open YAML file", err)
	close := func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}

	handler, err := urlshort.YAMLHandler(file, fallback)
	checkError("could not create YAMLHandler", err)

	return handler, close
}

func createDBHandler(filename string, fallback http.Handler) (http.HandlerFunc, func()) {
	store, err := storage.New(filename)
	checkError("could not create store", err)
	close := func() {
		if err := store.Close(); err != nil {
			fmt.Println(err)
		}
	}

	if err := store.Put("/b", "https://github.com/boltdb/bolt"); err != nil {
		close()
		checkError("could not put path/url to storage", err)
	}

	handler := urlshort.DBHandler(store, fallback)

	return handler, close
}

func checkError(message string, err error) {
	if err == nil {
		return
	}
	fmt.Println(message)
	fmt.Println(err)
	os.Exit(1)
}
