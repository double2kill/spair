package main

import (
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func main() {
	db, err := bolt.Open("spair.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/{namespace}/{key}/{value}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]
		value := vars["value"]
		tx, err := db.Begin(true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = bucket.Put([]byte(key), []byte(value))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	r.HandleFunc("/{namespace}/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]
		tx, err := db.Begin(true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		val := bucket.Get([]byte(key))
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(val)
	})
	srv := &http.Server{
		Handler: r,
		Addr:    ":28080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
