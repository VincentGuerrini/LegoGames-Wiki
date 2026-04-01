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
	"golang.org/x/crypto/bcrypt"
)

var (
	STEAM_API_KEY string
	PORT          string
	dbConn        *pgx.Conn
	httpClient    *http.Client

	// In-memory admin tokens map (token -> username)
	adminTokens = map[string]string{}
)

const (
	// Local default admin credentials (local quick login)
	adminDefaultUser = "Vincent"
	adminDefaultPass = "Root"
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

		// Ensure required tables exist
		if err := ensureGamesTable(); err != nil {
			log.Printf("⚠️  Échec création/verification de la table games: %v", err)
		}
		if err := ensureAdminsTable(); err != nil {
			log.Printf("⚠️  Échec création/verification de la table admins: %v", err)
		}
	}

	// Enable CORS for all routes
	http.HandleFunc("/", enableCORS(serveStaticFiles))
	http.HandleFunc("/api/game/", enableCORS(handleGameDetails))
	http.HandleFunc("/api/achievements/", enableCORS(handleAchievements))
	// DB-backed games list and simple admin endpoints
	http.HandleFunc("/api/games", enableCORS(handleGamesList))
	http.HandleFunc("/api/admin/login", enableCORS(handleAdminLogin))
	http.HandleFunc("/api/admin/add_game", enableCORS(handleAdminAddGame))

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

	// If the DATABASE_URL contains a placeholder like YOUR_PASSWORD or [YOUR-PASSWORD],
	// substitute it with the DB_PASSWORD environment variable (URL-encoded) so secrets are kept in .env.
	if strings.Contains(databaseURL, "YOUR_PASSWORD") || strings.Contains(databaseURL, "[YOUR-PASSWORD]") {
		dbPass := os.Getenv("DB_PASSWORD")
		if dbPass == "" {
			return fmt.Errorf("DATABASE_URL contient un placeholder pour le mot de passe mais DB_PASSWORD n'est pas défini")
		}
		// Use url.PathEscape to safely encode special characters in the password
		databaseURL = strings.ReplaceAll(databaseURL, "YOUR_PASSWORD", url.PathEscape(dbPass))
		databaseURL = strings.ReplaceAll(databaseURL, "[YOUR-PASSWORD]", url.PathEscape(dbPass))
	}

	// On essaye de parser l'URL pour donner des messages plus précis si elle est mal formée
	if _, err := url.Parse(databaseURL); err != nil {
		return fmt.Errorf("DATABASE_URL invalide: %w", err)
	}

	// Utiliser un contexte avec timeout pour la connexion
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	cfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("DATABASE_URL invalide pour pgx: %w", err)
	}

	// Supabase pooler (PgBouncer) compatibility:
	// avoid prepared statement cache conflicts like:
	// "prepared statement ... already exists" (SQLSTATE 42P05)
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("échec de connexion à la base de données: %w", err)
	}

	// Test minimal de la connexion (compatible pooler, sans prepared statement)
	if _, err := conn.Exec(ctx, "SELECT 1"); err != nil {
		conn.Close(context.Background())
		return fmt.Errorf("échec du test de connexion DB: %w", err)
	}

	log.Printf("🗄️  PostgreSQL connecté (test SELECT 1 OK)")
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



// ensureGamesTable creates the games table if it doesn't exist
func ensureGamesTable() error {
	if dbConn == nil {
		return fmt.Errorf("db non initialisée")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	create := `CREATE TABLE IF NOT EXISTS public.games (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		subtitle TEXT,
		series TEXT,
		tag TEXT,
		appid INTEGER DEFAULT 0
	);`

	if _, err := dbConn.Exec(ctx, create); err != nil {
		return fmt.Errorf("échec création table games: %w", err)
	}
	return nil
}

// ensureAdminsTable creates the admins table and inserts default admin if missing
func ensureAdminsTable() error {
	if dbConn == nil {
		return fmt.Errorf("db non initialisée")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	create := `CREATE TABLE IF NOT EXISTS public.admins (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT
	);`

	if _, err := dbConn.Exec(ctx, create); err != nil {
		return fmt.Errorf("échec création table admins: %w", err)
	}

	// Ensure password_hash column exists and backfill from legacy plaintext column if present
	if _, err := dbConn.Exec(ctx, "ALTER TABLE public.admins ADD COLUMN IF NOT EXISTS password_hash TEXT"); err != nil {
		return fmt.Errorf("échec ajout colonne password_hash: %w", err)
	}
	if _, err := dbConn.Exec(ctx, "UPDATE public.admins SET password_hash = password WHERE password_hash IS NULL AND password IS NOT NULL"); err != nil {
		return fmt.Errorf("échec migration ancien mot de passe vers password_hash: %w", err)
	}

	// Ensure default admin exists — store a bcrypt hash of the password instead of plaintext
	var exists bool
	if err := dbConn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM public.admins WHERE username=$1)", adminDefaultUser).Scan(&exists); err != nil {
		// if the check fails, log and continue (we don't want startup to crash on a minor check)
		log.Printf("⚠️  Impossible de vérifier l'existence de l'admin par défaut: %v", err)
	} else if !exists {
		// generate bcrypt hash
		hash, err := bcrypt.GenerateFromPassword([]byte(adminDefaultPass), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("échec génération hash mot de passe admin: %w", err)
		}
		if _, err := dbConn.Exec(ctx, "INSERT INTO public.admins (username, password) VALUES ($1,$2)", adminDefaultUser, string(hash)); err != nil {
			return fmt.Errorf("échec insertion admin par défaut: %w", err)
		}
		log.Printf("ℹ️  Compte admin '%s' créé (mot de passe par défaut fourni)", adminDefaultUser)
	}
	// ignore errors from select/insert above to avoid failing startup for small issues
	return nil
}

// handleGamesList returns games from the DB as JSON
func handleGamesList(w http.ResponseWriter, r *http.Request) {
	if dbConn == nil {
		// return empty list instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := dbConn.Query(ctx, "SELECT id, title, subtitle, series, tag, appid FROM public.games ORDER BY id ASC")
	if err != nil {
		log.Printf("❌ erreur requête jeux: %v", err)
		respondWithError(w, "Erreur lors de la récupération des jeux", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type DBGame struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Subtitle string `json:"subtitle"`
		Series   string `json:"series"`
		Tag      string `json:"tag"`
		AppID    int64  `json:"appid"`
	}

	var games []DBGame
	for rows.Next() {
		var g DBGame
		if err := rows.Scan(&g.ID, &g.Title, &g.Subtitle, &g.Series, &g.Tag, &g.AppID); err != nil {
			log.Printf("❌ erreur lecture ligne jeux: %v", err)
			continue
		}
		games = append(games, g)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

// handleAdminLogin authenticates an admin (DB-backed or local default) and returns a token
func handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		respondWithError(w, "Payload invalide", http.StatusBadRequest)
		return
	}
	// local default admin
	if creds.Username == adminDefaultUser && creds.Password == adminDefaultPass {
		token := fmt.Sprintf("adm-%d", time.Now().UnixNano())
		adminTokens[token] = creds.Username
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "token": token})
		return
	}

	// check DB admins if available: stored password is a bcrypt hash
	if dbConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var storedHash string
		err := dbConn.QueryRow(ctx, "SELECT password_hash FROM public.admins WHERE username=$1", creds.Username).Scan(&storedHash)
		if err == nil {
			// Compare bcrypt hash
			if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)) == nil {
				token := fmt.Sprintf("adm-%d", time.Now().UnixNano())
				adminTokens[token] = creds.Username
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "token": token})
				return
			}
		}
	}

	respondWithError(w, "Identifiants invalides", http.StatusUnauthorized)
}

// helper to extract Bearer token
func getBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

// handleAdminAddGame allows an authenticated admin to insert a new game into DB
func handleAdminAddGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	token := getBearerToken(r)
	username, ok := adminTokens[token]
	if !ok || username == "" {
		respondWithError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Title    string `json:"title"`
		Subtitle string `json:"subtitle"`
		Series   string `json:"series"`
		Tag      string `json:"tag"`
		AppID    int64  `json:"appid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, "Payload invalide", http.StatusBadRequest)
		return
	}

	if dbConn == nil {
		respondWithError(w, "Base de données non disponible", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := dbConn.Exec(ctx, "INSERT INTO public.games (title, subtitle, series, tag, appid) VALUES ($1,$2,$3,$4,$5)", payload.Title, payload.Subtitle, payload.Series, payload.Tag, payload.AppID)
	if err != nil {
		log.Printf("❌ Erreur insertion jeu: %v", err)
		respondWithError(w, "Erreur lors de l'ajout du jeu", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}
