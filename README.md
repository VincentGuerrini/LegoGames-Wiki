# 🧱 LEGO Games Wiki

Un site web moderne et animé pour découvrir tous les jeux LEGO disponibles sur Steam, avec des animations impressionnantes et un thème LEGO authentique.

## 📁 Structure du Projet

```
LegoGames-Wiki/
│
├── index.html      # Page HTML principale
├── style.css       # Fichier CSS unique avec thème LEGO
├── script.js       # Fichier JavaScript unique
├── main.go         # Backend Golang (API Steam)
└── README.md       # Ce fichier
```

## ✨ Fonctionnalités

- **Présentation LEGO** : Histoire et description de LEGO en page d'accueil
- **Catalogue complet** : 29 jeux LEGO Steam avec images cliquables
- **Détails des jeux** :
  - Grande image de présentation
  - Description complète
  - Trailer vidéo
  - Screenshots
  - Succès Steam avec pourcentages de réussite
  - Lien vers la page Steam

- **Animations modernes** :
  - Briques LEGO animées en arrière-plan
  - Traînée de briques au survol de la souris
  - Animations de cartes au scroll
  - Effets hover impressionnants

- **Design authentique LEGO** :
  - Couleurs officielles LEGO
  - Motifs de tenons (studs) sur les briques
  - Typographie Bangers pour le style LEGO
  - Footer avec lien vers GitHub de Yokasashi

## 🚀 Installation et Démarrage

### Prérequis

- **Go** (Golang) installé : [Télécharger Go](https://golang.org/dl/)
- Navigateur web moderne

### Étapes de démarrage

1. **Ouvrir un terminal** dans le dossier du projet

2. **Démarrer le serveur Golang** :
   ```powershell
   go run main.go
   ```

3. **Ouvrir votre navigateur** et accéder à :
   ```
   http://localhost:8080
   ```

Le serveur Golang :
- Sert les fichiers statiques (HTML, CSS, JS)
- Agit comme proxy pour l'API Steam (évite les problèmes CORS)
- Utilise la clé API Steam fournie

## 🎮 API Steam

Le backend Golang utilise l'API Steam avec votre clé :
```
Key: A5B21889E7F05C280C98E145335D99BD
```

### Endpoints disponibles :

- `GET /api/game/{appid}` - Récupère les détails d'un jeu
- `GET /api/achievements/{appid}` - Récupère les succès d'un jeu

## 🎨 Liste des Jeux

Le site affiche 29 jeux LEGO dont :
- LEGO® Star Wars™ (4 titres)
- LEGO® Indiana Jones™ (2 titres)
- LEGO® Harry Potter™ (2 titres)
- LEGO® Marvel™ (3 titres)
- LEGO® Batman™ (4 titres)
- Et beaucoup d'autres...

## 🔧 Technologies Utilisées

- **HTML5** - Structure sémantique
- **CSS3** - Animations et thème LEGO
- **JavaScript (Vanilla)** - Interactions et appels API
- **Golang** - Backend et proxy API Steam
- **Steam Web API** - Données des jeux

## 👤 Auteur

Fait par **Yokasashi**
- GitHub : [https://github.com/Yokasashii](https://github.com/Yokasashii)

## 📝 Notes

- Ce projet est **fan-made** et non officiel
- LEGO® est une marque déposée du Groupe LEGO
- Steam® est une marque de Valve Corporation
- Les données sont récupérées via l'API officielle Steam

## 🐛 Dépannage

### Le site ne charge pas les données Steam

1. Vérifiez que le serveur Go est démarré (`go run main.go`)
2. Vérifiez que le port 8080 est disponible
3. Consultez les logs du serveur dans le terminal

### Les images ne s'affichent pas

- Les images sont chargées directement depuis les CDN Steam
- Une connexion Internet est requise

### Erreur CORS

- Le backend Golang gère automatiquement les CORS
- Assurez-vous d'accéder au site via `http://localhost:8080` et non en ouvrant directement le fichier HTML

## 📄 Licence

Projet éducatif et non commercial. Toutes les marques appartiennent à leurs propriétaires respectifs.
