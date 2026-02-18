package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	STEAM_API_KEY = "A5B21889E7F05C280C98E145335D99BD"
	PORT          = "8080"
)

type GameDetailsResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type AchievementResponse struct {
	Success      bool                 `json:"success"`
	Achievements []AchievementWithPct `json:"achievements"`
}

type AchievementWithPct struct {
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Percent     float64 `json:"percent"`
}

func main() {
	// Enable CORS for all routes
	http.HandleFunc("/", enableCORS(serveStaticFiles))
	http.HandleFunc("/api/game/", enableCORS(handleGameDetails))
	http.HandleFunc("/api/achievements/", enableCORS(handleAchievements))

	log.Printf("🧱 LEGO Games Wiki - Backend Golang")
	log.Printf("🚀 Serveur démarré sur http://localhost:%s", PORT)
	log.Printf("📡 Serveur de fichiers statiques activé")
	log.Printf("🔑 Utilisation de la clé API Steam: %s", STEAM_API_KEY)

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}

// enableCORS permet les requêtes CORS depuis n'importe quelle origine
func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// serveStaticFiles sert les fichiers HTML, CSS, JS
func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Remove leading slash
	filePath := strings.TrimPrefix(path, "/")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Set content type based on extension
	if strings.HasSuffix(filePath, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(filePath, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasSuffix(filePath, ".html") {
		w.Header().Set("Content-Type", "text/html")
	}

	http.ServeFile(w, r, filePath)
}

// handleGameDetails récupère les détails d'un jeu depuis l'API Steam
func handleGameDetails(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimPrefix(r.URL.Path, "/api/game/")
	if appID == "" {
		respondWithError(w, "AppID manquant", http.StatusBadRequest)
		return
	}

	log.Printf("📥 Récupération des détails du jeu: %s", appID)

	// Appel à l'API Steam Store
	steamURL := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s&l=french", appID)

	resp, err := http.Get(steamURL)
	if err != nil {
		log.Printf("❌ Erreur lors de l'appel à l'API Steam: %v", err)
		respondWithError(w, "Erreur lors de la récupération des données Steam", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Erreur lors de la lecture de la réponse: %v", err)
		respondWithError(w, "Erreur lors de la lecture des données", http.StatusInternalServerError)
		return
	}

	// Parse the response
	var result map[string]GameDetailsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Erreur lors du parsing JSON: %v", err)
		respondWithError(w, "Erreur lors du parsing des données", http.StatusInternalServerError)
		return
	}

	// Get the game data
	gameData, exists := result[appID]
	if !exists || !gameData.Success {
		log.Printf("❌ Jeu non trouvé: %s", appID)
		respondWithError(w, "Jeu non trouvé", http.StatusNotFound)
		return
	}

	log.Printf("✅ Détails du jeu %s récupérés avec succès", appID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameData)
}

// handleAchievements récupère les succès d'un jeu depuis l'API Steam
func handleAchievements(w http.ResponseWriter, r *http.Request) {
	appID := strings.TrimPrefix(r.URL.Path, "/api/achievements/")
	if appID == "" {
		respondWithError(w, "AppID manquant", http.StatusBadRequest)
		return
	}

	log.Printf("🏆 Récupération des succès du jeu: %s", appID)

	// 1. Get achievement schema (names, descriptions, icons)
	schemaURL := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetSchemaForGame/v2/?key=%s&appid=%s&l=french&format=json",
		url.QueryEscape(STEAM_API_KEY), appID)

	schemaResp, err := http.Get(schemaURL)
	if err != nil {
		log.Printf("❌ Erreur lors de l'appel à l'API Steam (schema): %v", err)
		respondWithError(w, "Erreur lors de la récupération des succès", http.StatusInternalServerError)
		return
	}
	defer schemaResp.Body.Close()

	schemaBody, _ := io.ReadAll(schemaResp.Body)
	var schemaData map[string]interface{}
	json.Unmarshal(schemaBody, &schemaData)

	// 2. Get achievement percentages
	percentURL := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/?gameid=%s&format=json", appID)

	percentResp, err := http.Get(percentURL)
	if err != nil {
		log.Printf("❌ Erreur lors de l'appel à l'API Steam (percentages): %v", err)
		respondWithError(w, "Erreur lors de la récupération des pourcentages", http.StatusInternalServerError)
		return
	}
	defer percentResp.Body.Close()

	percentBody, _ := io.ReadAll(percentResp.Body)
	var percentData map[string]interface{}
	json.Unmarshal(percentBody, &percentData)

	// Build percentages map
	percentages := make(map[string]float64)
	if achievementPercentages, ok := percentData["achievementpercentages"].(map[string]interface{}); ok {
		if achievements, ok := achievementPercentages["achievements"].([]interface{}); ok {
			for _, ach := range achievements {
				if achMap, ok := ach.(map[string]interface{}); ok {
					name, _ := achMap["name"].(string)
					percent, _ := achMap["percent"].(float64)
					percentages[name] = percent
				}
			}
		}
	}

	// Extract achievements from schema
	var achievementsList []AchievementWithPct

	if game, ok := schemaData["game"].(map[string]interface{}); ok {
		if availableGameStats, ok := game["availableGameStats"].(map[string]interface{}); ok {
			if achievements, ok := availableGameStats["achievements"].([]interface{}); ok {
				for _, ach := range achievements {
					if achMap, ok := ach.(map[string]interface{}); ok {
						name, _ := achMap["name"].(string)
						displayName, _ := achMap["displayName"].(string)
						description, _ := achMap["description"].(string)
						icon, _ := achMap["icon"].(string)

						achievementsList = append(achievementsList, AchievementWithPct{
							Name:        name,
							DisplayName: displayName,
							Description: description,
							Icon:        icon,
							Percent:     percentages[name],
						})
					}
				}
			}
		}
	}

	log.Printf("✅ %d succès récupérés pour le jeu %s", len(achievementsList), appID)

	response := AchievementResponse{
		Success:      true,
		Achievements: achievementsList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// respondWithError envoie une réponse d'erreur JSON
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
