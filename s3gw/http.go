package s3gw

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
)

/*
	TEST `PUT`: curl -XPUT -d 'DataContent' 'http://127.0.0.1:3000/object/id42'
	TEST `GET`: curl -XGET 'http://127.0.0.1:3000/object/id42'
	TEST `404`: curl 'http://127.0.0.1:3000/object'
*/

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w,
		fmt.Sprintf("%d - Not Found", http.StatusNotFound),
		http.StatusNotFound,
	)
}

func HandleObjectPut(w http.ResponseWriter, r *http.Request, backends *Backends, bucketName string) {
	id := GetID(r, w)
	if id == "" {
		http.Error(w,
			"Invalid ID, must be alphanumeric and up to 32 characters",
			http.StatusBadRequest,
		)

		return
	}

	backendClient, backendContainerID := backends.Locate(id)
	if backendClient == nil {
		http.Error(w,
			"Failed to find S3 backend ID",
			http.StatusInternalServerError,
		)

		return
	}

	ctx := r.Context()

	err := EnsureBucketExists(ctx, backendClient, bucketName)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to ensure S3 bucket %q existance on %q",
				bucketName, backendContainerID),
			http.StatusInternalServerError,
		)

		return
	}

	if r.ContentLength == 0 {
		err = backendClient.RemoveObject(ctx, bucketName, id,
			minio.RemoveObjectOptions{
				ForceDelete: true,
			},
		)
		if err != nil {
			http.Error(w,
				fmt.Sprintf("Failed to remove object %q from %q: %v",
					id, backendContainerID, err),
				http.StatusInternalServerError,
			)

			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("Object %q deleted from %q", id, backendContainerID)

		return
	}

	defer r.Body.Close()

	_, err = backendClient.PutObject(ctx, bucketName, id, r.Body, r.ContentLength,
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to upload object %q to %q: %v",
				id, backendContainerID, err),
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Object %q uploaded to %q", id, backendContainerID)
}

func HandleObjectGet(w http.ResponseWriter, r *http.Request, backends *Backends, bucketName string) {
	id := GetID(r, w)
	if id == "" {
		http.Error(w,
			"Invalid ID, must be alphanumeric and up to 32 characters",
			http.StatusBadRequest,
		)
		return
	}

	backendClient, backendContainerID := backends.Locate(id)
	if backendClient == nil {
		http.Error(w,
			"Failed to find S3 backend ID",
			http.StatusInternalServerError,
		)

		return
	}

	ctx := r.Context()

	exists, err := backendClient.BucketExists(ctx, bucketName)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to check S3 bucket %q existance on %q",
				bucketName, backendContainerID),
			http.StatusInternalServerError,
		)

		return
	}

	if !exists {
		http.Error(w,
			fmt.Sprintf("Bucket %q not found on %q",
				bucketName, backendContainerID),
			http.StatusNotFound,
		)

		return
	}

	_, err = backendClient.StatObject(ctx, bucketName, id, minio.GetObjectOptions{})
	if err != nil {
		switch minio.ToErrorResponse(err).Code {
		case "NoSuchKey", "NotFoundObject":
			http.Error(w,
				fmt.Sprintf("Object %q not found on %q",
					id, backendContainerID),
				http.StatusNotFound,
			)

			log.Printf("Object %q not found on %q", id, backendContainerID)
		case "AccessDenied":
			http.Error(w,
				fmt.Sprintf("%d - Forbidden",
					http.StatusForbidden),
				http.StatusForbidden,
			)
		default:
			http.Error(w,
				fmt.Sprintf("Internal server error: %v",
					err),
				http.StatusInternalServerError)
		}

		return
	}

	object, err := backendClient.GetObject(ctx, bucketName, id, minio.GetObjectOptions{})
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Internal server error: %v",
				err),
			http.StatusInternalServerError)

		return
	}
	defer object.Close()

	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := io.Copy(w, object); err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to write object to response: %v",
				err),
			http.StatusInternalServerError)
	}

	log.Printf("Object %q fetched from %q", id, backendContainerID)
}
