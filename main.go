package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	// "os"
	"path/filepath"
	"strings"
	// "strings"
	// "time"
)

var client = redis.NewClient(&redis.Options{
	Addr:     "cache:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func sfmt(tmpl string, a ...interface{}) string {
	return fmt.Sprintf(tmpl, a...)
}

func write(w http.ResponseWriter, content string) {
	w.Write([]byte(content))
}

type SeedResults struct {
	Succeeded []string          `json:"succeeded"`
	Failed    map[string]string `json:"failed"`
}

type RedisKeys []string

func main() {
	router := mux.NewRouter()
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./node_modules"))))
	router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir("./ui"))))
	router.HandleFunc("/api/cache/seed", func(w http.ResponseWriter, r *http.Request) {
		seedsDir := "./seeds"
		files, err := ioutil.ReadDir(seedsDir)
		if err != nil {
			write(w, err.Error())
		}
		results := SeedResults{
			Succeeded: []string{},
			Failed:    make(map[string]string),
		}

		seeds := make(map[string]string)

		for _, f := range files {
			full := f.Name()
			ext := filepath.Ext(full)

			if f.IsDir() || ext != ".xml" {
				continue
			}

			base := strings.TrimSuffix(full, ext)
			seeds[base] = f.Name()
		}

		for k, f := range seeds {
			content, err := ioutil.ReadFile("./seeds/" + f)
			if err != nil {
				results.Failed[f] = err.Error()
				continue
			}

			client.Set(k, string(content), 0)
			results.Succeeded = append(results.Succeeded, f)
		}

		json, err := json.Marshal(results)

		if err != nil {
			panic(err)
		}
		write(w, sfmt(`{"data": %s }`, string(json)))
	})
	router.HandleFunc("/api/cache/keys", func(w http.ResponseWriter, r *http.Request) {
		keys, err := client.Keys("*").Result()
		if err != nil {
			write(w, sfmt(`{"data": "%s" }`, err.Error()))
			return
		}

		json, err := json.Marshal(keys)
		if err != nil {
			panic(err)
		}

		write(w, sfmt(`{"data": %s}`, json))
	})

	router.HandleFunc("/api/cache/{action}/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		action := vars["action"]
		key := vars["key"]

		switch action {
		case "get":
			xml, err := client.Get(key).Result()
			if err == redis.Nil {
				write(w, sfmt(`{ "data": "key [%s] not found"}`, key))
				return
			}

			if err != nil {
				write(w, sfmt(`{ "data": "%s" }`, err.Error()))
				return
			}

			write(w, xml)
			return

		case "set":
			var posted []byte

			posted, err := ioutil.ReadAll(r.Body)

			log.Printf("%#v", posted)

			if err != nil {
				write(w, sfmt(`{ "error": "%s" }`, err.Error()))
				return
			}

			log.Print(sfmt("POSTED %s", string(posted[:len(posted)])))

			client.Set(key, string(posted[:len(posted)]), 0)

			write(w, sfmt(`{ "data": "key %s was set"}`, key))

		case "seed":

		default:
			write(w, `{ "data": "404" }`)
			return
		}

	})

	log.Fatal(http.ListenAndServe(":5555", router))
}
