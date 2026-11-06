# AGENT.md

## 1. Purpose

You are a software development agent responsible for designing, implementing, testing, refactoring, and maintaining this Go project.

All code changes MUST follow these principles:

* **Test-Driven Development (TDD)**
* **SOLID principles**
* **Hexagonal Architecture (Ports and Adapters)**
* Clear separation of responsibilities
* Low coupling
* High cohesion
* Maintainability
* Testability
* Modern Go idioms and best practices
* Explicit and controlled dependencies

Code quality and architectural integrity take priority over implementation speed.

---

## 2. Core Architectural Principles

The project follows **Hexagonal Architecture**, also known as **Ports and Adapters Architecture**.

The architecture must protect the application core from external technical concerns.

The core application must not depend directly on:

* Web frameworks (e.g., Gin, Fiber, Echo, Chi)
* Databases or ORMs (e.g., GORM, sqlx, `database/sql` drivers directly in domain)
* External HTTP APIs
* File systems
* Message brokers (e.g., RabbitMQ, Kafka, NATS)
* CLI frameworks (e.g., Cobra, Urfave/cli)
* Specific infrastructure libraries

External technologies must be connected through well-defined ports (interfaces) and adapters.

---

## 3. Recommended Project Structure

Use a structure compliant with standard Go project layouts:

```text
internal/
├── domain/
│   ├── entities/
│   ├── value_objects/
│   ├── services/
│   ├── errors/
│   └── repositories/
│
├── application/
│   ├── commands/
│   ├── queries/
│   ├── usecases/
│   └── dto/
│
├── ports/
│   ├── inbound/
│   └── outbound/
│
└── adapters/
    ├── inbound/
    │   ├── http/
    │   ├── cli/
    │   └── ...
    │
    └── outbound/
        ├── persistence/
        ├── external_services/
        └── ...

```

Tests should follow the same architectural boundaries:

```text
tests/
├── unit/
│   ├── domain/
│   └── application/
│
├── integration/
│   └── adapters/
│
└── e2e/

```

The exact directory structure may be adapted to the project, but the architectural boundaries MUST remain explicit.

### Current GRZ package mapping

The current implementation keeps those boundaries in these packages:

```text
internal/domain/kb                          # KB entity and domain rules (tags, graph, wiki links)
internal/domain/wikilink                    # [[cible]] grammar and extraction (pure, no HTML/storage)
internal/domain/journal                     # journal entry (free markdown, dated — Logseq)
internal/domain/media                       # pasted-image rules (formats, size, content-addressed names)
internal/kbfile                             # canonical KB file format codec (Serialize/Parse, shared by server repo and local library)
internal/application/usecases                # KBService, JournalService and MediaService orchestration
internal/application/dto                     # application input/output models
internal/application/ports/{inbound,outbound} # application contracts
internal/adapters/outbound/persistence       # Markdown + media file repositories
web                                          # inbound HTTP adapters (kb_handler, journal_handler, media_handler, kb_local_handler)
html/pages                                   # HTML presentation adapter (incl. image_paste.go / local_kb.go scripts)
main.go                                      # dependency composition root
```

The journal is a **distinct domain** from KB: one entry per day, free
markdown content, no ID/tags (the date is the identity, `YYYY-MM-DD`).
Journal files live in the same `data/kb/` directory under the naming
convention `journal-<date>.md`, owned by `MDJournalRepository`; the KB
repository ignores them (index/journal exclusion, no migration). Journal
routes: `GET /journal`, `GET /journal/new`, `POST /journal`,
`GET /journal/{date}`, `GET /journal/{date}/edit`, `POST /journal/edit`,
`POST /journal/delete`.

Persistence formats and filesystem access belong in adapters; domain packages
must not serialize entities or calculate presentation-specific layout. HTTP
handlers depend on the inbound application port, never directly on a repository.

### Format des fichiers KB et contenu utilisateur

Un fichier KB délimite ses champs par des sections propres du fichier :
`## Erreur`, `## Notes`, `## Résolution` (plus l'en-tête `# ID` et le front
matter `tags:`). Tout contenu utilisateur qui commence une ligne par « ## »
(qu'une heading saisie dans un champ) **éclatait** la section propriétaire
et faisait perdre à la relecture tout ce qui la suivait. L'adapter
(`md_repository.go`) échappe maintenant ces lignes à l'écriture
(`\## …` — échappement de heading markdown valide) via `escapeUserHeadings`
et les restitue à la lecture via `unescapeUserHeadings` : le round-trip
préserve exactement le contenu saisi. Les liens wiki/images restent
inchangés ; le journal (markdown libre, sans délimiteur) n'est pas concerné.

### Liens wiki `[[...]]` (Logseq)

Les sections markdown `Notes` et `Résolution` acceptent des liens wiki
`[[ID]]` (syntaxe stricte : la cible doit être un identifiant KB valide, pas
d'alias). `Erreur` reste une trace brute, jamais liée.

- `internal/domain/wikilink` : grammaire et extraction pures (`Target`,
  `ValidTarget`, `Parse`). Les blocs de code fencés, le code inline, les
  liens échappés (`\[[`) et les cibles invalides sont ignorés ; un lien ne
  traverse jamais une ligne.
- `internal/domain/kb/links.go` : `LinkTargets`, `Backlinks`, `Neighbors`,
  `LinkEdges`. Les **rétroliens sont orientés** : A cite B ⇒ B affiche A,
  même sans réciprocité. `Neighbors` réunit tag partagé et **cibles
  sortantes** `[[...]]` : la section « KB liées » les liste, avec un badge
  « lien » pour celles qui ne partagent aucun tag. Les liens entrants
  restent dans « Références liées », jamais dupliqués.
- Rendu (`html/pages/wikilink.go`) : extension goldmark qui produit
  directement un nœud `Link` (priorité 150, avant le `LinkParser`). Les IDs
  connus transitent par le `parser.Context` goldmark, ce qui préserve le
  pool d'instances `Markdown` et sa sûreté en concurrence. Une cible absente
  devient un **lien mort** (`a.wikilink-missing`) ouvrant le formulaire de
  création prérempli (`GET /kb/new?id=...`).
- Graphe global (`html/pages/kb_graph.go`) : `mergedEdges` fusionne arêtes de
  tags et arêtes de liens explicites en une paire unique ; une arête liée
  porte `kg-edge kg-edge-link` (accentuée) et un tooltip qui mentionne le
  lien et/ou les tags partagés. La fiche détail n'embarque plus de
  mini-graphe : la section « KB liées » et le graphe global suffisent.
- Les liens wiki du journal restent hors périmètre (rétroliens non
  calculés) ; ils peuvent s'afficher comme liens simples sans résolution.

### Drawer latéral + métadonnées dérivées (présentation console)

L'accueil suit le canon « console » (Linear/Raycast docs) : métadonnées en
texte nu (`KB… · SAP MM · 2026-11-04`, zéro pilule), fiche adaptative,
vue tableau dense optionnelle et un **drawer slide-over** pour le détail.

- **Métadonnées dérivées** (`html/pages/meta.go` + tests) : `Modified`
  (`kb.KB`, alimenté par `os.Stat` dans `md_repository.go` — rien n'est
  sérialisé dans le .md, zéro ⇔ date masquée) ; `moduleSAP(tags)` (tag
  court alphabétique 2-4 lettres → « SAP MM », tag « sap »/projet ignoré) ;
  `extractTCodes(text)` (RE2 sans lookahead → frontières vérifiées par
  `tCodeAtBoundary`, tolère le préfixe `/n`, déduplique, borne à 3) ;
  `kBTcodes(k)` extrait des sections Notes + Résolution uniquement (la
  trace `Erreur` reste brute, jamais liée).
- **Vue tableau dense** (`html/pages/kb_table.go`) : cachée par défaut,
  bascule client via `data-view-target` + boutons/icônes (aria-pressed,
  `localStorage grz-view`, script `viewToggleScript`). Les deux vues
  vivent dans des conteneurs `data-fuzzy-view` (`id="kb-cards"` grid,
  `id="kb-table"` table, `hidden`) : le moteur fuzzy
  (`home_search.go`, `inVisibleView` via `closest`) ne compte chaque KB
  qu'une seule fois, sur la vue visible. Une ligne = `h-9` + hover, ID
  mono cliquable, colonnes ID/Module/Titre/Tag/T-codes/Date
  (tabular-nums).
- **Drawer** (routes, shell et fragment) : `GET /kb/panel?id=` sert le
  **fragment** `PageKBDetailPanel` (`html/pages/kb_panel.go`,
  `id="kb-panel"`, jamais layout/nav/doctype) réutilisant les builders de
  la page détail (addSection/monoBody/mdBody/rétroliens/KB liées — fiche
  adaptative). Le handler `web.KBHandler.Panel` réutilise
  `h.app.Detail` : `ValidateID` invalide → 400, inconnu → 404. Le shell
  (`html/pages/kb_drawer.go`, injecté par `PageAccueil` uniquement quand
  il y a des cartes) porte le dialogue accessible (`role="dialog"`,
  `aria-modal`, `aria-label="Détail de la KB"`), l'overlay
  `#kb-drawer-overlay`, le bouton `#kb-drawer-close` et le corps
  `#kb-drawer-body` ; fermé par CSS pur (translateX/opacity), ouvert par
  la classe `.open`. Le script `grz-drawer` fetch le fragment dans
  `#kb-drawer-body`, ferme (overlay/bouton/Échap) et restitue le focus ;
  en cas d'échec réseau il retombe sur la navigation native du `href`
  (`/kb/detail?id=…`). Les ancres des cartes (`kbCard`, lien étendu) et
  des lignes du tableau portent `data-kb-drawer="<ID>"` + href natif
  conservé — amélioration progressive, sans JS le détail fonctionne.
- Bandeau T-codes du fragment : chips mono `/nME22N` (notation logique
  de session SAP), surface `g-code` + filet `g-border`. (Le bloc « Copier
  la solution » — bouton, textarea presse-papiers et script `grz-copy` —
  a été retiré : sans valeur ajoutée, pressage sans effet. Il n'existe
  pas d'autre copie que celle des blocs de code markdown, `md.go`.)

### Création libre de KB (`/kb/new/quick`)

Deux formulaires de création coexistent, sous-tendus par le même use case
`KBService.Create` :

- formulaire complet : `GET /kb/new` → `POST /kb` (identifiant, Erreur,
  Notes, Résolution, tags) ;
- **mode libre** : `GET /kb/new/quick` → `POST /kb/quick` (identifiant saisi
  à la main, une seule note markdown libre avec images collables, tags
  optionnels). Les sections Erreur/Résolution sont absentes du formulaire
  (vide à la création) et restent complètables via l'édition. La fiche
  détail est **adaptative** : seules les sections non vides se rendent
  (`PageKBDetail`), donc une KB libre garde sa forme d'origine tant qu'elle
  n'est pas complétée ; il n'existe plus de mention « — vide — » de secours
 pour les fiches KB.

- `html/pages/kb_create.go` : `kbQuickFormFields` est une liste **sans
  dérive possible** des sections Erreur/Résolution ; `kbEditFieldsFor`
  dérive de `kbFormFields` (toute option ajoutée à la création se propage
  à l'édition — ce couplage a corrigé un bug d'images collables).
- `web/kb_handler.go` : `ShowNewQuickForm` / `CreateQuick` réutilisent
  `inputFromRequest` et les messages de validation ; les erreurs
  re-rendent `PageKBCreateQuickWith` (jamais le formulaire complet).
- Accueil : bouton secondaire « Nouvelle KB libre » à côté de « Créer une
  KB » (`html/pages/home.go`).

### Bibliothèque locale (File System Access, mondes étanches)

Chaque utilisateur de l'app hébergée (le serveur central) peut enregistrer
des articles en **vrais fichiers `.md`** dans un dossier **de sa propre
machine**, consultables uniquement par lui. Monde **étanche** : une fiche
locale n'apparaît jamais dans les listes, le graphe ou les rétroliens du
serveur ; un fichier local peut en revanche câbler un wikilink vers une ID
partagée (lien vivant vers la fiche serveur, lien mort sinon). Aucun
compte utilisateur : l'isolation vient du poste/navigateur, pas du serveur.

- **Un seul codec** : le format fichier (`# ID`, front matter `tags:`,
  sections `## Erreur` / `## Notes` / `## Résolution`, échappement
  `\##`) vit dans `internal/kbfile` (`Serialize`, `Parse`,
  `escapeUserHeadings`). Le dépôt serveur (`md_repository.go`) comme les
  cas d'usage du local consomment ce package : aucune réimplémentation
  client ne peut dériver du format canonique.
- **Cas d'usage sans état** (`KBService` + port `inbound.KBApplication`) :
  `SerializeLocal` (validation domaine → contenu .md canonique),
  `ParseLocal` (fichier → champs éclatés), `LocalKnownIDs` (IDs partagés
  pour le rendu des wikilinks). Le serveur ne persiste **jamais** le
  périmètre local.
- **Adaptateur HTTP** `web/kb_local_handler.go` : `POST /kb/local/serialize`
  (form → `text/markdown`, 400 si validation refuse), `POST /kb/local/parse`
  (corps brut → JSON : champs éclatés **plus les dérivations serveur**
  `module` et `tcodes`, calculées par `pages.ModuleSAPFor` /
  `pages.TCodesFor` — les mêmes règles Go que les cartes partagées, aucune
  réimplémentation client), `POST /kb/local/render` (contrat JSON
  `{notes, resolution}` → `{notesHTML, resolutionHTML}` via la pipeline
  goldmark partagée, IDs partagés fournis par `LocalKnownIDs` ;
  `pages.RenderMarkdownHTML` — l'Erreur n'emprunte jamais ce pipeline :
  trace brute rendue en mono côté client).
- **Script client** `html/pages/local_kb.go` (`grz-local`, injecté dans la
  barre locale de l'accueil et sous les formulaires de création — jamais
  dans le journal ni à l'édition partagée) : connexion du dossier
  (`showDirectoryPicker`, handle `readwrite` persisté dans IndexedDB
  `grz-local-kb`, `queryPermission` à chaque session — reconnexion
  explicite par clic, jamais silencieuse), index des `<id>.md` (hors
  `journal-<date>.md`), **injection DANS la grille partagée** `#kb-cards`
  (rendue toujours, même vide) et dans la vue tableau `#kb-table-rows` :
  les articles locaux s'affichent AVEC les KB partagées, comptés dans
  « n/total » du moteur fuzzy (re-scan via `window.grzFuzzyRefresh`) ;
  **carte locale = clone exact de `kbCard`** (heading `ID · module · date`
  avec module/T-codes dérivés par le serveur, titre = 1re ligne de
  l'Erreur, extrait borné à 140, tag principal + premier T-code +
  « Consulter → » dans le pied) et **ligne tableau = clone de
  `kbTableRow`** ; purge de l'injection précédente via `data-local-entry`,
  bascule de l'état vide via `#kb-empty` ; un filtre tag actif
  (`data-tag-filtered` posé par `PageAccueil`) suspend l'injection (les
  tags locaux ne sont pas filtrables) ; **fiche locale = clone de
  `PageKBDetailPanel`** dans le tiroir (bandeau `ID · module · date`,
  chips T-codes `/n`, Erreur trace brute mono, Notes/Résolution rendues,
  fiche adaptative, actions **Modifier** (préremplissage via
  `/kb/local/parse` + `?local=1&id=…`) et **Supprimer**
  (`dir.removeEntry`) **vit uniquement dans la fiche**, comme côté
  serveur) ;
  **ordre DOM garanti** : le shell du tiroir est rendu AVANT la barre
  (`drawerShell()` puis `PageLocalBar()` dans `PageAccueil`) — le script
  s'exécute au parsing et ouvre les fiches dans le tiroir ; de plus la
  capture des éléments du tiroir est PARESSEUSE (`drawerRoots()`
  re-résolus à chaque usage, régression d'ordre déjà vue) ;
  les scripts ne s'exécutent pas via innerHTML, le câblage des boutons
  copier vit dans `bindCodeCopy` ; interception en capture du submit
  (`preventDefault` + `stopPropagation` devant htmx) quand la case
  `#kb-local-save` est cochée, écriture via `createWritable`.
- **Amélioration progressive** : la barre locale est rendue `hidden`
  côté serveur ; le script ne la révèle que si `window.showDirectoryPicker`
  existe (Chrome/Edge). Sans JavaScript ou avec Firefox/Safari,
  l'application est inchangée.
- **Limites v1 assumées** : les images collées restent servies par le
  serveur (`/media/<hash>`) — hors ligne, elles tombent ; la ligne
  tableau locale affiche « — » quand module/T-code ne sont pas
  dérivables (même comportement qu'une carte serveur sans module) ;
  l'unicité d'ID tient par monde (étanche), pas à travers les deux.

### Recherche fuzzy partagée (`data-fuzzy`)

Le filtre instantané est 100 % client (`html/pages/home_search.go`) : le
serveur rend déjà tout ce qui est filtrable, chaque élément porte une clé
`data-search` minuscule, le scoreur (`fnFuzzyScore`) matche une
sous-séquence (bonus lettres consécutives et débuts de mot) et le
raccourci `/` focuse le champ. Le markup est **déclaratif** : l'input
porte `data-fuzzy` et déclare ses acolytes sur lui-même

- `data-fuzzy-count` : id du compteur « n/total » ;
- `data-fuzzy-no-results` : id du bloc « aucun résultat » (caché) ;
- `data-fuzzy-empty` : libellé du message vide ;
- `data-fuzzy-order` : (optionnel) re-trier par pertinence via `style.order`
  (grille d'accueil) — sans lui la liste conserve son ordre (journal).

Utilisé par l'accueil (cartes KB, `searchText` = ID + 1ʳᵉ ligne de l'erreur
+ tags) et le journal (`journalSearchKey` = date brute + date FR + 3
premières lignes non vides de la note). Rendu seulement quand il y a
quelque chose à filtrer (progressive enhancement).

### Images collées

Les zones markdown `notes`, `resolution` (KB) et `content` (journal)
acceptent le collage et le glisser-déposer d'images. `Erreur` reste une
trace brute, sans image.

- `internal/domain/media` : règles pures (`NameFor`, `ValidName`,
  `ContentType`, `MaxSize`). Formats acceptés : PNG, JPEG, GIF, WebP —
  **SVG refusé** (script embarqué). Taille maximale 5 Mio. Nom =
  `<sha256>.<ext>`, adressé par le contenu : deux images identiques
  partagent le même fichier.
- `web/media_handler.go` : `POST /media` reçoit le binaire brut et
  recalcule le type avec `http.DetectContentType` (l'en-tête client n'est
  jamais une preuve). `GET /media/{name}` sert le fichier avec
  `X-Content-Type-Options: nosniff` et un cache immuable (nom = hash).
  Erreurs : `415` format refusé, `413` taille dépassée, `400` corps vide
  ou nom invalide, `404` nom inconnu.
- `internal/adapters/outbound/persistence/local_media_repository.go` :
  stockage dans `data/media/` (couvert par `.gitignore` via `data/`),
  écriture atomique (fichier temporaire puis `rename`), déduplication par
  nom.
- `html/pages/image_paste.go` : script inclus **uniquement** dans les pages
  de formulaire. Il insère un marqueur pendant l'envoi, le remplace par
  `![image](/media/<hash>.<ext>)`, désactive les boutons pendant l'upload
  et retire le marqueur en cas d'échec. Le surlignage de dépôt utilise un
  `outline` inline (`var(--g-sep)`) : aucune classe Tailwind nouvelle, donc
  pas de régénération CSS.
- Rendu : goldmark rend déjà `![alt](url)` ; la règle inline `.md img`
  (`layout.go`) borne la largeur à la fiche. Les images orphelines ne sont
  pas supprimées quand une KB disparaît (hors périmètre).

---

## 4. Dependency Rule

Dependencies MUST always point toward the application core.

```text
Adapters
    ↓
Ports
    ↓
Application
    ↓
Domain

```

The domain must remain independent of infrastructure.

### Forbidden

```go
// internal/domain/user.go

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

```

### Preferred

Go interfaces define the ports directly in the application or domain layers:

```go
// internal/application/ports/outbound/user_repository.go
package outbound

import (
    "context"
    "yourproject/internal/domain/entities"
)

type UserRepository interface {
    FindByID(ctx context.Context, id entities.UserID) (*entities.User, error)
    Save(ctx context.Context, user *entities.User) error
}

```

The concrete implementation belongs to an adapter:

```go
// internal/adapters/outbound/persistence/postgres_user_repository.go
package persistence

import (
    "context"
    "database/sql"
    "yourproject/internal/domain/entities"
)

type PostgresUserRepository struct {
    db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
    return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *entities.User) error {
    // Concrete SQL implementation
    return nil
}

```

---

## 5. Domain Layer

The domain contains the core business rules.

It should contain concepts such as:

* Entities (structs with business logic methods)
* Value Objects (immutable structs)
* Domain Services
* Domain Errors
* Domain Events
* Domain-specific interfaces

The domain MUST NOT contain infrastructure concerns.

The domain should be:

* Framework-independent
* Database-independent
* Highly testable
* Deterministic whenever possible

Example:

```go
package valueobjects

import "errors"

var ErrCurrencyMismatch = errors.New("currencies do not match")
var ErrNegativeAmount = errors.New("amount cannot be negative")

type Money struct {
    amount   int64 // stored in smallest currency unit (e.g. cents)
    currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    return Money{amount: amount, currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

```

---

## 6. Application Layer

The application layer orchestrates use cases.

A use case should:

1. Receive a command or input DTO.
2. Retrieve required data through ports.
3. Invoke domain behavior.
4. Persist changes through ports.
5. Return an appropriate result or error.

Example:

```go
package usecases

import (
    "context"
    "yourproject/internal/application/dto"
    "yourproject/internal/application/ports/outbound"
    "yourproject/internal/domain/entities"
)

type CreateUserUseCase struct {
    repo outbound.UserRepository
}

func NewCreateUserUseCase(repo outbound.UserRepository) *CreateUserUseCase {
    return &CreateUserUseCase{repo: repo}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, cmd dto.CreateUserCommand) (*entities.User, error) {
    user, err := entities.NewUser(cmd.Name, cmd.Email)
    if err != nil {
        return nil, err
    }

    if err := uc.repo.Save(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

```

A use case MUST NOT:

* Create database connections directly.
* Call external HTTP APIs directly.
* Depend directly on an ORM or SQL driver.
* Depend directly on a web framework or HTTP request/response objects.
* Contain infrastructure-specific logic.

---

## 7. Ports

Ports define contracts between the application core and external systems using Go interfaces.

### 7.1 Inbound Ports

Inbound ports define what the application exposes.

Example:

```go
type CreateUserUseCase interface {
    Execute(ctx context.Context, cmd dto.CreateUserCommand) (*entities.User, error)
}

```

Inbound adapters call inbound ports.

```text
HTTP Controller / REST API
CLI Command
Message Consumer
Scheduled Job (Cron)
       ↓
Inbound Port
       ↓
Use Case

```

---

### 7.2 Outbound Ports

Outbound ports define what the application requires from external systems.

Example:

```go
type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByID(ctx context.Context, id entities.UserID) (*entities.User, error)
}

```

The application defines the contract. The infrastructure implements it.

```text
Application Core
       ↓
UserRepository (Interface)
       ↑
PostgresUserRepository (Adapter Struct)

```

---

## 8. Adapters

Adapters connect the application to the outside world.

### 8.1 Inbound Adapters

Examples:

* HTTP / REST Handlers
* gRPC Controllers
* CLI Handlers
* Message consumers (AMQP/Kafka listeners)
* Scheduled jobs

Their responsibility is to translate external input into application commands/queries, invoke the use case, and serialize output.

Controllers MUST NOT contain business logic.

---

### 8.2 Outbound Adapters

Examples:

* PostgreSQL / SQLite implementations
* Redis cache adapters
* External HTTP API clients
* File system storage
* Event publishers

They implement outbound port interfaces.

```text
Use Case
   ↓
Repository Port Interface
   ↓
Persistence Adapter
   ↓
Database (sql.DB)

```

---

## 9. Test-Driven Development

TDD is mandatory for new behavior.

Follow the cycle:

```text
RED
 ↓
GREEN
 ↓
REFACTOR

```

### RED

Write a test describing the expected behavior before implementing the functionality. The test must fail for the expected reason.

Example:

```go
func TestUserCannotBeCreatedWithInvalidEmail(t *testing.T) {
    _, err := entities.NewUser("John", "invalid-email")
    if !errors.Is(err, entities.ErrInvalidEmail) {
        t.Errorf("expected ErrInvalidEmail, got %v", err)
    }
}

```

### GREEN

Implement the minimum amount of code required to make the test pass. Do not implement speculative functionality.

### REFACTOR

Once the tests are green:

* Remove duplication.
* Improve naming.
* Simplify the implementation.
* Apply SOLID principles.
* Verify architectural boundaries.

Tests must remain green after refactoring.

---

## 10. Testing Strategy

Organize tests using standard Go tools (`go test`).

Preferred test pyramid:

```text
       E2E
      /   \
  Integration
  /         \
  Unit Tests

```

### 10.1 Unit Tests

Unit tests should cover:

* Domain entities & value objects
* Domain services
* Use cases

External dependencies should be replaced with simple **fakes** or mock implementations. Prefer hand-crafted mock structs or `testify/mock` over heavy mocking frameworks.

Use Go table-driven tests (`struct{ name string; ... }`) for test case coverage.

For performance-sensitive code, add benchmarks and compare before/after using
`go test <package> -run '^$' -bench . -benchmem`. Keep optimizations that improve
representative workloads without changing correctness or observable ordering.

---

### 10.2 Integration Tests

Integration tests verify that adapters interact correctly with external systems.

Examples:

```text
Repository Adapter + Real DB / Testcontainers
HTTP Handler + Application Layer (httptest.NewServer)
External Service Adapter + Mock HTTP Server

```

Do not mock away the exact dependency that an integration test aims to verify.

---

### 10.3 End-to-End Tests

E2E tests verify full user journeys across the entire system. Keep E2E tests limited to critical flows.

---

## 11. SOLID Principles

All design decisions should respect SOLID:

* **S — Single Responsibility Principle:** A struct or package should have one clear responsibility. Avoid monolithic handler or service structs.
* **O — Open/Closed Principle:** Extend capabilities by injecting interface implementations rather than modifying existing code blocks.
* **L — Liskov Substitution Principle:** Any struct implementing a port interface must be substitutable without breaking expectations or contract semantics.
* **I — Interface Segregation Principle:** Keep interfaces small and client-focused (e.g., single-method interfaces like `Reader`, `Writer` or specific domain interfaces).
* **D — Dependency Inversion Principle:** High-level application code depends on abstractions (Go interfaces), not on concrete structs or external packages.

---

## 12. Dependency Injection

Dependencies MUST be explicitly injected (usually via constructor functions `New...`), never instantiated directly inside business logic.

Preferred:

```go
type OrderService struct {
    repo    OrderRepository
    payment PaymentGateway
}

func NewOrderService(repo OrderRepository, payment PaymentGateway) *OrderService {
    return &OrderService{
        repo:    repo,
        payment: payment,
    }
}

```

Avoid:

```go
func NewOrderService() *OrderService {
    return &OrderService{
        repo:    persistence.NewPostgresOrderRepo(), // Direct concrete binding
    }
}

```

Concrete dependency resolution and wiring MUST happen in a **Composition Root** (typically in `cmd/main.go`).

---

## 13. Framework Isolation

Frameworks and infrastructure libraries are implementation details.

The core packages (`internal/domain`, `internal/application`) MUST NOT import:

```text
net/http (except context definitions if needed, preferred context package only)
github.com/gin-gonic/gin
github.com/labstack/echo
gorm.io/gorm
github.com/jmoiron/sqlx
github.com/go-redis/redis

```

The goal is to allow framework or infrastructure replacement without altering core domain logic.

---

## 14. Idiomatic Go & Type Safety

* **Strong Typing:** Use custom types (e.g., `type UserID string`) to prevent primitive obsession.
* **Explicit Error Handling:** Check errors explicitly (`if err != nil`). Never ignore errors using `_`.
* **Context Handling:** Pass `ctx context.Context` as the first parameter in use cases, ports, and IO-bound methods.
* **Avoid Empty Interfaces:** Avoid `interface{}` / `any` unless strictly necessary (e.g., dynamic JSON payloads or generic utilities).
* **Zero-Values:** Design structs and types to have usable or safe zero-values whenever possible.

---

## 15. Error Handling

Business errors should be sentinel errors or custom error types defined within the domain package.

Example:

```go
var ErrUserAlreadyExists = errors.New("user already exists")

```

Infrastructure errors must not leak into the core. Wrap technical errors at the adapter boundary using `fmt.Errorf("adapter error: %w", err)`.

---

## 16. Configuration and Secrets

Configuration must remain separate from code.

Never hard-code:

* API keys
* Database credentials
* Tokens
* Environment-specific URLs

Use environment variables or explicit configuration structs loaded at startup (e.g., via `envconfig` or `viper` in `main.go`).

---

## 17. Code Quality

Code must strictly adhere to standard Go formatting and linting conventions:

* Format code using `gofmt` or `goimports`.
* Run static analysis with `go vet` and `golangci-lint`.
* Write clear, concise standard Go doc comments for exported types and functions.
* Keep functions small and cyclomatic complexity low.

`golangci-lint` est configuré par `.golangci.yml` (v2) : les dot-imports
`import . "maragu.dev/gomponents/html"` sont l'idiome officiel de la
bibliothèque, utilisé partout dans le projet — la règle `ST1001` de
staticcheck y est explicitement désactivée. Le binaire est installé depuis
la source (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`,
disponible dans `~/go/bin`) : préférez toujours cette méthode, les binaires
pré-compilés d'une version de Go incompatible échouent à typechecker le
module sans raison apparente.

---

## 18. Avoid Premature Abstraction

Do not create an interface for every single struct unless it serves as an explicit port, has multiple implementations, or is required for testing.

Prefer:

```text
Simple + Testable + Decoupled

```

over over-engineered hierarchies.

---

## 19. Development Workflow

For every new feature or change, follow this process:

### Step 1 — Understand

Identify requirements, domain boundaries, existing tests, and ports involved.

### Step 2 — Identify Responsibilities

Assign responsibilities to Domain, Application, Port, or Adapter.

### Step 3 — Write Tests (TDD)

Write unit/integration tests describing expected behavior first.

### Step 4 — Implement

Implement minimal code to pass tests.

### Step 5 — Run Tests & Linters

```bash
go test -v ./...
golangci-lint run

```

### Step 6 — Refactor

Refactor for clarity, idiomatic Go, and SOLID rules while maintaining green tests.

### Step 7 — Add Integration Tests

Add adapter-level integration tests if external systems were touched.

### Step 8 — Update Documentation

Update documentation and `AGENT.md` if conventions or architecture were modified.

### Step 9 — Final Validation

Ensure all items on the checklist pass.

---

## 20. Modifying Existing Code

1. Read existing implementation and tests.
2. Identify architectural boundaries.
3. Add/modify tests first.
4. Implement change and verify tests pass.
5. Refactor without altering public contracts unannounced.
6. Keep documentation up to date.

---

## 21. Documentation Synchronization

Documentation must reflect code implementation.

Verify before completing a task:

```text
[ ] Code matches documentation
[ ] Architecture documentation reflects actual code layout
[ ] Commands documented in README/AGENT.md work
[ ] AGENT.md is updated with any project rule changes

```

---

## 22. AGENT.md Maintenance

`AGENT.md` is a living document. It MUST be updated whenever project conventions, folder structure, testing practices, or rules evolve.

---

## 23. Forbidden Practices

Do not:

* Put business logic in HTTP handlers or database adapters.
* Access database drivers directly inside use cases.
* Import web or database packages inside `internal/domain`.
* Ignore returned errors (`_ = operation()`).
* Use global mutable variables or unchecked singletons.
* Skip TDD or unit tests.
* Leave `AGENT.md` or architectural documentation outdated.

---

## 24. Definition of Done

A task is complete only when:

```text
[ ] Expected behavior is covered by tests
[ ] TDD workflow was followed
[ ] `go test ./...` passes without errors
[ ] `golangci-lint run` passes cleanly
[ ] Hexagonal architecture boundaries are respected
[ ] Core layers remain framework- and DB-independent
[ ] Dependencies are explicitly injected
[ ] Code follows idiomatic Go standards
[ ] No secrets or credentials are hard-coded
[ ] Project documentation & AGENT.md are fully up to date

```

---

## 25. Final Decision Rules

Prioritize:

1. **Correctness**
2. **Business Rules Integrity**
3. **Testability**
4. **Simplicity & Idiomatic Go**
5. **Decoupling**

---

## 26. Server Instances & Port 8080 (Development Hygiene)

The app binds permanently to `:8080` (`http.ListenAndServe(":8080", mux)` in `main.go`).

Rules for any agent session or human developer testing the server locally:

1. **Before starting** the server, free the port if an old instance holds it:
   `fuser -k 8080/tcp` (or `pkill -f "grz"` for a previous build), and check with
   `ss -tln | grep 8080`.
2. **Never leave an instance running** at the end of a session. Explicitly kill
   it: `fuser -k 8080/tcp; pkill -f "/path/to/grz"` — an orphaned process keeps
   port 8080 occupied and silently breaks the next run (the startup log only
   shows "port 8080 already in use" failure at exit).
3. Prefer building a test binary in a scratch directory (e.g. `/tmp`) and
   running `./grz` there, so the project's `data/kb/` directory stays clean:
   test KBs must not pollute the real storage.
4. **After smoke tests**, verify the port is actually released
   (`ss -tln | grep 8080` returns nothing) before declaring the task complete.

---

## 27. Local CSS Generation (Tailwind, No CDN)

The styles come from a **locally generated** stylesheet, embedded in the
binary via `go:embed` (`web/assets.go`, served at `GET /assets/tailwind.css`).
There is no Tailwind CDN or Node toolchain.

- Source of truth for colors: `html/pages/palette.go` (Kally tokens —
  Dracula-flavoured dark `#282a36` canvas / `#21222c` cards / purple
  `#bd93f9` accents with dark ink on them; near-white `#fcfcfd` canvas /
  white cards in light, accents deepened for AA). Every text/background
  pair is validated at 4.5:1 by `html/pages/contrast_test.go`.
  `tailwind.config.js` mirrors them — tests (`html/pages/palette_test.go`)
  fail if the two drift apart.
- Source of truth for typography: Plus Jakarta Sans (body) and JetBrains
  Mono (IDs `KB…`/`INC…`, SAP T-codes, tabular counters), variable fonts
  self-hosted in `web/assets/fonts/*.woff2` (`go:embed`, served at
  `GET /assets/fonts/{name}`, immutable cache). `@font-face` rules and the
  body font stack live in `layout.go`; `tailwind.config.js` maps
  `font-sans` / `font-mono` to the same families.
- Utility classes live inside Go source strings (`html/**`, `web/**`,
  `main.go`), scanned by the content globs in `tailwind.config.js`.
- **Any change adding or renaming a utility class in Go requires
  regenerating the stylesheet:**

  ```bash
  ./tools/build-css.sh          # one-shot generation (CLI standalone, pinned) — Linux
  ./tools/build-css.sh --watch  # watch mode during development
  ```

  ```powershell
  powershell -File tools/build-css.ps1   # same contract — Windows
  $env:CSS_WATCH = "1"; ...              # watch mode (PowerShell)
  ```

  The script caches the `tailwindcss` standalone CLI (no Node) in
  `$HOME/.cache/grz` and writes the minified output to
  `web/static/tailwind.css`, which is committed with the binary-side embed.
- Page-level styles (body, `.md` markdown sections, SVG graph) remain
  **inline** in `layout.go` / `kb_graph.go`, consuming `var(--g-*)` —
  do not hard-code hex colors in CSS strings; change the values in
  `palette.go` instead.

### Theming & Palette (Phases 3–4)

- Colors are **CSS variables** (`--g-*`), generated in the inline `<style>`
  by `layout.go` (`themeCSSRules`): `:root` = Kally obsidian dark
  (default), `[data-theme="light"]` = slate/white light ladder.
  `palette.go` holds both maps (`paletteDark`, `paletteLight`) — same token
  set, tests fail on drift (`TestThemePalettes`, `TestLayoutThemeVarRules`).
- Tailwind tokens in `tailwind.config.js` reference `var(--g-*)`; the theme
  is switched at runtime via `<html data-theme="light|dark">` set by the
  **anti-FOUC init script** (place it BEFORE the stylesheet link) and the
  navbar toggle (`#theme-toggle`, persisted in localStorage `grz-theme`).
  Default follows `prefers-color-scheme`.
- SVG colors must go through **CSS classes** (`kg-edge`, `kg-dot`,
  `kg-label` set in the SVG stylesheet, colors `var(--g-*)`) — presentation
  attributes cannot use `var()`.
- Markdown code highlighting: goldmark-highlighting + Chroma in **classes
  mode** (`chromahtml.WithClasses(true)`, prefix `chroma-`) — never inline
  colors. Token-family colors live in the inline stylesheet (`.chroma-*`)
  and MUST use the `g-syn-*` tokens (`g-syn-keyword`, `g-syn-string`,
  `g-syn-comment`, `g-syn-error`) or `g-hi` — never `g-cyan`, `g-positive`,
  `g-text` or `g-negative-text`, which do not reach 4.5:1 on `g-code`.
  Extended families (kd/kt/fm/mb/…): update the rules in `layout.go` when a
  new lexer family shows up unstyled.

### Text Colors & Contrast (WCAG AA)

- Every text/background pair used by the UI MUST reach **4.5:1**
  (`html/pages/contrast_test.go`, `TestPaletteContrast`, both themes).
- Use `g-text` for body copy, excerpts, labels, placeholders, tags,
  secondary actions and empty-state mentions. `g-muted` / `g-subdue` are
  decorative only and MUST NOT be used as a text color.
- `g-accent` and `g-negative` are decorative tokens (SVG, borders, halos,
  fills). Any **text** using those hues MUST use `g-accent-text` /
  `g-negative-text`, which are tuned per theme for contrast.
- `g-ink` is reserved for text placed on `g-accent` (primary buttons).
- Read-only form fields reuse `g-widget` + `g-text` + `cursor-not-allowed`;
  editable fields keep `g-block`.
- **Nested surfaces use their own token.** `g-widget` is the card surface;
  never reuse it as the background of an element *inside* a card (that made
  `<code>`/`<pre>` invisible). Code surfaces use `g-code` plus a
  `1px solid var(--g-border)` outline; code text is `g-hi` (`g-text` falls
  under 4.5:1 on `g-code`). `html/pages/contrast_test.go`
  (`TestCodeSurfaceDistinctFromWidget`) guards against the regression.

### Form Errors: htmx Swap vs Native Fallback

htmx does **not** swap the target of a `4xx`/`5xx` response. Therefore:

- Request carries `HX-Request: true` → respond **200 + the message node
  only** so `#reponse-serveur` is replaced and the error is visible.
- Native POST (no JavaScript) → respond the **full pre-filled form page**
  with the validation message and the proper status (`400` invalid,
  `409` duplicate). Values already typed are preserved through
  `pages.FormulaireKB{Valeurs, Message}` and `PageKBCreateWith` /
  `PageKBEditWith`.
- Same rule for **navigation after a POST** (Create, Update, Delete via
  `web.redirectTo`): htmx → `200 + HX-Redirect` (the AJAX request would
  follow a 303 and stay on the current URL); native → `303`.
- The faulty field is flagged with the `.field-error` class (outline
  `var(--g-negative-text)`), removed on the first `input` event. Never
  inject an inline `#dc322f` outline: colors must follow the active theme.
- Labels are bound to controls (`for="field-<nom>"` / matching `id`), and a
  control carries exactly **one** `class` attribute (append modifiers to
  the class string, never emit a second `Class(...)`).

