package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func GetAuditLogs() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/rest/v1/audit_logs?select=*&order=created_at.desc", os.Getenv("SUPABASE_URL"))
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", os.Getenv("SUPABASE_KEY"))
	req.Header.Set("Authorization", "Bearer "+os.Getenv("SUPABASE_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var logs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func FetchAllBannedIPs() ([]string, error) {
	resp, err := CallSupabase("GET", "blacklisted_ips", nil, "select=ip")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []struct {
		IP string `json:"ip"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	var ips []string
	for _, r := range results {
		ips = append(ips, r.IP)
	}
	return ips, nil
}

// InsertAuditLog guarda el intento de intrusión
func InsertAuditLog(ip, event, data string) error {
	payload := map[string]interface{}{
		"ip_address": ip,
		"event_type": event,
		"input_data": data,
	}
	// Cambiado: Pasamos "POST", el payload, y un string vacío para el filtro
	resp, err := CallSupabase("POST", "audit_logs", payload, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// BanIP registra el baneo permanente
func BanIP(ip, reason string) error {
	payload := map[string]interface{}{
		"ip":     ip,
		"reason": reason,
	}
	// Cambiado: Ajustado a los 4 argumentos requeridos
	resp, err := CallSupabase("POST", "blacklisted_ips", payload, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
