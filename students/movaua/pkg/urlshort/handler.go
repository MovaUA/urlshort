package urlshort

import (
	"encoding/json"
	"io"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dest, ok := pathsToUrls[r.URL.Path]
		if !ok {
			fallback.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, dest, http.StatusFound)
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml io.Reader, fallback http.Handler) (http.HandlerFunc, error) {
	var pathURLs []pathURL
	if err := yaml.NewDecoder(yml).Decode(&pathURLs); err != nil {
		return nil, err
	}

	pathsToURLs := make(map[string]string, len(pathURLs))
	for _, item := range pathURLs {
		pathsToURLs[item.Path] = item.URL
	}

	return MapHandler(pathsToURLs, fallback), nil
}

type pathURL struct {
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`
}

// JSONHandler will parse the provided JSON and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the JSON, then the
// fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
//
// {
//    "path1": "url1",
//    "path2": "url2"
// }
//
// The only errors that can be returned all related to having
// invalid JSON data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func JSONHandler(jsonReader io.Reader, fallback http.Handler) (http.HandlerFunc, error) {
	var pathsToURLs map[string]string
	if err := json.NewDecoder(jsonReader).Decode(&pathsToURLs); err != nil {
		return nil, err
	}

	return MapHandler(pathsToURLs, fallback), nil
}

// DBHandler will search path in the provided store and then return an http.HandleFunc
func DBHandler(store Store, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dest, ok := store.Get(r.URL.Path)
		if !ok {
			fallback.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, dest, http.StatusFound)
	}
}

// Store is an interface to interact with storage
type Store interface {
	// Get retrieves URL from the storage by the provided path.
	// It returns false if the path is not found.
	Get(path string) (url string, found bool)
}
