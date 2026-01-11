package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// performRequest aísla la lógica HTTP para reducir la complejidad cognitiva
func performRequest(client *http.Client, url, user, pass string) bool {
	req, _ := http.NewRequest("GET", url, nil)
	if user != "" {
		req.SetBasicAuth(user, pass)
	}

	resp, err := client.Do(req)
	if err == nil && resp != nil {
		defer resp.Body.Close()
		return resp.StatusCode == 200
	}
	return false
}

func runStress(name string, url string, user string, pass string) {
	total, concurrency := 250, 25
	var wg sync.WaitGroup
	ch := make(chan struct{}, concurrency)
	success, failed := 0, 0
	var mu sync.Mutex

	start := time.Now()
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < total; i++ {
		wg.Add(1)
		ch <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-ch }()

			isOk := performRequest(client, url, user, pass)

			mu.Lock()
			if isOk {
				success++
			} else {
				failed++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("[%s] Exitosas: %d | Fallidas: %d | Tiempo: %v\n", name, success, failed, time.Since(start))
}

func main() {
	fmt.Println("🔥 Iniciando Tormento...")
	runStress("PÚBLICO", "http://localhost:8080/public", "", "")
}
