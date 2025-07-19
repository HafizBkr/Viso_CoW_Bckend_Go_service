# Coworking Visio Conference Service (Go).

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

##  A faire apres

# Typologie des visioconférences (rooms) et gestion des accès

## Objectif

Permettre à chaque workspace de créer différents types de visioconférences, afin de répondre à des besoins variés : réunions internes, événements publics, collaborations inter-workspaces, sessions privées ou sur invitation, etc.

---

## Catégories de rooms

### 1. Room **publique**
- **Usage** : Meetups, conférences ouvertes, webinaires, présentations publiques.
- **Accès** :
  - Accessible à tous les membres de la plateforme.
  - Possibilité d’accès via un lien public pour des personnes sans compte (invités externes).
  - Option : inscription préalable ou accès direct.

### 2. Room **privée**
- **Usage** : Réunions internes, sessions confidentielles, gestion d’équipe.
- **Accès** :
  - Réservée uniquement aux membres du workspace concerné.
  - Contrôle strict via l’appartenance au workspace.

### 3. Room **sur invitation**
- **Usage** : Consultations, rendez-vous clients, sessions avec des partenaires externes.
- **Accès** :
  - Seuls les utilisateurs explicitement invités peuvent rejoindre (même s’ils ne sont pas membres du workspace).
  - Invitations possibles par email ou lien unique/token.
  - Option : accès temporaire ou limité dans le temps.

### 4. Room **cross-workspace**
- **Usage** : Projets collaboratifs entre plusieurs workspaces, partenariats, groupes de travail inter-entreprises.
- **Accès** :
  - Membres de plusieurs workspaces peuvent être invités à la même visioconférence.
  - Gestion des droits d’accès croisés.

### 5. Room **événement**
- **Usage** : Ateliers, formations, événements à capacité limitée, sessions nécessitant une inscription.
- **Accès** :
  - Inscription préalable obligatoire.
  - Limite de places, gestion de liste d’attente.
  - Possibilité de tickets ou de paiement pour l’accès.

---

## Implications techniques et UX

- Chaque room possède un champ `type` qui définit sa catégorie et sa logique d’accès.
- Le backend contrôle l’accès lors du join en fonction du type de room et du statut de l’utilisateur (membre, invité, externe…).
- Pour les rooms publiques ou événementielles, un lien d’accès public peut être généré et partagé.
- Les rooms sur invitation ou cross-workspace nécessitent une gestion fine des invitations et des droits.
- L’interface frontend doit permettre de filtrer et d’afficher les rooms selon leur type et l’accès de l’utilisateur.

---

## Bénéfices

- Flexibilité maximale pour les workspaces selon leurs besoins (confidentiel, ouvert, collaboratif…)
- Ouverture à des usages variés : coworking, formation, événementiel, consulting, etc.
- Possibilité d’attirer des utilisateurs externes ou de nouveaux membres via des événements publics ou sur invitation.


## Notes

Ce service est conçu pour fonctionner en complément du backend principal [Co-Working-B](https://github.com/HafizBkr/Co-Working-B).
Il ne gère pas l'inscription, l'authentification utilisateur ou la gestion des espaces de travail, mais s'appuie sur les JWT émis par le backend principal.
