package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"code.local/homework-object-storage/s3gw"
)

func init() { // KISS configuration
	os.Setenv(s3gw.S3ContainerNamePatternEnvKey, "amazin-object-storage-node")
	os.Setenv(s3gw.S3APIPortEnvKey, "9000")

	os.Setenv(s3gw.ConsistentHashPartitionCountEnvKey, "71")
	os.Setenv(s3gw.ConsistentHashReplicationFactorEnvKey, "20")
	os.Setenv(s3gw.ConsistentHashLoadEnvKey, "1.25")

	os.Setenv(s3gw.S3DefaultBucketNameEnvKey, "objects")
}

func main() {
	backends, err := s3gw.Configure(context.Background())
	if err != nil {
		log.Fatal(s3gw.CapitalizeErrorString(err))
	}

	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s3gw.HandleNotFound(w, r)
	})

	defaultBuckerName := os.Getenv(s3gw.S3DefaultBucketNameEnvKey)

	r.HandleFunc("/object/{id}", func(w http.ResponseWriter, r *http.Request) {
		s3gw.HandleObjectPut(w, r, backends, defaultBuckerName)
	}).Methods(http.MethodPut)

	r.HandleFunc("/object/{id}", func(w http.ResponseWriter, r *http.Request) {
		s3gw.HandleObjectGet(w, r, backends, defaultBuckerName)
	}).Methods(http.MethodGet)

	r.HandleFunc("/object", func(w http.ResponseWriter, r *http.Request) {
		s3gw.HandleObjectList(w, r, backends, defaultBuckerName)
	}).Methods(http.MethodGet)

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatalf("Failed to start S3 Gateway service: %v", err)
	}
}
