# Authentification et autorisations

L'interface d'administration utilise le flux OIDC Authorization Code. Le serveur
demande les scopes `openid`, `profile`, `email` et `groups`, vérifie l'ID token,
puis crée une session signée valable huit heures dans le cookie HTTP-only
`gfw_session`. En production, le fournisseur d'identité doit donc inclure une
liste de chaînes dans le claim `groups` de l'ID token.

## Groupes OIDC

| Groupe recommandé | Alias acceptés | Rôle interne | Niveau effectif |
| --- | --- | --- | --- |
| `read` | `gfw-read`, `viewer`, `gfw-viewer` | Viewer | lecture |
| `write` | `gfw-write`, `editor`, `gfw-editor` | Editor | lecture et écriture courante |
| `admin` | `gfw-admin` | Admin | administration complète |

Les noms sont insensibles à la casse. Les droits sont cumulatifs : `write`
inclut `read`, et `admin` inclut `write` et `read`. Si plusieurs groupes sont
présents, le niveau le plus élevé gagne. Tout utilisateur correctement
authentifié sans groupe reconnu reçoit actuellement le niveau Viewer en lecture
seule.

## Droits réels

| Fonction | `read` | `write` | `admin` |
| --- | :---: | :---: | :---: |
| Consulter le tableau de bord | oui | oui | oui |
| Consulter les écosystèmes | oui | oui | oui |
| Créer ou modifier un écosystème | non | non | oui |
| Consulter les machines | oui | oui | oui |
| Ajouter ou mettre à jour une machine | non | oui | oui |
| Supprimer une machine | non | oui | oui |
| Consulter les services de confiance | oui | oui | oui |
| Créer, modifier ou supprimer un service de confiance | non | oui | oui |
| Forcer immédiatement la résolution DNS des services de confiance | non | non | oui |
| Lister, créer ou révoquer les clés d'ingestion | non | non | oui |
| Consulter les listes effectives d'adresses IP | oui | oui | oui |

Les routes publiques de consultation du pare-feu ne dépendent pas de ces
groupes. L'endpoint d'ingestion des machines utilise pour sa part une clé Bearer
créée par un administrateur, et non la session OIDC de l'utilisateur.

## Token administrateur pour l'API

`GFW_ADMIN_TOKEN` fournit une identité technique disposant du rôle Admin sur
les routes protégées. Le client transmet le secret avec
`Authorization: Bearer <token>`. Le token doit contenir au moins 32 caractères
en production, reste dans le Secret Kubernetes et n'est jamais renvoyé par
l'API. Cette authentification est destinée aux automatisations administratives
et ne remplace pas les clés d'ingestion limitées à l'inventaire des machines.

Le frontend masque les actions indisponibles, mais le contrôle faisant foi est
appliqué par le backend sur chaque route. Une requête sans session valide reçoit
`401 authentication_required`; une session valide dont le rôle est insuffisant
reçoit `403 insufficient_role`. Les mutations effectuées avec le cookie de
session sont aussi protégées par un contrôle de même origine.

## Vérification

Le mapping des groupes et la hiérarchie des rôles sont couverts par les tests du
package `internal/auth`. La suite complète se lance avec `make verify`; les tests
backend seuls avec `make test` (la cible exécute également les tests frontend).
L'environnement Docker Compose démarre un véritable Keycloak avec un realm
importé, un client confidentiel et trois utilisateurs Viewer, Editor et Admin.
Il valide donc le même échange Authorization Code, le claim `groups`, la session
signée et les contrôles RBAC que la production. Le mode `--dev-auth` reste
disponible uniquement pour certains tests backend isolés ; il n'est plus utilisé
par l'environnement de développement complet.
