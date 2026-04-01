package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var (
	STEAM_API_KEY string
	PORT          string
	dbConn        *pgx.Conn
	httpClient    *http.Client
)

// Réponses / modèles
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

const (
	// Durées de timeout pour les appels externes
	defaultHTTPTimeout   = 10 * time.Second
	apiCallTimeout       = 8 * time.Second
	dbConnectTimeout     = 5 * time.Second
	maxJSONBodyLogLength = 1024
)

func main() {
	// Charger les variables d'environnement depuis .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Fichier .env non trouvé, utilisation des variables d'environnement système")
	}

	// Initialiser le client HTTP avec timeout
	httpClient = &http.Client{
		Timeout: defaultHTTPTimeout,
	}

	// Récupérer les variables d'environnement
	STEAM_API_KEY = os.Getenv("STEAM_API_KEY")
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080" // Valeur par défaut
	}

	if STEAM_API_KEY == "" {
		log.Fatal("❌ STEAM_API_KEY non définie dans le fichier .env")
	}

	// Initialize database connection (optionnel)
	if err := initDB(); err != nil {
		// Fournir des indices utiles pour l'utilisateur au-delà du message d'erreur brut
		log.Printf("⚠️  Base de données non disponible: %v", err)
		hint := databaseErrorHint(err)
		if hint != "" {
			log.Printf("💡 Astuce: %s", hint)
		}
		log.Printf("ℹ️  Le serveur continuera sans base de données")
	} else {
		defer dbConn.Close(context.Background())
		log.Printf("✅ Base de données connectée avec succès")
	}

	// Enable CORS for all routes
	http.HandleFunc("/", enableCORS(serveStaticFiles))
	http.HandleFunc("/api/game/", enableCORS(handleGameDetails))
	http.HandleFunc("/api/achievements/", enableCORS(handleAchievements))

	log.Printf("🧱 LEGO Games Wiki - Backend Golang")
	log.Printf("🚀 Serveur démarré sur http://localhost:%s", PORT)
	log.Printf("📡 Serveur de fichiers statiques activé")
	log.Printf("🔑 Clé API Steam chargée (masquée pour la sécurité)")

	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}

// initDB initialise la connexion à la base de données PostgreSQL
func initDB() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL non définie")
	}

	// On essaye de parser l'URL pour donner des messages plus précis si elle est mal formée
	if _, err := url.Parse(databaseURL); err != nil {
		return fmt.Errorf("DATABASE_URL invalide: %w", err)
	}

	// Utiliser un contexte avec timeout pour la connexion
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("échec de connexion à la base de données: %w", err)
	}

	// Test de la connexion
	var version string
	if err := conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		conn.Close(context.Background())
		return fmt.Errorf("échec de la requête de test: %w", err)
	}

	log.Printf("🗄️  PostgreSQL connecté: %s", version)
	dbConn = conn
	return nil
}

// databaseErrorHint retourne un indice en langage clair selon l'erreur
func databaseErrorHint(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	// Common error: invalid userinfo when password contains special characters and is not URL-encoded
	if strings.Contains(msg, "invalid userinfo") || strings.Contains(msg, "failed to parse as URL") {
		return "Le mot de passe dans DATABASE_URL contient probablement des caractères spéciaux (ex: '@', ':', '#'). URL-encode le mot de passe ou entoure correctement la partie userinfo. Exemple: postgresql://user:pass%40word@host:5432/dbname"
	}
	if strings.Contains(msg, "connection refused") || strings.Contains(strings.ToLower(msg), "no such host") {
		return "Vérifie que l'hôte et le port dans DATABASE_URL sont corrects et accessibles depuis cette machine."
	}
	return ""
}

// enableCORS permet les requêtes CORS depuis n'importe quelle origine
func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
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
	steamURL := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s&l=french", url.QueryEscape(appID))

	body, statusCode, err := httpGetWithTimeout(steamURL, apiCallTimeout)
	if err != nil {
		log.Printf("❌ Erreur lors de l'appel à l'API Steam: %v", err)
		respondWithError(w, "Erreur lors de la récupération des données Steam", http.StatusInternalServerError)
		return
	}
	if statusCode != http.StatusOK {
		log.Printf("❌ API Steam retourné status %d pour %s", statusCode, steamURL)
		respondWithError(w, "Erreur externe: Steam API", http.StatusBadGateway)
		return
	}

	// Parse the response
	var result map[string]GameDetailsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logJSONParseError("appdetails", body, err)
		respondWithError(w, "Erreur lors du parsing des données Steam", http.StatusInternalServerError)
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
		url.QueryEscape(STEAM_API_KEY), url.QueryEscape(appID))

	schemaBody, statusCode, err := httpGetWithTimeout(schemaURL, apiCallTimeout)
	if err != nil {
		log.Printf("❌ Erreur lors de l'appel à l'API Steam (schema): %v", err)
		respondWithError(w, "Erreur lors de la récupération des succès", http.StatusInternalServerError)
		return
	}
	if statusCode != http.StatusOK {
		log.Printf("❌ API Steam (schema) retourné status %d pour %s", statusCode, schemaURL)
		respondWithError(w, "Erreur externe: Steam API (schema)", http.StatusBadGateway)
		return
	}

	var schemaData map[string]interface{}
	if err := json.Unmarshal(schemaBody, &schemaData); err != nil {
		logJSONParseError("schema", schemaBody, err)
		respondWithError(w, "Erreur lors du parsing des données de succès", http.StatusInternalServerError)
		return
	}

	// 2. Get achievement percentages
	percentURL := fmt.Sprintf("https://api.steampowered.com/ISteamUserStats/GetGlobalAchievementPercentagesForApp/v2/?gameid=%s&format=json", url.QueryEscape(appID))

	percentBody, statusCode2, err := httpGetWithTimeout(percentURL, apiCallTimeout)
	if err != nil {
		log.Printf("❌ Erreur lors de l'appel à l'API Steam (percentages): %v", err)
		respondWithError(w, "Erreur lors de la récupération des pourcentages", http.StatusInternalServerError)
		return
	}
	if statusCode2 != http.StatusOK {
		log.Printf("❌ API Steam (percentages) retourné status %d pour %s", statusCode2, percentURL)
		respondWithError(w, "Erreur externe: Steam API (percentages)", http.StatusBadGateway)
		return
	}

	var percentData map[string]interface{}
	if err := json.Unmarshal(percentBody, &percentData); err != nil {
		logJSONParseError("percentages", percentBody, err)
		respondWithError(w, "Erreur lors du parsing des pourcentages", http.StatusInternalServerError)
		return
	}

	// Build percentages map
	percentages := make(map[string]float64)
	if achievementPercentages, ok := percentData["achievementpercentages"].(map[string]interface{}); ok {
		if achievements, ok := achievementPercentages["achievements"].([]interface{}); ok {
			for _, ach := range achievements {
				if achMap, ok := ach.(map[string]interface{}); ok {
					name, _ := achMap["name"].(string)
					// percent might be float64 or other numeric type
					switch v := achMap["percent"].(type) {
					case float64:
						percentages[name] = v
					case float32:
						percentages[name] = float64(v)
					case int:
						percentages[name] = float64(v)
					case int64:
						percentages[name] = float64(v)
					default:
						// try to decode via json marshal/unmarshal as float64 fallback
						var tmp float64
						if b, err := json.Marshal(achMap["percent"]); err == nil {
							_ = json.Unmarshal(b, &tmp)
							percentages[name] = tmp
						} else {
							percentages[name] = 0.0
						}
					}
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

// Helper: effectue un GET avec timeout et renvoie le corps et le status
func httpGetWithTimeout(urlStr string, timeout time.Duration) ([]byte, int, error) {
	// Utiliser un contexte avec timeout spécifique à l'appel
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		// Si contexte dépassé, renvoyer une erreur claire
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, 0, fmt.Errorf("timeout lors de l'appel HTTP à %s", urlStr)
		}
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024)) // limiter lecture à 10MB pour éviter OOM
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

// respondWithError envoie une réponse d'erreur JSON
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// logJSONParseError logge des informations utiles lors d'une erreur de parsing JSON
func logJSONParseError(kind string, body []byte, parseErr error) {
	trimmed := string(body)
	if len(trimmed) > maxJSONBodyLogLength {
		trimmed = trimmed[:maxJSONBodyLogLength] + "...(truncated)"
	}
	log.Printf("❌ Erreur parsing JSON (%s): %v -- body: %s", kind, parseErr, trimmed)
}
