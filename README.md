# 🎵 FAQears — A Spotify-like Music Streaming Platform

> A distributed, cloud-native music streaming service built on a microservices architecture using **Go**, **gRPC**, **Kafka**, and modern cloud technologies. This is a pet project designed to demonstrate production-grade backend engineering practices.

---

## 📋 Table of Contents

- [🎵 FAQears — A Spotify-like Music Streaming Platform](#-faqears--a-spotify-like-music-streaming-platform)
  - [📋 Table of Contents](#-table-of-contents)
  - [🎯 Project Overview](#-project-overview)
    - [Core Use Cases](#core-use-cases)
  - [👥 Team \& Responsibilities](#-team--responsibilities)
  - [✅ Functional Requirements](#-functional-requirements)
  - [⚙️ Non-Functional Requirements](#️-non-functional-requirements)
  - [🛠 Technology Stack](#-technology-stack)
  - [🏛 System Architecture](#-system-architecture)
    - [Architectural Principles](#architectural-principles)
  - [🧩 Microservices Breakdown](#-microservices-breakdown)
    - [1. `auth-service` — *Owner: Andrew Rudov*](#1-auth-service--owner-andrew-rudov)
    - [2. `user-service` — *Owner: Andrew Rudov*](#2-user-service--owner-andrew-rudov)
    - [3. `catalog-service` — *Owner: Danial Boranbayev*](#3-catalog-service--owner-danial-boranbayev)
    - [4. `streaming-service` — *Owner: Danial Boranbayev*](#4-streaming-service--owner-danial-boranbayev)
    - [5. `playlist-service` — *Owner: Daniyar Abdrakhmanov*](#5-playlist-service--owner-daniyar-abdrakhmanov)
    - [6. `recommendation-service` — *Owner: Daniyar Abdrakhmanov*](#6-recommendation-service--owner-daniyar-abdrakhmanov)
    - [7. `payment-service` — *Owner: Andrew Rudov*](#7-payment-service--owner-andrew-rudov)
    - [8. `generation-service` — *Owner: Daniyar Abdrakhmanov*](#8-generation-service--owner-daniyar-abdrakhmanov)
  - [💾 Data Storage Strategy](#-data-storage-strategy)
  - [🔌 Inter-Service Communication](#-inter-service-communication)
  - [🚪 API Gateway \& Routing](#-api-gateway--routing)
  - [🔐 Authentication \& Authorization](#-authentication--authorization)
  - [📊 Observability \& Monitoring](#-observability--monitoring)
  - [☁️ Deployment \& Infrastructure](#️-deployment--infrastructure)
  - [📁 Repository Structure](#-repository-structure)
  - [🔄 Development Workflow](#-development-workflow)
  - [🗺 Roadmap \& Milestones](#-roadmap--milestones)
  - [📚 Further Reading \& References](#-further-reading--references)
  - [📜 License](#-license)

---

## 🎯 Project Overview

**faqears** is a backend-focused music streaming platform inspired by Spotify. The goal is not to compete with Spotify on UX or catalog, but to build a realistic, scalable distributed system that handles the core domains of a streaming service: user identity, music catalog, audio streaming, playlists, search, and recommendations.

The project is intentionally split into independently deployable microservices, each owned by one team member. Services communicate primarily over **gRPC** for synchronous calls and **Apache Kafka** for asynchronous, event-driven workflows. The system is containerized with **Docker**, orchestrated with **Kubernetes**, and exposed through a single **API Gateway**.

### Core Use Cases

- A user can register, log in, and manage their profile.
- A user can browse a catalog of artists, albums, and tracks.
- A user can stream audio with low latency and adaptive quality.
- A user can create, edit, and share playlists.
- A user can search across tracks, albums, and artists with fuzzy matching.
- A user receives personalized recommendations based on listening history.
- The system records every play event and exposes analytics.
- A new user can optionally upgrade to a **premium tier** by paying with crypto via **NowPayments** during or after registration; payment status arrives asynchronously via signed webhooks.
- A premium user can **generate their own songs** in two stages: first the lyrics (via an LLM), then the audio (via a text-to-music model); the resulting track is ingested into their personal catalog and streamable like any other track.

---

## 👥 Team & Responsibilities

The team consists of three engineers, each owning **two microservices** plus shared infrastructure responsibilities. Ownership means the engineer is accountable for design, implementation, testing, deployment, and on-call for those services — but pull requests and design reviews are cross-team.

| Engineer | Owned Microservices | Cross-Cutting Responsibilities |
|---|---|---|
| **Andrew Rudov** | `auth-service`, `user-service`, `payment-service` | Identity & monetization domain, JWT issuance, OAuth2 integration, NowPayments webhook handling, API Gateway auth middleware |
| **Danial Boranbayev** | `catalog-service`, `streaming-service` | Music catalog domain, S3-compatible object storage, audio chunking and CDN integration |
| **Daniyar Abdrakhmanov** | `playlist-service`, `recommendation-service`, `generation-service` | Personalization & generative domain, Kafka consumers for play events, AI lyric + music generation pipeline |

Shared responsibilities (rotating ownership):

- **API Gateway** — entry point for all client traffic, owned collectively but Andrew leads.
- **Infrastructure** (Kubernetes manifests, Terraform, CI/CD) — Danial leads.
- **Observability stack** (Prometheus, Grafana, Jaeger, Loki) — Daniyar leads.

---

## ✅ Functional Requirements

The system must support the following capabilities end-to-end:

**Identity & Account Management.** Users can register with email and password or via OAuth2 (Google, GitHub). Sessions are stateless and managed via short-lived JWT access tokens and longer-lived refresh tokens. Users can update their profile, avatar, and display name. Password resets work via email-delivered tokens.

**Music Catalog.** Administrators can ingest artists, albums, and tracks via an internal API. Each track has metadata (title, duration, genre, release date, ISRC) and an associated audio file stored in object storage. Tracks are immutable once published; metadata corrections create new versions.

**Audio Streaming.** Authenticated users can request a track and receive a signed, time-limited URL pointing to chunked audio segments (HLS or MPEG-DASH). Streaming supports byte-range requests for seeking. Multiple bitrates are available (96, 192, 320 kbps) and the client picks based on network conditions.

**Playlists.** Users can create unlimited playlists, add and reorder tracks, mark playlists public or private, and share them via a permalink. Collaborative playlists allow multiple owners.

**Search.** Users can search tracks, albums, and artists with prefix matching, fuzzy tolerance, and ranking by popularity. Search latency must be under 200 ms at p95.

**Recommendations.** The system produces a "Daily Mix" and "Discover Weekly" for each active user based on their play history. Recommendations are precomputed nightly and cached.

**Play Tracking & Analytics.** Every play event (track started, track completed, track skipped) is published to Kafka. Downstream consumers update play counts, feed the recommendation pipeline, and populate analytics dashboards.

**Payments & Monetization.** Premium-tier upgrades are processed through **NowPayments** (crypto). During or shortly after registration, the user may request a premium upgrade and receives a NowPayments invoice with a deposit address and amount. NowPayments delivers status callbacks via signed webhooks (IPN — Instant Payment Notifications). The system verifies each webhook with **HMAC-SHA512** against the IPN secret, idempotently records the payment, and on a `finished` status upgrades the user tier from `free` to `premium`. Status transitions follow `waiting → confirming → confirmed → sending → finished`, with failure paths `failed`, `expired`, and `refunded`. The webhook receiver is at-least-once: duplicate deliveries are dropped by a unique key on `(payment_id, status, occurred_at)`.

**AI Song Generation.** Premium users can compose songs in two sequential stages. **Stage 1 — Lyrics:** the user submits a prompt (theme, mood, language, genre); the system calls **Anthropic Claude** with a structured template and returns lyrics tagged by section (verse, chorus, bridge). The user can edit and re-run the stage until satisfied. **Stage 2 — Music:** the finalized lyrics are submitted to a text-to-music provider (**Suno API** in v1, with **Replicate** and **Udio** as configurable fallbacks); generation runs as an async job (typically 30–120 s). When the audio is ready, the service uploads it to MinIO, runs it through the same HLS transcoding pipeline used for ingested catalog tracks (96/192/320 kbps), and registers it in the catalog under a `user_generated = true` flag owned by the requesting user. Job progress is streamed to the client via Server-Sent Events backed by Kafka.

---

## ⚙️ Non-Functional Requirements

The system targets the following operational characteristics, sized for a pet project but designed to scale:

- **Availability:** 99.5% monthly uptime for the streaming path; 99.0% for non-critical paths (recommendations, analytics).
- **Latency:** p95 under 150 ms for catalog and playlist reads; p95 under 200 ms for search; first audio byte under 500 ms.
- **Scalability:** Each service is horizontally scalable behind Kubernetes. The streaming-service is the most resource-sensitive and is sharded by track ID hash.
- **Consistency:** Strong consistency within a service boundary; eventual consistency across services via Kafka. Playlist edits are read-your-writes per user.
- **Security:** TLS everywhere (mTLS between services via a service mesh), JWT for client auth, secrets in HashiCorp Vault, rate limiting at the gateway.
- **Observability:** Every request is traced end-to-end; every service exposes Prometheus metrics; structured JSON logs are shipped to Loki.
- **Cost:** Designed to run on a single Kubernetes cluster (3 nodes minimum) for development; production sizing assumes ~10k DAU.

---

## 🛠 Technology Stack

The stack is chosen to balance modern best practices with the team's existing Go expertise.

**Languages & Frameworks.** All services are written in **Go 1.22+**. HTTP handlers (where needed) use `chi` or the standard `net/http` package. gRPC services use `google.golang.org/grpc` with `protoc-gen-go` for code generation. Validation is handled by `go-playground/validator`. Database access uses `pgx` for PostgreSQL and `redis/go-redis/v9` for Redis.

**Inter-Service Communication.** **gRPC** with Protocol Buffers v3 is the default for synchronous RPCs. **Apache Kafka** (via the `segmentio/kafka-go` client) handles asynchronous events. **Redis Pub/Sub** is used for lightweight, low-latency notifications that don't need durability.

**Data Storage.** **PostgreSQL 16** is the primary relational store, one logical database per service (no shared schemas). **Redis 7** is used for caching, session data, and rate limiting. **MinIO** (S3-compatible) stores audio files and album artwork. **Elasticsearch 8** powers search. **ClickHouse** stores play events for analytics queries.

**API Gateway.** A custom Go-based gateway built on `chi` and `grpc-gateway` translates external REST/JSON requests into internal gRPC calls. Alternatively, `Kong` or `Traefik` could be used; we chose to build our own to control auth and rate-limiting logic tightly.

**Service Mesh & Networking.** **Linkerd** provides mTLS, retries, and traffic shifting between service versions. We chose Linkerd over Istio for its lower operational complexity.

**Container & Orchestration.** **Docker** for containerization, **Kubernetes** (k3s for local dev, EKS or DigitalOcean for prod) for orchestration. **Helm** charts for deployment, **Terraform** for cloud infrastructure.

**Observability.** **Prometheus** for metrics, **Grafana** for dashboards, **Jaeger** with **OpenTelemetry** for distributed tracing, **Loki** with **Promtail** for log aggregation. **Sentry** for error reporting from the application layer.

**CI/CD.** **GitHub Actions** runs tests, builds Docker images, and pushes to a container registry on every PR. **ArgoCD** handles GitOps-based deployment to the Kubernetes cluster.

**Payments.** **NowPayments** as the crypto payment processor — REST API for invoice creation, IPN webhook with **HMAC-SHA512** signature verification, supports BTC, ETH, USDT, USDC. Secrets (`NOWPAYMENTS_API_KEY`, `NOWPAYMENTS_IPN_SECRET`) live in **HashiCorp Vault**.

**Generative AI.** **Anthropic Claude** (`anthropic-sdk-go`) for lyric generation; **Suno API** for text-to-music in v1, with **Replicate** and **Udio** as drop-in alternatives behind a provider interface in `internal/adapter/aigen/`. The same HLS pipeline as `streaming-service` handles transcoding of generated audio.

**Auxiliary.** **HashiCorp Vault** for secrets, **Buf** for protobuf linting and breaking-change detection, **golangci-lint** for static analysis, **k6** for load testing.

---

## 🏛 System Architecture

The architecture follows a classic microservices pattern with a single API Gateway as the entry point. Internal services communicate over a mesh of gRPC calls and a Kafka event bus. Each service owns its data and exposes a well-defined contract.

```
                             ┌─────────────────────────┐
                             │       Web / Mobile      │
                             │         Clients         │
                             └────────────┬────────────┘
                                          │ HTTPS (REST/JSON)
                                          ▼
                             ┌─────────────────────────┐
                             │      API Gateway        │
                             │  (auth, rate limiting,  │
                             │   request translation)  │
                             └────────────┬────────────┘
                                          │ gRPC (mTLS via Linkerd)
        ┌─────────────────┬───────────────┼────────────────┬──────────────────┐
        ▼                 ▼               ▼                ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌────────────┐  ┌──────────────┐  ┌────────────────┐
│ auth-service │  │ user-service │  │  catalog-  │  │  playlist-   │  │ recommendation-│
│              │  │              │  │  service   │  │   service    │  │    service     │
└──────┬───────┘  └──────┬───────┘  └─────┬──────┘  └──────┬───────┘  └────────┬───────┘
       │                 │                │                │                   │
       ▼                 ▼                ▼                ▼                   ▼
   ┌───────┐         ┌───────┐        ┌───────┐        ┌───────┐           ┌──────────┐
   │  PG   │         │  PG   │        │  PG   │        │  PG   │           │ ClickHse │
   │ Redis │         │       │        │  ES   │        │ Redis │           │  Redis   │
   └───────┘         └───────┘        └───────┘        └───────┘           └──────────┘
                                          │                                     ▲
                                          ▼                                     │
                                  ┌───────────────┐                             │
                                  │  streaming-   │                             │
                                  │   service     │                             │
                                  └───────┬───────┘                             │
                                          │                                     │
                                          ▼                                     │
                                      ┌───────┐                                 │
                                      │ MinIO │                                 │
                                      │  CDN  │                                 │
                                      └───────┘                                 │
                                                                                │
                  ┌─────────────────────── Kafka Event Bus ──────────────────────┤
                  │   Topics: play.events, user.events, catalog.events, ...     │
                  └──────────────────────────────────────────────────────────────┘
```

### Architectural Principles

1. **Single Responsibility per Service.** Each microservice owns one bounded context. No service reaches into another's database.
2. **Database per Service.** Data ownership is explicit. Cross-service data needs go through gRPC calls or are derived from Kafka events into local read models.
3. **Async by Default for Side Effects.** Anything that doesn't need to block the user (analytics, recommendation updates, notifications) is published to Kafka.
4. **Idempotency Everywhere.** All write APIs accept an idempotency key. All Kafka consumers are idempotent.
5. **Backwards-Compatible Contracts.** Protobuf schemas follow Buf's breaking-change rules. Old clients keep working through at least one major version.
6. **Failure Isolation.** Circuit breakers (`sony/gobreaker`) wrap external calls. The streaming path degrades gracefully if recommendations are down.

---

## 🧩 Microservices Breakdown

Each of the six core microservices is described below with its responsibilities, data model, key endpoints, and owner.

### 1. `auth-service` — *Owner: Andrew Rudov*

The authentication service is the gatekeeper of the system. It handles user registration, login, token issuance, and token validation.

**Responsibilities.** Validate credentials, hash passwords with `bcrypt` (cost 12), issue short-lived JWT access tokens (15 min) and refresh tokens (30 days), maintain a token revocation list in Redis, integrate OAuth2 with Google and GitHub via the `golang.org/x/oauth2` package, and expose a `ValidateToken` gRPC endpoint that the API Gateway calls on every request.

**Data.** PostgreSQL for credentials and OAuth identities (table: `auth_users`, `oauth_identities`, `refresh_tokens`). Redis for the revocation list and rate limiting (failed login attempts).

**Key gRPC Methods.** `Register`, `Login`, `RefreshToken`, `Logout`, `ValidateToken`, `OAuthCallback`.

**Events Published.** `auth.user_registered`, `auth.user_logged_in`.

### 2. `user-service` — *Owner: Andrew Rudov*

The user service owns the user profile domain. It is intentionally separated from auth so that profile data can scale and evolve independently of credentials.

**Responsibilities.** Manage user profiles (display name, avatar URL, country, language), follow/unfollow relationships between users, and expose user-facing read APIs. Listens to `auth.user_registered` to provision a profile row.

**Data.** PostgreSQL (tables: `users`, `user_follows`, `user_preferences`). Avatars stored in MinIO with the URL persisted in Postgres.

**Key gRPC Methods.** `GetUser`, `UpdateProfile`, `FollowUser`, `UnfollowUser`, `ListFollowers`, `ListFollowing`, `UploadAvatar`.

**Events Consumed.** `auth.user_registered`.

**Events Published.** `user.profile_updated`, `user.followed`.

### 3. `catalog-service` — *Owner: Danial Boranbayev*

The catalog service owns the music metadata domain — artists, albums, tracks, and genres — and provides the search backend.

**Responsibilities.** CRUD operations on artists, albums, and tracks (admin-only writes). Maintain an Elasticsearch index synchronized with Postgres via change-data-capture (Debezium → Kafka → Elasticsearch consumer). Expose fast read APIs for the gateway and other services. Serve public catalog browse endpoints.

**Data.** PostgreSQL as the source of truth (tables: `artists`, `albums`, `tracks`, `genres`, `track_genres`). Elasticsearch as the search read model. Album art and other images in MinIO.

**Key gRPC Methods.** `GetTrack`, `GetAlbum`, `GetArtist`, `ListTracksByAlbum`, `ListAlbumsByArtist`, `Search` (with filters: type, genre, year), `IngestTrack` (admin), `UpdateTrack` (admin).

**Events Published.** `catalog.track_added`, `catalog.track_updated`, `catalog.album_added`.

### 4. `streaming-service` — *Owner: Danial Boranbayev*

The streaming service handles audio delivery. It is the most performance-sensitive service in the system.

**Responsibilities.** Validate playback authorization (does this user have access to this track?), generate signed URLs for HLS manifests and segments stored in MinIO, transcode uploaded audio into multiple bitrates as a background job, emit play events to Kafka.

**Audio Pipeline.** When a track is ingested via `catalog-service`, an event is published. The streaming-service consumes it, downloads the source audio from MinIO, transcodes to HLS at 96/192/320 kbps using `ffmpeg`, and writes the segments back to MinIO. The HLS manifest references segment URLs that are signed at request time.

**Data.** PostgreSQL for playback session metadata and transcoding job state. MinIO for audio storage. Redis for active session tracking and play-rate limiting.

**Key gRPC Methods.** `GetStreamManifest`, `RegisterPlay`, `RegisterSkip`, `RegisterComplete`, `GetPlaybackSession`.

**HTTP Endpoints (via gateway).** `GET /stream/{trackId}/manifest.m3u8` returns the signed HLS manifest.

**Events Consumed.** `catalog.track_added` (triggers transcoding).

**Events Published.** `play.started`, `play.completed`, `play.skipped`.

### 5. `playlist-service` — *Owner: Daniyar Abdrakhmanov*

The playlist service manages user-curated track collections.

**Responsibilities.** CRUD on playlists, ordered track membership, public/private visibility, collaborative playlists with multiple owners, sharing via permalinks (short, opaque IDs). Caches hot playlists (recently accessed) in Redis with a 5-minute TTL.

**Data.** PostgreSQL (tables: `playlists`, `playlist_tracks`, `playlist_collaborators`). The `playlist_tracks` table uses a fractional ranking column (a `numeric` value) to allow O(1) reorder operations without rewriting all rows. Redis caches serialized playlist views.

**Key gRPC Methods.** `CreatePlaylist`, `GetPlaylist`, `UpdatePlaylist`, `DeletePlaylist`, `AddTrack`, `RemoveTrack`, `ReorderTracks`, `ListUserPlaylists`, `AddCollaborator`.

**Events Consumed.** `catalog.track_updated` (to invalidate cache when track metadata changes).

**Events Published.** `playlist.created`, `playlist.track_added`.

### 6. `recommendation-service` — *Owner: Daniyar Abdrakhmanov*

The recommendation service is the personalization brain. It is read-heavy and tolerant of staleness.

**Responsibilities.** Consume `play.*` events from Kafka into ClickHouse to build a user-track interaction matrix, run a nightly batch job that computes recommendations using a simple collaborative filtering approach (item-item similarity via cosine, computed on co-listening counts), serve precomputed recommendations from Redis with a 24-hour TTL, expose endpoints for "Daily Mix", "Discover Weekly", and "Because you listened to X".

**Algorithm.** v1 ships with item-based collaborative filtering — for each track, precompute the top-50 most similar tracks based on co-occurrence in user listening sessions within a 7-day window. v2 will add a content-based component using genre and acoustic features. v3 may introduce an embedding model.

**Data.** ClickHouse for the play event log and similarity computations. Redis for the served recommendations. Postgres for user-level recommendation metadata (last-generated-at, model version).

**Key gRPC Methods.** `GetDailyMix`, `GetDiscoverWeekly`, `GetSimilarTracks`, `GetRecommendationsForTrack`.

**Events Consumed.** `play.started`, `play.completed`, `play.skipped`.

### 7. `payment-service` — *Owner: Andrew Rudov*

The payment service handles monetization. NowPayments is the only payment provider in v1; the service is structured so additional providers (Stripe, Paddle, native cards) can be added behind a `PaymentProvider` port.

**Responsibilities.** Issue invoices via the NowPayments REST API on user request, receive and verify NowPayments IPN webhooks (HMAC-SHA512 against the IPN secret), idempotently persist payment state transitions, publish `payment.*` events to Kafka, and signal `user-service` to upgrade a user's tier on a `finished` payment.

**Webhook Flow.** A dedicated HTTP endpoint (`POST /api/v1/payments/nowpayments/webhook`) is exposed through the API Gateway. On each call: validate the `x-nowpayments-sig` header against the canonical JSON body, look up the payment by `payment_id`, apply the state transition under an idempotency key, and publish `payment.status_changed`. Because IPN delivery is at-least-once, the `payments_events` table has a unique constraint on `(payment_id, status, occurred_at)` to drop duplicates.

**Registration Hook.** On `auth.user_registered`, the service optionally pre-creates a draft payment record so the user can immediately request an upgrade without a roundtrip to provider account setup. The draft is garbage-collected after 24 hours if unused.

**Data.** PostgreSQL (tables: `payments`, `payments_events`, `webhook_log`). Redis for outbound-API rate-limiting and short-lived idempotency keys.

**Key gRPC Methods.** `CreateInvoice`, `GetPayment`, `ListPaymentsByUser`, `HandleWebhook` (internal, called by the gateway's HTTP-to-gRPC bridge).

**HTTP Endpoints (via gateway).** `POST /api/v1/payments/nowpayments/webhook` — public, signature-protected.

**Events Consumed.** `auth.user_registered`.

**Events Published.** `payment.invoice_created`, `payment.status_changed`, `payment.finished`, `payment.failed`.

### 8. `generation-service` — *Owner: Daniyar Abdrakhmanov*

The generation service produces user-created songs via a two-stage AI pipeline: lyrics first, then music. It is the most user-visible "magic" of the platform and is restricted to premium users.

**Responsibilities.** Run a stateful, two-stage job per song.

- **Stage 1 — Lyrics.** Call the **Anthropic Claude** API with a structured prompt template. Returned text is parsed into sections (verse, chorus, bridge) and stored as a `lyrics_version`. The user can submit revisions; each becomes a new version.
- **Stage 2 — Music.** Submit finalized lyrics to a text-to-music provider (Suno API in v1; Replicate/Udio configurable). Poll or receive a webhook on completion, download the audio, upload to MinIO, then trigger the same HLS transcoding pipeline used by `catalog-service`. Once transcoded, the track is registered in the catalog under `user_generated = true`, owned by the requesting user.

**Premium Gate.** Generation is restricted to users with `tier = premium`. The service reads the `tier` claim from the JWT and additionally calls `user-service.GetUser` for defense-in-depth. Free users can preview Stage 1 (lyrics only) up to N times per day, configurable via env, with results not persisted beyond the session.

**Job Lifecycle.** Stages: `lyrics_queued → lyrics_ready → music_queued → music_ready → ingesting → published`, with `failed` reachable from any non-terminal step. Each transition publishes a Kafka event; the client subscribes via Server-Sent Events through the API Gateway.

**Data.** PostgreSQL (tables: `generation_jobs`, `lyrics_versions`, `provider_calls`). MinIO for intermediate and final audio. Redis for active job state and per-user provider rate-limit counters.

**Key gRPC Methods.** `GenerateLyrics`, `ReviseLyrics`, `GenerateMusic`, `GetJob`, `ListUserJobs`, `CancelJob`.

**HTTP Endpoints (via gateway).** `GET /api/v1/generation/jobs/{id}/events` — Server-Sent Events stream for live progress.

**Events Consumed.** `payment.finished` (to unlock premium features), `catalog.track_added` (to confirm successful ingestion of a generated track).

**Events Published.** `generation.lyrics_ready`, `generation.music_ready`, `generation.published`, `generation.failed`.

---

## 💾 Data Storage Strategy

The "database per service" rule is enforced strictly. Each service has its own PostgreSQL logical database, and no service queries another's tables directly. Cross-service data is replicated via Kafka events into local read models when needed.

| Service | Primary Store | Cache | Other |
|---|---|---|---|
| auth-service | PostgreSQL | Redis (revocation, rate limit) | — |
| user-service | PostgreSQL | — | MinIO (avatars) |
| catalog-service | PostgreSQL | Redis (hot tracks) | Elasticsearch (search), MinIO (art) |
| streaming-service | PostgreSQL | Redis (sessions) | MinIO (audio segments) |
| playlist-service | PostgreSQL | Redis (hot playlists) | — |
| recommendation-service | ClickHouse | Redis (precomputed lists) | PostgreSQL (metadata) |
| payment-service | PostgreSQL | Redis (rate limit, idempotency) | — |
| generation-service | PostgreSQL | Redis (job state, provider quotas) | MinIO (raw + final audio) |

**Schema Migrations.** Each service uses `golang-migrate/migrate` with version-controlled SQL files. Migrations run as a Kubernetes Job before the service deployment. Backwards-incompatible changes follow the expand-contract pattern.

**Backup & Recovery.** PostgreSQL: daily logical backups via `pg_dump` to object storage, plus continuous WAL archiving for point-in-time recovery. ClickHouse: daily snapshots. MinIO: replicated bucket. RPO target: 1 hour. RTO target: 4 hours.

---

## 🔌 Inter-Service Communication

**Synchronous (gRPC).** Used when a caller needs an immediate response — for example, the API Gateway calling `auth-service.ValidateToken` or `playlist-service` calling `catalog-service.GetTrack` to enrich a playlist response. All gRPC calls go through Linkerd, which provides mTLS, retries with exponential backoff, and per-route timeouts. Deadlines are propagated via `context.Context`.

**Asynchronous (Kafka).** Used for events that don't require an immediate response or that fan out to multiple consumers. Topics are partitioned by a natural key (user ID for user events, track ID for play events) to preserve ordering where it matters. Consumer groups use at-least-once delivery; consumers are idempotent.

**Event Schema.** All Kafka events use Protobuf-encoded payloads with a schema registered in a Buf-managed schema registry. Every event has a standard envelope: `event_id` (UUID), `event_type`, `occurred_at`, `producer`, `trace_id`, and a typed `payload`.

**Topic Naming Convention.** `{domain}.{aggregate}.{event}` — for example `play.session.started`, `catalog.track.added`, `user.profile.updated`.

---

## 🚪 API Gateway & Routing

The API Gateway is the single ingress point for all client traffic. It is a custom Go service that performs five jobs: TLS termination, authentication, request validation, rate limiting, and gRPC translation.

**Authentication.** On every request, the gateway extracts the JWT from the `Authorization: Bearer <token>` header and calls `auth-service.ValidateToken` (with a 200 ms timeout and a local cache of valid tokens for 30 seconds to reduce load). The resulting user context is propagated to downstream services via gRPC metadata.

**Rate Limiting.** Token bucket implementation backed by Redis. Limits are tiered: 100 req/min for unauthenticated, 1000 req/min for authenticated, 10000 req/min for premium accounts.

**Translation.** The gateway uses `grpc-gateway` to map REST/JSON routes to internal gRPC calls. The mapping is defined in proto annotations.

Example route mapping:

```
GET  /api/v1/tracks/{id}              → catalog.CatalogService/GetTrack
GET  /api/v1/playlists/{id}           → playlist.PlaylistService/GetPlaylist
POST /api/v1/playlists                → playlist.PlaylistService/CreatePlaylist
GET  /api/v1/stream/{trackId}/manifest.m3u8 → streaming.StreamingService/GetStreamManifest
GET  /api/v1/recommendations/daily    → recommendation.RecommendationService/GetDailyMix
GET  /api/v1/search?q={query}         → catalog.CatalogService/Search
POST /api/v1/payments/invoices               → payment.PaymentService/CreateInvoice
GET  /api/v1/payments/{id}                   → payment.PaymentService/GetPayment
POST /api/v1/payments/nowpayments/webhook    → payment.PaymentService/HandleWebhook   (public, HMAC-protected)
POST /api/v1/generation/lyrics               → generation.GenerationService/GenerateLyrics
POST /api/v1/generation/lyrics/{id}/revise   → generation.GenerationService/ReviseLyrics
POST /api/v1/generation/music                → generation.GenerationService/GenerateMusic
GET  /api/v1/generation/jobs/{id}            → generation.GenerationService/GetJob
GET  /api/v1/generation/jobs/{id}/events     → SSE stream backed by Kafka generation.* events
```

---

## 🔐 Authentication & Authorization

**Authentication.** Email + password (bcrypt-hashed) or OAuth2 (Google, GitHub). On successful authentication, the user receives a JWT access token (signed with RS256, 15-minute TTL) and a refresh token (opaque, stored in Postgres, 30-day TTL).

**JWT Claims.** Standard claims (`sub`, `iat`, `exp`, `iss`, `aud`) plus custom claims: `user_id`, `email`, `roles[]`, `tier` (free | premium).

**Authorization.** Role-based access control (RBAC) enforced at the service level. Roles are checked in gRPC interceptors. The roles in use are `user` (default), `premium`, `admin` (catalog ingestion), and `service` (for service-to-service calls).

**Service-to-Service Auth.** Each service has its own JWT issued at startup and rotated daily. mTLS via Linkerd provides identity at the network layer; the JWT carries the application-level identity.

---

## 📊 Observability & Monitoring

Observability is treated as a first-class concern. Every service is instrumented from day one — adding observability later is an anti-pattern.

**Metrics.** Each service exposes a `/metrics` endpoint scraped by Prometheus. Standard metrics include request count, request duration histograms, in-flight requests, error rate, and Go runtime metrics. Custom business metrics include plays per minute, signups per hour, search latency by query type, and recommendation cache hit rate.

**Tracing.** OpenTelemetry SDK instruments all gRPC and HTTP handlers. Traces flow into Jaeger. Trace context (`traceparent`) is propagated through gRPC metadata and Kafka headers, giving us end-to-end visibility across async boundaries.

**Logging.** Structured JSON logs written to stdout, collected by Promtail, stored in Loki. Every log line includes `service`, `trace_id`, `user_id` (when known), and `event`. Log levels are configurable per service via env var.

**Dashboards.** Grafana hosts a per-service dashboard plus a top-level "faqears Overview" dashboard with the four golden signals (latency, traffic, errors, saturation) for the streaming path.

**Alerting.** Alertmanager routes alerts to Slack. SLO-based alerting: a burn-rate alert fires when error budget is being consumed too fast over a 1-hour or 6-hour window.

---

## ☁️ Deployment & Infrastructure

**Local Development.** `docker-compose` brings up Postgres, Redis, Kafka (Redpanda for simplicity), MinIO, Elasticsearch, and ClickHouse. Each service runs locally with `air` for hot reload. Skaffold optionally deploys to a local k3s cluster.

**Staging / Production.** Kubernetes cluster with three node pools: general (services), data (stateful sets for databases — though for production we recommend managed services), and streaming (higher network bandwidth nodes for `streaming-service`). Each service deploys as a `Deployment` with 2+ replicas, a `Service`, and a `HorizontalPodAutoscaler` based on CPU and custom metrics.

**Infrastructure as Code.** Terraform provisions the cloud resources (VPC, EKS/DOKS cluster, RDS instances, S3 buckets). Helm charts define each service's Kubernetes manifests. ArgoCD watches the Helm chart repo and applies changes automatically (GitOps).

**CI/CD Pipeline.** GitHub Actions workflow on every PR: lint with `golangci-lint`, run unit tests with race detector, build Docker image, push to GHCR with the commit SHA tag. On merge to `main`, the image tag is updated in the Helm values file, and ArgoCD picks up the change and deploys to staging. Promotion to production requires a manual approval gate.

---

## 📁 Repository Structure

We use a **monorepo** managed with Go workspaces. This simplifies cross-service refactors and shared code (proto definitions, common middleware) while allowing each service to be deployed independently.

```
faqears/
├── README.md
├── go.work
├── docker-compose.yml
├── .github/
│   └── workflows/         # CI/CD definitions
├── proto/                 # Shared protobuf definitions
│   ├── auth/
│   ├── user/
│   ├── catalog/
│   ├── streaming/
│   ├── playlist/
│   ├── recommendation/
│   └── events/            # Kafka event schemas
├── pkg/                   # Shared Go libraries
│   ├── logger/
│   ├── tracing/
│   ├── metrics/
│   ├── errors/
│   ├── auth/              # JWT validation helpers
│   └── kafka/
├── services/
│   ├── auth-service/
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── handler/   # gRPC handlers
│   │   │   ├── service/   # Business logic
│   │   │   ├── repository/# Data access
│   │   │   └── domain/    # Domain models
│   │   ├── migrations/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── user-service/
│   ├── catalog-service/
│   ├── streaming-service/
│   ├── playlist-service/
│   ├── recommendation-service/
│   ├── payment-service/
│   └── generation-service/
├── gateway/               # API Gateway
├── deploy/
│   ├── helm/              # Helm charts per service
│   ├── terraform/         # Infrastructure
│   └── k8s/               # Raw manifests for local
└── docs/
    ├── adr/               # Architecture Decision Records
    └── api/               # OpenAPI specs (generated)
```

**Why Monorepo?** Easier to share proto files and refactor across boundaries. The trade-off is that CI must be smart about only rebuilding changed services — we use path filters in GitHub Actions for this.

---

## 🔄 Development Workflow

**Branching.** Trunk-based development. Short-lived feature branches off `main`, merged via PR with at least one review. No long-running develop or release branches.

**Commits.** Conventional Commits format (`feat(catalog): add genre filter`). Semantic versioning is derived from commit messages.

**Pull Requests.** Every PR must pass CI (lint, test, build), include tests for new code, and update relevant docs. Cross-service contract changes require updating the proto file and both producer and consumer in the same PR (or using expand-contract for breaking changes).

**Testing Strategy.** Unit tests for business logic (target 70% coverage on `internal/service`), integration tests against ephemeral Postgres/Redis via `testcontainers-go`, contract tests for gRPC services, end-to-end smoke tests in staging via `k6`.

**Code Review Checklist.** Before merging: error handling complete, context propagation correct, observability hooks added, no SQL injection or N+1 queries, idempotency for writes, backwards-compatible proto changes.

---

## 🗺 Roadmap & Milestones

The project is planned in five milestones over roughly four months of part-time work.

**Milestone 1 — Foundations (Weeks 1–3).** Repository setup, CI/CD pipeline, Docker Compose dev environment, shared `pkg/` libraries (logger, tracing, metrics), proto schema layout, Helm chart template. Deliverable: a "hello world" service deployable to k3s with full observability.

**Milestone 2 — Identity (Weeks 4–6).** `auth-service` and `user-service` complete with registration, login, JWT issuance, profile management. API Gateway with auth middleware. Deliverable: a user can register, log in, and view their profile end-to-end.

**Milestone 3 — Catalog & Streaming (Weeks 7–10).** `catalog-service` with admin ingestion and Elasticsearch-backed search. `streaming-service` with HLS transcoding pipeline. Deliverable: an admin can ingest a track, a user can search for it and stream it.

**Milestone 4 — Playlists & Personalization (Weeks 11–14).** `playlist-service` with full CRUD and collaborative playlists. `recommendation-service` with v1 collaborative filtering. Kafka pipeline for play events. Deliverable: a user can build playlists and receive a Daily Mix.

**Milestone 5 — Hardening (Weeks 15–16).** Load testing with k6, chaos testing with Litmus, security audit, SLO definition and burn-rate alerts, documentation polish, public demo deployment. Deliverable: the system runs in a production-like environment with documented SLOs.

**Milestone 6 — Monetization & Generative AI (Weeks 17–20).** `payment-service` with NowPayments invoice issuance and HMAC-verified webhook handling; the `auth.user_registered → payment.draft_created → payment.finished → user.tier_changed` pipeline. `generation-service` with two-stage Claude lyrics + Suno music generation, premium-tier gating, reuse of `streaming-service`'s HLS pipeline for user-generated tracks, and SSE-based progress streaming through the API Gateway. Deliverable: a premium user can pay in crypto, write a song with AI assistance, and stream their own track from their personal catalog.

---

## 📚 Further Reading & References

The architecture draws on standard distributed systems patterns. Recommended reading for the team:

- *Designing Data-Intensive Applications* — Martin Kleppmann
- *Building Microservices* (2nd ed.) — Sam Newman
- *Database Internals* — Alex Petrov
- *Site Reliability Engineering* — Google SRE Team
- The gRPC documentation at grpc.io
- The Buf documentation at buf.build for protobuf workflow

---

## 📜 License

This is a learning project. Code is licensed under the MIT License. Audio content used during development is either owned by the team or licensed for educational use; no copyrighted material is redistributed.

---

**Maintained by:** Andrew Rudov, Danial Boranbayev, Daniyar Abdrakhmanov.
