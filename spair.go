package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

type ValueData struct {
	Value json.RawMessage `json:"value"`
	UpdateTime int64 `json:"update_time"`
}

type ListItem struct {
	Key string `json:"key"`
	UpdateTime int64 `json:"update_time"`
	Value interface{} `json:"value"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	db, err := bolt.Open("spair.db", 0644, nil)
	if err != nil {
		log.Fatal(err)
	}
	router := mux.NewRouter()

	router.Use(addResponsCORSHeader)
	router.Use(loggerHandler)

	router.HandleFunc("/{namespace}/{key}", func(w http.ResponseWriter, r *http.Request) {

    if r.Method == http.MethodOptions {
        return
    }

		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
			if err != nil {
				return err
			}
			value := bucket.Get([]byte(key))
			var data ValueData
			jsonErr:=json.Unmarshal(value,&data)
			response := []byte(data.Value)
			
			//兼容旧数据
			if(jsonErr != nil) {
				response = value
			}
			_, err = w.Write(response)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods(http.MethodGet, http.MethodOptions)

	router.HandleFunc("/{namespace}/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]

		decoder := json.NewDecoder(r.Body)
    var data ValueData
    err := decoder.Decode(&data)
    if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.UpdateTime = time.Now().UnixNano() / 1e6
		
    jsonData, err := json.Marshal(data)
    if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
    }
		
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
			if err != nil {
				return err
			}
			err = bucket.Put([]byte(key), []byte(string(jsonData)))
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
	}).Methods(http.MethodPost)

	router.HandleFunc("/{namespace}/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		namespace := vars["namespace"]
		key := vars["key"]
		err = db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(namespace))
			if err != nil {
				return err
			}
			err = bucket.Delete([]byte(key))
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods(http.MethodDelete)

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

			var list []ListItem
			b.ForEach(func(k, value []byte) error {

				var listItem ListItem
				jsonErr:=json.Unmarshal(value,&listItem)
				
				//兼容旧数据
				if(jsonErr != nil) {
					listItem.UpdateTime = 0
					listItem.Value = string(value)
				}

				listItem.Key = string(k)

				list = append(list, listItem)
				return nil
			})
			res, _ := json.Marshal(list)
			
		  w.Write(res)
			return nil
		})
	})

	router.Use(mux.CORSMethodMiddleware(router))

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

func addResponsCORSHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func loggerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time_request := time.Now()
		next.ServeHTTP(w, r)
		time_close := time.Now()
		duration := time_close.Sub(time_request)
		log.Print(r.Method + " "+  r.URL.Path + " " + duration.String())
	})
}
