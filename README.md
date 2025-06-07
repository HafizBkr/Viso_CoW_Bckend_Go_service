# Coworking Visio Conference Service (Go)

Ce dépôt contient le service de visioconférence du backend principal [Co-Working-B](https://github.com/HafizBkr/Co-Working-B).
Il permet la gestion des salles de réunion vidéo, l'échange de messages en temps réel et la gestion des participants dans un espace de coworking virtuel.

## Fonctionnalités

- Création de salles de visioconférence (rooms)
- Connexion WebSocket pour la signalisation WebRTC
- Authentification JWT pour sécuriser les accès
- Gestion des participants (mute, vidéo, partage d'écran)
- Persistance des messages et des salles dans MongoDB
- Intégration avec le backend principal pour la gestion des espaces de travail

## Architecture

- **Go** (Gin, Gorilla WebSocket)
- **MongoDB** pour la persistance
- **JWT** pour l'authentification
- **WebSocket** pour la signalisation temps réel

## Démarrage rapide

1. Clone ce dépôt et place-toi dans le dossier :
   ```bash
   git clone https://github.com/HafizBkr/Coworking-Visio-conf-servicexGo-.git
   cd Coworking-Visio-conf-servicexGo-


Configure les variables d'environnement dans `.env` :
   ```
   MONGO_URI=...
   MONGO_DBNAME=...
   JWT_SECRET=...
   PORT=8081
   ```

3. Installe les dépendances et lance le serveur :
   ```bash
   go mod tidy
   go run main.go
   ```

## Endpoints principaux

- `GET /health` : Vérifie que le service est opérationnel
- `POST /api/visio/room` : Crée une nouvelle salle (JWT requis)
- `GET /ws/room/:id` : Connexion WebSocket à une salle (JWT requis)

## Dépendances

- Gin
- Gorilla WebSocket
- MongoDB Go Driver
- Golang JWT

## Notes

Ce service est conçu pour fonctionner en complément du backend principal [Co-Working-B](https://github.com/HafizBkr/Co-Working-B).
Il ne gère pas l'inscription, l'authentification utilisateur ou la gestion des espaces de travail, mais s'appuie sur les JWT émis par le backend principal.
