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


audio-muted** / **audio-unmuted**
  - Indique qu’un utilisateur a coupé ou réactivé son micro.
  - Exemple :
    ```json
    { "type": "audio-muted", "data": { "userID": "682e3af466a5b24cf18515e9" } }
    { "type": "audio-unmuted", "data": { "userID": "682e3af466a5b24cf18515e9" } }
    ```

- **video-off** / **video-on**
  - Indique qu’un utilisateur a coupé ou activé sa caméra.
  - Exemple :
    ```json
    { "type": "video-off", "data": { "userID": "682e3af466a5b24cf18515e9" } }
    { "type": "video-on", "data": { "userID": "682e3af466a5b24cf18515e9" } }
    ```

- **screen-sharing-started** / **screen-sharing-stopped**
  - Indique qu’un utilisateur commence ou arrête le partage d’écran.
  - Exemple :
    ```json
    { "type": "screen-sharing-started", "data": { "userID": "682e3af466a5b24cf18515e9" } }
    { "type": "screen-sharing-stopped", "data": { "userID": "682e3af466a5b24cf18515e9" } }
