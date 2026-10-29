# Documentation technique

Cette documentation décrit le backend Great Firewall, depuis son démarrage jusqu'à la persistance PostgreSQL.

## Parcours recommandé

1. [Architecture](architecture.md) — composants, responsabilités et dépendances autorisées.
2. [Implémentation](implementation.md) — choix Cobra, Viper et GORM.
3. [Authentification et autorisations](authentication.md) — connexion OIDC, groupes et droits réels.
4. [Configuration](configuration.md) — YAML, environnement, flags et priorités.
5. [Migrations et exploitation](operations.md) — ordre de déploiement, santé et arrêt gracieux.
6. [Tests](testing.md) — stratégie, couverture et commandes de vérification.
7. [Décisions d'architecture](decisions.md) — compromis structurants et raisons associées.

## Schémas

| Schéma | Question traitée |
| --- | --- |
| [Architecture des composants](architecture.svg) | Quels composants existent et quelles dépendances utilisent-ils ? |
| [Flux d'une requête](request-flow.svg) | Comment une requête traverse-t-elle le backend et revient-elle au client ? |
| [Cycle de déploiement](lifecycle.svg) | Dans quel ordre exécuter les migrations et démarrer le serveur ? |

Les SVG possèdent un `viewBox`, des titres et descriptions accessibles, et utilisent une grille conçue pour rester lisible dans GitHub, un navigateur ou une documentation générée.

## Commandes essentielles

```sh
gfw migrate up --config /etc/gfw/config.yaml
gfw serve --config /etc/gfw/config.yaml
gfw version
make verify
```

Le contrat HTTP public reste défini dans [`../openapi.yaml`](../openapi.yaml).
