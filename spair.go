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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := bolt.Open("spair.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()

	router.Use(addAccessControlAllowOrigin)
	router.Use(loggerHandler)

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

	PORT := "28080"
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + PORT,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Print("Server running at http://127.0.0.1:" + PORT)
	log.Fatal(srv.ListenAndServe())
}

func addAccessControlAllowOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func loggerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time_request := time.Now()
		next.ServeHTTP(w, r)
		time_close := time.Now()
		duration := time_close.Sub(time_request)
		log.Print(r.URL.Path + " " + duration.String())
	})
}
