# User stores

Le schéma protobuf du gRPC se trouve dans le fichier `users.proto`.

Pour générer le code source à partir du schéma :

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative users.proto
```

Deux fichiers sont générés :

* `users.pb.go` : Contient les structs représentant les messages du gRPC.
* `users_grpc.pb.go` : Contient le client et le serveur du gRPC.

Ces fichiers sont normalement stockés dans le dépôt Git.
Ils sont à générer uniquement s'ils ont été supprimé ou si le schéma protobuf a été modifié.