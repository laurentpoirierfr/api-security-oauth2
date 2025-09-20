# README

## Présentation du composant

Ce projet est un composant de sécurité API basé sur OAuth2, développé en Go. Il permet de protéger les endpoints d'une API en vérifiant et en validant les jetons d'accès OAuth2. Le composant s'intègre facilement dans des applications Go et offre une gestion centralisée de l'authentification et des autorisations.

### Fonctionnalités principales

- Vérification des jetons OAuth2
- Protection des routes API
- Configuration flexible des scopes et des permissions
- Intégration avec différents fournisseurs OAuth2

### Configuration

Personnalisez les paramètres de sécurité selon vos besoins dans le fichier de configuration ou via des variables d'environnement.

``ỳaml
application:
  name: api-security-oauth2
  description: API Security with OAuth2 Example
  version: 1.0.0

server:
  port: 8081
  default_target: http://localhost:3000
  timeout: 10
  oauth2:
    client_id: "backend"
    client_secret: "mysecret"
    redirect_url: "http://localhost:8081/callback"
    endpoints:
      auth_url: "http://localhost:8080/realms/demo/protocol/openid-connect/auth"
      token_url: "http://localhost:8080/realms/demo/protocol/openid-connect/token"
      tokeninfo_url: "http://localhost:8080/realms/demo/protocol/openid-connect/tokeninfo"
      userinfo_url: "http://localhost:8080/realms/demo/protocol/openid-connect/userinfo"

routes:
  - path: "/api/opensource"
    target: "http://localhost:3000/api/opensource"
    teams:
      - name: "opensourceguild"
        description: "OpenSource Guild"
      - name: "opensource-contributors"
        description: "OpenSource Contributors"
      - name: "backend"
        description: "Default backend access" 

  - path: "/api/platform"
    target: "http://localhost:3000/api/platform"
    teams:
      - name: "tm"
        description: "Platform / System"
      - name: "teapot"
        description: "Platform / Cloud API"
      - name: "platform-ops"
        description: "Platform Operations"

  - path: "/api/admin"
    target: "http://localhost:3000/api/admin"
    teams:
      - name: "admin-team"
        description: "Administration Team"
      - name: "security"
        description: "Security Team"

  - path: "/api/public"
    target: "http://localhost:3000/api/public"
    teams: [] # Aucune restriction d'accès
```
### Contribution

Les contributions sont les bienvenues ! Veuillez ouvrir une issue ou soumettre une pull request.

### Licence

Ce projet est sous licence MIT.
