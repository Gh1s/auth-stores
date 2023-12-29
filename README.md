# User stores

Ce projet contient les user stores utilisés par l'Identity Provider, pour authentifier et récupérer les informations des utilisateurs.

## 📁 Structure du projet

Le projet est structuré en multi modules Go, avec un module principal et des [sous modules](https://github.com/go-modules-by-example/submodules) par gRPC.

* `go.mod` : Le fichier du [module Go](https://golang.org/ref/mod) principal.
* `go.sum` : Les dépendances directes et indirectes du module principal.

### `/grpc`

Les implémentations des gRPC `ldap` et `accounts`.

### `/tools`

Les outils partagés avec les autres modules.

### `/users`

Le schéma `protobuf` et le code généré des gRPC.

### `/docker`

Les fichiers de génération des images Docker.

### `/charts`

Les charts Helm.

## 🖥️ gRPC

Les user stores contiennent les informations d'identification des utilisateurs, et sont également résponsables de l'authentification des credentials utilisateurs.
Ce sont des gRPC qui sont consommés l'Identity Provider.

Pour les agents de la CSB, ces informations sont stockées dans l'annuaire LDAP.

Pour les clients de la CSB, ces informations sont stockées dans le gestionnaire de compte CSB.

Pour plus d'informations, consultez le [README](users/README.md) dédié.

### LDAP gRPC

Le gRPC LDAP valide les credentials d'un utilisateur avec l'annuaire LDAP.

Consultez son [README](grpc/ldap/README.md) pour plus d'informations.

### CSB accounts gRPC

Le gRPC des comptes CSB valide les credentials d'un utilisateur avec le gestionnaire de compte utilisateur de la CSB.

Consultez son [README](grpc/accounts/README.md) pour plus d'informations.

## 🧰 Tooling

### Golang

Le code source du projet est écris en [Go](https://golang.org).

Pour le compiler il vous faudra la [CLI Go](https://golang.org/dl).

### IDE

Comme IDE vous pouvez utiliser [VS Code](https://code.visualstudio.com) avec l'extension [Go](https://marketplace.visualstudio.com/items?itemName=golang.Go), ou [IntelliJ IDEA](https://www.jetbrains.com/idea) avec le plugin [Go](https://plugins.jetbrains.com/plugin/9568-go).

### gRPC

Les gRPC utilisent [Protocol Buffers](https://developers.google.com/protocol-buffers) comme protocole de communication et de sérialisation.

Pour générer le code à partir des schémas, il vous faudra la CLI [protoc](https://developers.google.com/protocol-buffers/docs/downloads).

Pour tester les gRPC, la [CLI gRPCurl](https://github.com/fullstorydev/grpcurl) est très utile.

Si vous utilisez VS Code, l'extension [vscode-proto3](https://marketplace.visualstudio.com/items?itemName=zxh404.vscode-proto3) apporte du support pour l'écriture des schéma protobuf 3.

## ⚙️ Configuration

La configuration est gérée avec [Viper](https://github.com/spf13/viper).
Elle est lue depuis plusieurs sources :

* Variables d'environnement : Les variables d'environnement sont lues à chaque accès la configuration et ne sont pas sensibles à la casse. Le caractère `_` sert de séparateur. Par exemple, pour définir la valeur pour les propriétés `logLevel` et `listen.protocol`, il faut définir les variables d'environnement `LOGLEVEL` et `LISTEN_PROTOCOL`.
* Fichiers `.yaml` : Les fichiers `.yaml` sont lus par le gestionnaire de configuration dans l'ordre dans lequel ils ont été passés. Si des memes valeurs sont définies dans plusieurs fichiers, celles lues depuis le dernier fichier sont prises en compte.

### Environnement

Le nom de l'environnement représente l'environnement dans lequel s'exécute l'application, et doit être une chaîne de caractères alphanumériques.

Exemples :

* `development`
* `staging`
* `production`

ℹ️ L'environnement par défaut est `development`.

## 🔑 Key materials

### TLS

Les terminaisons TLS des gRPCs sont assurées par un certificat x509.

#### Développement

En développement, nous utilisons un certificat autosigné.

Pour importer le certificat dans le store « Autorités de certification racines de confiance », depuis un invite de commande `Powershell` en mode admin :

```powershell
Import-Certificate -FilePath .\certs\tls.crt -CertStoreLocation cert:\CurrentUser\Root
```

Si le certificat n'est plus présent dans les sources, vous pouvez en générer un nouveau.

Pour générer la clé privée :

```bash
openssl genrsa -out certs/tls.key 4096
```

Pour générer le certificat : 

```bash
openssl req -new -x509 -sha256 -key certs/tls.key -out certs/tls.crt -days 9999 -subj "/CN=localhost/C=NC/L=Nouméa/OU=APP-DEV" -addext "subjectAltName = DNS:localhost"
```

#### Production/Pré-production/Recette/Qualification

TODO: 🚧 Describe self signed certificate issuance with the Kubernetes cert-manager plugin.

## 🏭 Build

Les gRPC sont conteneurisés dans des images Docker.

---

Pour build les images, depuis le dossier `/docker` :

```bash
./build.sh
```

* gRPC LAP :
    * `gcr.io/csb-anthos/auth/stores/ldap:latest`
    * `gcr.io/csb-anthos/auth/stores/ldap:v{x.x.x}`
* gRPC accounts :
    * `gcr.io/csb-anthos/auth/stores/accounts:latest`
    * `gcr.io/csb-anthos/auth/stores/accounts:v{x.x.x}`

> ℹ️ `{x.x.x}` ou `v{x.x.x}` correspond à la version du service généré.

---

Pour build les images avec le tag `stable`, depuis le dossier `/docker` :

```bash
./build.sh -s
```

* gRPC LAP :
    * `gcr.io/csb-anthos/auth/stores/ldap:latest`
    * `gcr.io/csb-anthos/auth/stores/ldap:stable`
    * `gcr.io/csb-anthos/auth/stores/ldap:v{x.x.x}`
* gRPC accounts :
    * `gcr.io/csb-anthos/auth/stores/accounts:latest`
    * `gcr.io/csb-anthos/auth/stores/accounts:stable`
    * `gcr.io/csb-anthos/auth/stores/accounts:v{x.x.x}`

> ℹ️ `{x.x.x}` ou `v{x.x.x}` correspond à la version du service généré.

---

Pour les build les images et les push sur le registry Docker, depuis le dossier `/docker` :

```
./build.sh -p
```

* gRPC LAP :
    * `gcr.io/csb-anthos/auth/stores/ldap:latest`
    * `gcr.io/csb-anthos/auth/stores/ldap:v{x.x.x}`
* gRPC accounts :
    * `gcr.io/csb-anthos/auth/stores/accounts:latest`
    * `gcr.io/csb-anthos/auth/stores/accounts:v{x.x.x}`

> ℹ️ `{x.x.x}` ou `v{x.x.x}` correspond à la version du service généré.

---

Pour les build les images avec le tag `stable` et les push sur le registry Docker, depuis le dossier `/docker` :

```
./build.sh -s -p
```

* gRPC LAP :
    * `gcr.io/csb-anthos/auth/stores/ldap:latest`
    * `gcr.io/csb-anthos/auth/stores/ldap:stable`
    * `gcr.io/csb-anthos/auth/stores/ldap:v{x.x.x}`
* gRPC accounts :
    * `gcr.io/csb-anthos/auth/stores/accounts:latest`
    * `gcr.io/csb-anthos/auth/stores/accounts:stable`
    * `gcr.io/csb-anthos/auth/stores/accounts:v{x.x.x}`

> ℹ️ `{x.x.x}` ou `v{x.x.x}` correspond à la version du service généré.

