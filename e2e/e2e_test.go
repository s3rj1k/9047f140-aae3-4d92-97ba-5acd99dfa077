package main_test

import (
	"io"
	"net/http"
	"testing"
)

func TestObjectLifecycle(t *testing.T) {
	id := generateID()
	body := generateBody()

	{ // Check none-existing object
		resp, err := httpGetObject(id)
		if err != nil {
			t.Fatalf("Failed to GET object: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	}

	{ // Create object
		resp, err := httpPutObject(id, body)
		if err != nil {
			t.Fatalf("Failed to PUT object: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d for PUT, got %d", http.StatusCreated, resp.StatusCode)
		}
	}

	{ // Get object
		resp, err := httpGetObject(id)
		if err != nil {
			t.Fatalf("Failed to GET object: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d for GET, got %d", http.StatusOK, resp.StatusCode)
		}

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if string(responseBody) != body {
			t.Errorf("Expected body %s, got %s", body, string(responseBody))
		}
	}

	{ // Delete object
		resp, err := httpPutObject(id, "")
		if err != nil {
			t.Fatalf("Failed to DELETE object: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("Expected status %d for DELETE, got %d", http.StatusNoContent, resp.StatusCode)
		}
	}

	{ // Check deletion
		resp, err := httpGetObject(id)
		if err != nil {
			t.Fatalf("Failed to GET after DELETE object: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d after DELETE, got %d", http.StatusNotFound, resp.StatusCode)
		}
	}
}

func BenchmarkObjectLifecycle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id := generateID()
		body := generateBody()

		{ // Check none-existing object
			resp, err := httpGetObject(id)
			if err != nil {
				b.Fatalf("Failed to GET object: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				b.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
			}
		}

		{ // Create object
			resp, err := httpPutObject(id, body)
			if err != nil {
				b.Fatalf("Failed to PUT object: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				b.Errorf("Expected status %d for PUT, got %d", http.StatusCreated, resp.StatusCode)
			}
		}

		{ // Get object
			resp, err := httpGetObject(id)
			if err != nil {
				b.Fatalf("Failed to GET object: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected status %d for GET, got %d", http.StatusOK, resp.StatusCode)
			}

			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				b.Errorf("Expected no error, got %v", err)
			}

			if string(responseBody) != body {
				b.Errorf("Expected body %s, got %s", body, string(responseBody))
			}
		}

		{ // Delete object
			resp, err := httpPutObject(id, "")
			if err != nil {
				b.Fatalf("Failed to DELETE object: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNoContent {
				b.Errorf("Expected status %d for DELETE, got %d", http.StatusNoContent, resp.StatusCode)
			}
		}

		{ // Check deletion
			resp, err := httpGetObject(id)
			if err != nil {
				b.Fatalf("Failed to GET after DELETE object: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				b.Errorf("Expected status %d after DELETE, got %d", http.StatusNotFound, resp.StatusCode)
			}
		}
	}
}

func BenchmarkObjectCreationP(b *testing.B) {
	body := generateBody()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := generateID()

			{ // Create object
				resp, err := httpPutObject(id, body)
				if err != nil {
					b.Fatalf("Failed to PUT object: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					b.Errorf("Expected status %d for PUT, got %d", http.StatusCreated, resp.StatusCode)
				}
			}

			{ // Get object
				resp, err := httpGetObject(id)
				if err != nil {
					b.Fatalf("Failed to GET object: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					b.Errorf("Expected status %d for GET, got %d", http.StatusOK, resp.StatusCode)
				}

				responseBody, err := io.ReadAll(resp.Body)
				if err != nil {
					b.Errorf("Expected no error, got %v", err)
				}

				if string(responseBody) != body {
					b.Errorf("Expected body %s, got %s", body, string(responseBody))
				}
			}
		}
	})
}

func TestInvalidEndpoints(t *testing.T) {
	t.Run("InvalidEndpoint GET", func(t *testing.T) {
		resp, err := http.Get("http://localhost:3000/invalidEndpoint")
		if err != nil {
			t.Fatalf("Failed to request invalid endpoint with GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d for invalid endpoint with GET, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("InvalidEndpoint PUT", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, "http://localhost:3000/invalidEndpoint", nil)
		if err != nil {
			t.Fatalf("Failed to create PUT request for invalid endpoint: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to request invalid endpoint with PUT: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status %d for invalid endpoint with PUT, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

func TestInvalidID(t *testing.T) {
	resp, err := httpGetObject("invalid-id[***]")
	if err != nil {
		t.Fatalf("Failed to GET with invalid ID: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid ID, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
