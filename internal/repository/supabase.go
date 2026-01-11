package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// --- MOTOR PRINCIPAL ---

// CallSupabase es la única función que toca la red
func CallSupabase(method, table string, body interface{}, filter string) (*http.Response, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", os.Getenv("SUPABASE_URL"), table)
	if filter != "" {
		url += "?" + filter
	}

	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req, _ := http.NewRequest(method, url, &buf)
	req.Header.Set("apikey", os.Getenv("SUPABASE_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	client := &http.Client{}
	return client.Do(req)
}

// handleResponse procesa la respuesta de CallSupabase para ahorrar repetición
func handleResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error supabase: %d - %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// --- IMPLEMENTACIÓN DE INSERTS ---

func InsertSentence(english, spanish string) error {
	data := map[string]interface{}{"english": english, "spanish": spanish}
	return handleResponse(CallSupabase("POST", "sentences", data, ""))
}

func InsertResource(t, u, r string) error {
	data := map[string]interface{}{"title": t, "url": u, "type": r}
	return handleResponse(CallSupabase("POST", "resources", data, ""))
}

func InsertQuiz(q string, o []string, c string) error {
	data := map[string]interface{}{"question": q, "options": o, "correct": c}
	return handleResponse(CallSupabase("POST", "quizzes", data, ""))
}

// --- IMPLEMENTACIÓN DE UPDATES (PATCH) ---

func patchToSupabase(table string, id string, data map[string]interface{}) error {
	filter := fmt.Sprintf("id=eq.%s", id)
	return handleResponse(CallSupabase("PATCH", table, data, filter))
}

func UpdateSentence(id, t, tr string) error {
	return patchToSupabase("sentences", id, map[string]interface{}{"english": t, "spanish": tr})
}

func UpdateResource(id, t, u, r string) error {
	return patchToSupabase("resources", id, map[string]interface{}{"title": t, "url": u, "type": r})
}

func UpdateQuiz(id, q string, o []string, c string) error {
	return patchToSupabase("quizzes", id, map[string]interface{}{"question": q, "options": o, "correct": c})
}

// --- AUTENTICACIÓN ---

func AuthenticateUser(email, password string) (string, error) {
	url := os.Getenv("SUPABASE_URL") + "/auth/v1/token?grant_type=password"
	authData := map[string]string{"email": email, "password": password}
	jsonData, _ := json.Marshal(authData)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("apikey", os.Getenv("SUPABASE_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login fallido: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no se encontró el token")
	}
	return token, nil
}
