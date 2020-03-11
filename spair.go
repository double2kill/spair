package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

func main() {
	db, err := bolt.Open("spair.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()

	router.Use(addAccessControlAllowOrigin)

	router.HandleFunc("/{namespace}/{key}/{value}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]
		value := vars["value"]
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(key), []byte(value))
			if err != nil {
				return err
			}
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("ok"))
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	router.HandleFunc("/{namespace}/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
			if err != nil {
				return err
			}
			value := bucket.Get([]byte(key))
			_, err = w.Write(value)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods(http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodOptions)

	router.HandleFunc("/{namespace}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(namespace))
			if err != nil {
				return err
			}
			return err
		})
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(namespace))

			var st []string
			b.ForEach(func(k, v []byte) error {
				st = append(st, string(k))
				return nil
			})
			res, err := json.Marshal(st)
			if err != nil {
			}
			
		  w.Write(res)
			return nil
		})
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         ":28080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func addAccessControlAllowOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
