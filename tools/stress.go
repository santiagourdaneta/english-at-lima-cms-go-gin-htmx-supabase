package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func runStress(name string, url string, user string, pass string) {
	total, concurrency := 250, 25 
	var wg sync.WaitGroup
	ch := make(chan struct{}, concurrency)
	success, failed := 0, 0
	var mu sync.Mutex

	start := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		ch <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-ch }()
			client := &http.Client{Timeout: 2 * time.Second}
			req, _ := http.NewRequest("GET", url, nil)
			if user != "" { req.SetBasicAuth(user, pass) }
			resp, err := client.Do(req)
			mu.Lock()
			if err == nil && resp.StatusCode == 200 { success++ } else { failed++ }
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("[%s] Exitosas: %d | Fallidas: %d | Tiempo: %v\n", name, success, failed, time.Since(start))
}

func main() {
	fmt.Println("🔥 Iniciando Tormento Dual...")
	go runStress("PÚBLICO", "http://localhost:8080/public", "", "")
	runStress("ADMIN", "http://localhost:8080/admin/", "admin", "lima2026")
}