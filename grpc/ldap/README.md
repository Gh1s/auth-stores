# LDAP gRPC

Le gRPC LDAP héberge le serveur gRPC pour authentifier les utilisateurs avec l'annuaire LDAP.

## ⚡️ Quickstart

Pour restaurer les modules Go sur votre poste :

```bash
go mod download
```

Lancer le gRPC sous Windows :

```bat
set LDAP_USERNAME={username}
set LDAP_PASSWORD={password}
go run main.go
```

Lancer le gRPC sous Linux :

```bash
export LDAP_USERNAME={username}
export LDAP_PASSWORD={password}
go run main.go
```

## ⚙️ Configuration

Les fichiers de configuration sont les suivants :

* `/grpc/ldap/config.yaml` : Configure les valeurs par défaut.
* `/grpc/ldap/config.development.yaml` : Configure les valeurs pour l'environnement `development`.
* `/grpc/ldap/config.test.yaml` : Configure les valeurs pour l'environnement `test`.

### LDAP

La connexion avec l'annuaire se fait avec un compte de service.
Les informations d'identification de ce compte sont définies par les clés `ldap.username` et `ldap.password`.

🚨 Ces informations doivent toujours être définies par des variables d'environnement, pour ne pas prendre le risque de les commit dans Git si elles sont définies dans les fichiers de config.

| Clé             | Variable d'environnement  |
| --------------- | ------------------------- |
| `ldap.username` | `LDAP_USERNAME` |
| `ldap.password` | `LDAP_PASSWORD` |

## 🧪 Tests

Il est possible de tester les endpoints du gRPC avec [gRPCurl](https://github.com/fullstorydev/grpcurl).

### Authentification

Pour tester l'endpoint d'authentification avec bash :

```bash
grpcurl -d "{\"Username\":\"$username\",\"Password\":\"$password\"}" -import-path ../../users -proto users.proto localhost:5500 auth.User.Authenticate
```

> `$username` et `$password` sont des variables d'environnement représentant les crendentials du compte que vous voulez authentifier.<br />
> ⚠️ Ne pas confondre avec les informations du compte de service définies plus haut.

### Récupérer des claims

Pour tester l'endpoint de récupération des claims avec bash :

```bash
grpcurl -d "{\"Identifier\":\"$identifier\",\"IdentifierType\":$identifier_type,\"Claims\":[\"sub\",\"given_name\",\"family_name\",\"email\",\"phone_number\"]}" -import-path ../../users -proto users.proto localhost:5500 auth.User.FindClaims
```

> `$identifier` est une variable d'environnement qui représente l'identifiant de l'utilisateur dont vous voulez récupérer les claims.<br />
> `identifier_type` est une variable d'environnement qui représente le type d'identifiant utilisé. 0 = objectGUID, 1 = sAMAccountName

### Rechercher des claims

Pour tester l'endpoint de recherche des claims avec bash :

```bash
grpcurl -d "{\"Search\":\"$search\",\"Claims\":[\"sub\",\"given_name\",\"family_name\",\"email\",\"phone_number\"]}" -import-path ../../users -proto users.proto localhost:5500 auth.User.SearchClaims
```

> `$search` est une variable d'environnement qui représente la recherche de claims à effectuer.<br />
