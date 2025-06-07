## Messages WebSocket échangés

La communication en temps réel entre clients et serveur s’effectue via des messages WebSocket structurés comme suit :

   ```json
              {
                "type": "<event_type>",
                "data": { ... }
              }
   ```


Types de messages

- **join**
  - Envoyé lorsqu’un utilisateur rejoint une salle.
  - Exemple :
    ```json
    { "type": "join", "data": { "user": "Hafiz" } }
    ```

- **participants**
  - Liste à jour des participants dans la salle.
  - Exemple :
    ```json
    {
      "type": "participants",
      "data": [
        {
          "userID": "682e3af466a5b24cf18515e9",
          "username": "Hafiz",
          "role": "participant",
          "audioMuted": false,
          "videoOff": false,
          "screenSharing": false
        }
      ]
    }
    ```

- **chat**
  - Message de chat envoyé dans la salle.
  - Exemple :
    ```json
    { "type": "chat", "data": { "message": "Hello", "user": "Hafiz" } }
