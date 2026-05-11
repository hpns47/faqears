#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

BOLD='\033[1m'
DIM='\033[2m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GREY='\033[0;90m'
NC='\033[0m'

AUTH_ADDR=${AUTH_ADDR:-localhost:50061}
USER_ADDR=${USER_ADDR:-localhost:50062}
CATALOG_ADDR=${CATALOG_ADDR:-localhost:50063}
GATEWAY_URL=${GATEWAY_URL:-http://localhost:8080}

PASS=0
FAIL=0

banner() {
    printf "${BOLD}${CYAN}"
    printf "================================================================\n"
    printf "  FAQears  --  end-to-end demo (docker compose + gRPC + HTTP)\n"
    printf "  auth: %s   user: %s\n" "$AUTH_ADDR" "$USER_ADDR"
    printf "  catalog: %s   gateway: %s\n" "$CATALOG_ADDR" "$GATEWAY_URL"
    printf "================================================================${NC}\n"
}

section() {
    printf "\n${BOLD}${CYAN}== %s ==${NC}\n" "$1"
}

step() {
    printf "${GREY}  > %s${NC}\n" "$1"
}

show() {
    printf "${GREY}    %s${NC}\n" "$1"
}

ok() {
    printf "  ${GREEN}[ OK ]${NC} %s\n" "$1"
    PASS=$((PASS+1))
}

bad() {
    printf "  ${RED}[FAIL]${NC} %s\n" "$1"
    if [[ -n "${2:-}" ]]; then
        while IFS= read -r line; do
            printf "         ${RED}%s${NC}\n" "$line"
        done <<< "$2"
    fi
    FAIL=$((FAIL+1))
}

require_tool() {
    if ! command -v "$1" >/dev/null 2>&1; then
        printf "${RED}missing required tool: %s${NC}\n" "$1"
        printf "${YELLOW}install: %s${NC}\n" "$2"
        exit 1
    fi
}

resolve_grpcurl() {
    if command -v grpcurl >/dev/null 2>&1; then
        command -v grpcurl
    elif [[ -x "$HOME/go/bin/grpcurl" ]]; then
        echo "$HOME/go/bin/grpcurl"
    else
        return 1
    fi
}

wait_for_health() {
    local name=$1
    local tries=${2:-90}
    while ((tries > 0)); do
        local status
        status=$(docker inspect --format='{{.State.Health.Status}}' "$name" 2>/dev/null || echo missing)
        if [[ "$status" == "healthy" ]]; then
            return 0
        fi
        sleep 1
        tries=$((tries-1))
    done
    return 1
}

call_auth() {
    "$GRPCURL" -plaintext -d "$2" "$AUTH_ADDR" "auth.v1.AuthService/$1" 2>&1
}

call_user() {
    "$GRPCURL" -plaintext -d "$2" "$USER_ADDR" "user.v1.UserService/$1" 2>&1
}

call_catalog() {
    "$GRPCURL" -plaintext -d "$2" "$CATALOG_ADDR" "catalog.v1.CatalogService/$1" 2>&1
}

gw_get() {
    local path=$1 token=${2:-}
    if [[ -n "$token" ]]; then
        curl -sS -H "Authorization: Bearer $token" -o /tmp/gw_body -w '%{http_code}' "$GATEWAY_URL$path"
    else
        curl -sS -o /tmp/gw_body -w '%{http_code}' "$GATEWAY_URL$path"
    fi
}

gw_post() {
    local path=$1 body=$2 token=${3:-}
    if [[ -n "$token" ]]; then
        curl -sS -X POST -H "Authorization: Bearer $token" -H 'Content-Type: application/json' \
            -d "$body" -o /tmp/gw_body -w '%{http_code}' "$GATEWAY_URL$path"
    else
        curl -sS -X POST -H 'Content-Type: application/json' \
            -d "$body" -o /tmp/gw_body -w '%{http_code}' "$GATEWAY_URL$path"
    fi
}

gw_body() {
    cat /tmp/gw_body 2>/dev/null
}

expect_error() {
    local addr=$1 fqm=$2 data=$3 expected=$4
    local out
    if out=$("$GRPCURL" -plaintext -d "$data" "$addr" "$fqm" 2>&1); then
        echo "$out"
        return 1
    fi
    if echo "$out" | grep -q "$expected"; then
        return 0
    fi
    echo "$out"
    return 1
}

CLEAN=0
SKIP_UP=0
for arg in "$@"; do
    case "$arg" in
        -c|--clean) CLEAN=1 ;;
        -s|--skip-up) SKIP_UP=1 ;;
        -h|--help)
            cat <<EOF
demo.sh -- start the FAQears stack and run end-to-end gRPC tests.

usage: ./demo.sh [options]
  -c, --clean      docker compose down -v before starting (fresh DBs)
  -s, --skip-up    skip docker compose up (assume stack already running)
  -h, --help       this help

env overrides:
  AUTH_ADDR       default localhost:50061
  USER_ADDR       default localhost:50062
  CATALOG_ADDR    default localhost:50063
  GATEWAY_URL     default http://localhost:8080
EOF
            exit 0
            ;;
    esac
done

banner

require_tool docker "Docker Desktop or apt-get install docker.io docker-compose-plugin"
require_tool jq     "apt-get install jq    /    brew install jq"

GRPCURL=$(resolve_grpcurl) || {
    printf "${RED}grpcurl not found${NC}\n"
    printf "${YELLOW}install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest${NC}\n"
    exit 1
}

if [[ $CLEAN -eq 1 ]]; then
    section "Tearing down previous stack"
    docker compose down -v --remove-orphans 2>&1 | tail -5
fi

if [[ $SKIP_UP -eq 0 ]]; then
    section "Starting stack via docker compose up -d --build"
    docker compose up -d --build 2>&1 | tail -20
fi

section "Waiting for health checks"
for svc in faqears-postgres-auth-1 faqears-postgres-users-1 faqears-postgres-catalog-1 faqears-kafka-1 faqears-redis-1 faqears-auth-service-1 faqears-user-service-1 faqears-catalog-service-1; do
    step "$svc"
    if wait_for_health "$svc" 90; then
        ok "$svc healthy"
    else
        bad "$svc never became healthy"
        docker compose logs "$svc" 2>&1 | tail -30
        exit 1
    fi
done

step "faqears-api-gateway-1 (HTTP, no Docker healthcheck — probe /healthz)"
gateway_up=0
for i in $(seq 1 30); do
    if curl -sf -o /dev/null "$GATEWAY_URL/healthz"; then
        gateway_up=1
        break
    fi
    sleep 1
done
if [[ $gateway_up -eq 1 ]]; then
    ok "api-gateway responds on /healthz"
else
    bad "api-gateway did not respond on /healthz within 30s"
    docker compose logs api-gateway 2>&1 | tail -30
    exit 1
fi

STAMP=$(date +%s)
ALICE="alice-$STAMP@example.com"
BOB="bob-$STAMP@example.com"
PASSWORD="supersecret123"

section "auth.Register  --  happy path"
step "Register $ALICE"
RES=$(call_auth Register "{\"email\":\"$ALICE\",\"password\":\"$PASSWORD\"}")
ALICE_ID=$(echo "$RES" | jq -r '.userId // empty')
show "response: $RES"
if [[ -n "$ALICE_ID" ]]; then
    ok "Alice registered, user_id=$ALICE_ID"
else
    bad "Register Alice failed" "$RES"
    exit 1
fi

section "auth.Register  --  duplicate email rejected"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/Register" "{\"email\":\"$ALICE\",\"password\":\"$PASSWORD\"}" "AlreadyExists"; then
    ok "duplicate email rejected with AlreadyExists"
else
    bad "expected AlreadyExists on duplicate"
fi

section "auth.Register  --  validation"
step "weak password"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/Register" "{\"email\":\"weak-$STAMP@example.com\",\"password\":\"12\"}" "InvalidArgument"; then
    ok "weak password rejected with InvalidArgument"
else
    bad "expected InvalidArgument on weak password"
fi
step "bad email"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/Register" "{\"email\":\"notanemail\",\"password\":\"longenough123\"}" "InvalidArgument"; then
    ok "bad email rejected with InvalidArgument"
else
    bad "expected InvalidArgument on bad email"
fi

section "auth.Login  --  happy path"
LOGIN=$(call_auth Login "{\"email\":\"$ALICE\",\"password\":\"$PASSWORD\"}")
ACCESS=$(echo "$LOGIN" | jq -r '.accessToken // empty')
REFRESH=$(echo "$LOGIN" | jq -r '.refreshToken // empty')
show "access  ${ACCESS:0:60}..."
show "refresh ${REFRESH:0:60}..."
if [[ -n "$ACCESS" && -n "$REFRESH" ]]; then
    ok "Login returned access + refresh tokens"
else
    bad "Login response malformed" "$LOGIN"
    exit 1
fi

section "auth.Login  --  wrong password rejected"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/Login" "{\"email\":\"$ALICE\",\"password\":\"WRONG\"}" "Unauthenticated"; then
    ok "wrong password rejected with Unauthenticated"
else
    bad "expected Unauthenticated on wrong password"
fi

section "auth.ValidateToken  --  valid token"
VAL=$(call_auth ValidateToken "{\"access_token\":\"$ACCESS\"}")
VAL_USER=$(echo "$VAL" | jq -r '.userId // empty')
VAL_EMAIL=$(echo "$VAL" | jq -r '.email // empty')
show "claims: $VAL"
if [[ "$VAL_USER" == "$ALICE_ID" && "$VAL_EMAIL" == "$ALICE" ]]; then
    ok "claims match Alice"
else
    bad "claims mismatch" "$VAL"
fi

section "auth.ValidateToken  --  bad token rejected"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/ValidateToken" "{\"access_token\":\"not.a.real.token\"}" "Unauthenticated"; then
    ok "bad token rejected with Unauthenticated"
else
    bad "expected Unauthenticated on bad token"
fi

section "auth.RefreshToken  --  rotation"
REF=$(call_auth RefreshToken "{\"refresh_token\":\"$REFRESH\"}")
NEW_ACCESS=$(echo "$REF" | jq -r '.accessToken // empty')
NEW_REFRESH=$(echo "$REF" | jq -r '.refreshToken // empty')
show "new access  ${NEW_ACCESS:0:60}..."
show "new refresh ${NEW_REFRESH:0:60}..."
if [[ -n "$NEW_ACCESS" && -n "$NEW_REFRESH" && "$NEW_REFRESH" != "$REFRESH" ]]; then
    ok "RefreshToken issued a new pair (refresh rotated)"
else
    bad "refresh failed or refresh token not rotated" "$REF"
fi

section "auth.RefreshToken  --  old refresh rejected"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/RefreshToken" "{\"refresh_token\":\"$REFRESH\"}" "Unauthenticated"; then
    ok "old refresh token no longer accepted"
else
    bad "old refresh token still works -- rotation broken"
fi

section "auth.Logout"
LOUT=$(call_auth Logout "{\"access_token\":\"$NEW_ACCESS\"}")
show "$LOUT"
ok "Logout returned response"

section "kafka  --  user-service auto-provisions Alice from auth.user_registered"
step "waiting up to 15s for event consumption"
PROVISIONED=0
for i in $(seq 1 15); do
    if call_user GetUser "{\"user_id\":\"$ALICE_ID\"}" | grep -q "$ALICE"; then
        PROVISIONED=1
        show "provisioned after ${i}s"
        break
    fi
    sleep 1
done
if [[ $PROVISIONED -eq 1 ]]; then
    ok "Alice profile auto-provisioned via Kafka event"
else
    bad "Alice profile not provisioned within 15s"
fi

section "user.GetUser  --  happy path"
USR=$(call_user GetUser "{\"user_id\":\"$ALICE_ID\"}")
show "$USR"
if echo "$USR" | grep -q "$ALICE"; then
    ok "GetUser returns Alice profile"
else
    bad "GetUser failed" "$USR"
fi

section "user.GetUser  --  unknown id rejected"
if expect_error "$USER_ADDR" "user.v1.UserService/GetUser" "{\"user_id\":\"00000000-0000-0000-0000-000000000000\"}" "NotFound"; then
    ok "unknown id returned NotFound"
else
    bad "expected NotFound"
fi

section "user.UpdateProfile"
UP=$(call_user UpdateProfile "{\"user_id\":\"$ALICE_ID\",\"display_name\":\"Alice the Tester\",\"country\":\"KZ\",\"language\":\"ru\"}")
show "$UP"
if echo "$UP" | grep -q "Alice the Tester" && echo "$UP" | grep -q "\"KZ\""; then
    ok "profile updated (display_name + country applied)"
else
    bad "UpdateProfile failed" "$UP"
fi

section "auth.Register  --  second account (Bob)"
BOB_RES=$(call_auth Register "{\"email\":\"$BOB\",\"password\":\"$PASSWORD\"}")
BOB_ID=$(echo "$BOB_RES" | jq -r '.userId // empty')
if [[ -n "$BOB_ID" ]]; then
    ok "Bob registered, user_id=$BOB_ID"
else
    bad "Bob register failed" "$BOB_RES"
    exit 1
fi
step "waiting for Bob provision"
for i in $(seq 1 15); do
    if call_user GetUser "{\"user_id\":\"$BOB_ID\"}" | grep -q "$BOB"; then
        break
    fi
    sleep 1
done

section "user.FollowUser  --  Alice follows Bob"
call_user FollowUser "{\"follower_id\":\"$ALICE_ID\",\"followee_id\":\"$BOB_ID\"}" >/dev/null
ok "FollowUser ok"

section "user.ListFollowers(Bob) contains Alice"
LF=$(call_user ListFollowers "{\"user_id\":\"$BOB_ID\",\"limit\":50}")
if echo "$LF" | grep -q "$ALICE_ID"; then
    ok "Bob.followers contains Alice"
else
    bad "Alice not in Bob.followers" "$LF"
fi

section "user.ListFollowing(Alice) contains Bob"
LFG=$(call_user ListFollowing "{\"user_id\":\"$ALICE_ID\",\"limit\":50}")
if echo "$LFG" | grep -q "$BOB_ID"; then
    ok "Alice.following contains Bob"
else
    bad "Bob not in Alice.following" "$LFG"
fi

section "user.FollowUser  --  self-follow rejected"
if expect_error "$USER_ADDR" "user.v1.UserService/FollowUser" "{\"follower_id\":\"$ALICE_ID\",\"followee_id\":\"$ALICE_ID\"}" "InvalidArgument"; then
    ok "self-follow rejected with InvalidArgument"
else
    bad "expected InvalidArgument on self-follow"
fi

section "user.UnfollowUser"
call_user UnfollowUser "{\"follower_id\":\"$ALICE_ID\",\"followee_id\":\"$BOB_ID\"}" >/dev/null
LF2=$(call_user ListFollowers "{\"user_id\":\"$BOB_ID\",\"limit\":50}")
if echo "$LF2" | grep -q "$ALICE_ID"; then
    bad "Alice still in Bob.followers after unfollow" "$LF2"
else
    ok "Unfollow removed the relationship"
fi

section "catalog.IngestArtist + GetArtist"
ART_RES=$(call_catalog IngestArtist '{"name":"Saryarka Demo","country":"KZ","biography":"end-to-end test"}')
ARTIST_ID=$(echo "$ART_RES" | jq -r '.artistId // empty')
show "ingest: $ART_RES"
if [[ -n "$ARTIST_ID" ]]; then
    ok "Artist ingested, id=$ARTIST_ID"
else
    bad "IngestArtist failed" "$ART_RES"
    exit 1
fi
G_ART=$(call_catalog GetArtist "{\"artist_id\":\"$ARTIST_ID\"}")
if echo "$G_ART" | grep -q "Saryarka Demo"; then
    ok "GetArtist returns the record"
else
    bad "GetArtist mismatch" "$G_ART"
fi

section "catalog.IngestAlbum + GetAlbum + ListAlbumsByArtist"
ALB_RES=$(call_catalog IngestAlbum "{\"artist_id\":\"$ARTIST_ID\",\"title\":\"Steppe Echoes\",\"year\":2025,\"cover_url\":\"http://minio/x.jpg\"}")
ALBUM_ID=$(echo "$ALB_RES" | jq -r '.albumId // empty')
if [[ -n "$ALBUM_ID" ]]; then
    ok "Album ingested, id=$ALBUM_ID"
else
    bad "IngestAlbum failed" "$ALB_RES"
    exit 1
fi
LBA=$(call_catalog ListAlbumsByArtist "{\"artist_id\":\"$ARTIST_ID\",\"limit\":10}")
if echo "$LBA" | grep -q "$ALBUM_ID"; then
    ok "ListAlbumsByArtist contains the new album"
else
    bad "ListAlbumsByArtist missing album" "$LBA"
fi

section "catalog.IngestTrack + GetTrack + ListTracksByAlbum + Redis cache"
TRK_TITLE="Wind on Saryarka $STAMP"
TRK_RES=$(call_catalog IngestTrack "{\"album_id\":\"$ALBUM_ID\",\"artist_id\":\"$ARTIST_ID\",\"title\":\"$TRK_TITLE\",\"duration_sec\":195,\"isrc\":\"KZ-DEMO-25-$STAMP\",\"genres\":[\"folk\",\"electronic\"]}")
TRACK_ID=$(echo "$TRK_RES" | jq -r '.trackId // empty')
if [[ -n "$TRACK_ID" ]]; then
    ok "Track ingested, id=$TRACK_ID"
else
    bad "IngestTrack failed" "$TRK_RES"
    exit 1
fi
GT1=$(call_catalog GetTrack "{\"track_id\":\"$TRACK_ID\"}")
GT2=$(call_catalog GetTrack "{\"track_id\":\"$TRACK_ID\"}")
if echo "$GT1" | grep -q "$TRK_TITLE" && echo "$GT2" | grep -q "$TRK_TITLE"; then
    ok "GetTrack served twice (second hit should be Redis-cached)"
else
    bad "GetTrack inconsistent" "$GT1 // $GT2"
fi
if docker exec -t faqears-redis-1 redis-cli KEYS 'catalog:track:*' 2>&1 | grep -q "$TRACK_ID"; then
    ok "Redis contains catalog:track:$TRACK_ID cache entry"
else
    show "$(docker exec -t faqears-redis-1 redis-cli KEYS 'catalog:track:*' 2>&1)"
    bad "no Redis cache entry observed for the track"
fi

section "catalog.Search  --  ILIKE on title"
SR=$(call_catalog Search "{\"query\":\"$STAMP\",\"limit\":50}")
if echo "$SR" | grep -q "$TRACK_ID"; then
    ok "Search finds the new track by stamped title"
else
    bad "Search did not find the new track" "$SR"
fi

section "kafka  --  catalog.events topic exists"
if docker exec -t faqears-kafka-1 rpk topic list 2>&1 | grep -q "catalog.events"; then
    ok "catalog.events topic created by publisher"
else
    bad "catalog.events topic missing"
fi

section "gateway  --  /healthz + /readyz"
HC=$(gw_get /healthz)
RC=$(gw_get /readyz)
show "healthz: $(gw_body)  http=$HC"
if [[ "$HC" == "200" ]]; then ok "/healthz returns 200"; else bad "/healthz returned $HC"; fi
if [[ "$RC" == "200" ]]; then ok "/readyz returns 200"; else bad "/readyz returned $RC"; fi

section "gateway  --  OpenAPI 3.1 spec served at /openapi.yaml"
HC=$(gw_get /openapi.yaml)
if [[ "$HC" == "200" ]] && gw_body | head -1 | grep -q "^openapi: 3"; then
    ok "/openapi.yaml served, valid OpenAPI 3.x header"
else
    bad "/openapi.yaml not served" "http=$HC"
fi

section "gateway  --  Swagger UI served at /swagger/"
HC=$(gw_get /swagger/)
if [[ "$HC" == "200" ]] && gw_body | grep -q "swagger-ui"; then
    ok "Swagger UI rendered — open http://localhost:8080/swagger/ in browser"
else
    bad "/swagger/ not served" "http=$HC"
fi

section "gateway  --  POST /api/v1/auth/register (HTTP→gRPC)"
GW_STAMP=$(date +%s%N | tail -c 10)
GW_EMAIL="gw-$GW_STAMP@example.com"
GW_PASS="longenough123"
HTTP=$(gw_post /api/v1/auth/register "{\"email\":\"$GW_EMAIL\",\"password\":\"$GW_PASS\"}")
GW_USER_ID=$(gw_body | jq -r '.user_id // .userId // empty')
show "body: $(gw_body)  http=$HTTP"
if [[ ( "$HTTP" == "200" || "$HTTP" == "201" ) && -n "$GW_USER_ID" ]]; then
    ok "register via gateway → user_id=$GW_USER_ID"
else
    bad "gateway register failed" "http=$HTTP body=$(gw_body)"
    exit 1
fi

section "gateway  --  POST /api/v1/auth/login"
HTTP=$(gw_post /api/v1/auth/login "{\"email\":\"$GW_EMAIL\",\"password\":\"$GW_PASS\"}")
GW_ACCESS=$(gw_body | jq -r '.access_token // .accessToken // empty')
if [[ "$HTTP" == "200" && -n "$GW_ACCESS" ]]; then
    ok "login via gateway returned access_token"
else
    bad "gateway login failed" "http=$HTTP body=$(gw_body)"
    exit 1
fi

section "gateway  --  GET /api/v1/tracks/{id} without Bearer  →  401"
HTTP=$(gw_get "/api/v1/tracks/$TRACK_ID")
if [[ "$HTTP" == "401" ]]; then
    ok "unauthenticated request rejected with 401"
else
    bad "expected 401, got $HTTP" "$(gw_body)"
fi

section "gateway  --  GET /api/v1/tracks/{id} with Bearer  →  200"
HTTP=$(gw_get "/api/v1/tracks/$TRACK_ID" "$GW_ACCESS")
if [[ "$HTTP" == "200" ]] && gw_body | grep -q "Wind on Saryarka"; then
    ok "track fetched through gateway"
else
    bad "gateway track fetch failed" "http=$HTTP body=$(gw_body)"
fi

section "gateway  --  GET /api/v1/search?q=$STAMP"
HTTP=$(gw_get "/api/v1/search?q=$STAMP&limit=50" "$GW_ACCESS")
if [[ "$HTTP" == "200" ]] && gw_body | grep -q "$TRACK_ID"; then
    ok "search through gateway found the track"
else
    bad "gateway search failed" "http=$HTTP body=$(gw_body)"
fi

section "gateway  --  user auto-provision  +  GET /api/v1/users/{id}"
for i in $(seq 1 15); do
    HTTP=$(gw_get "/api/v1/users/$GW_USER_ID" "$GW_ACCESS")
    if [[ "$HTTP" == "200" ]] && gw_body | grep -q "$GW_EMAIL"; then
        break
    fi
    sleep 1
done
if [[ "$HTTP" == "200" ]] && gw_body | grep -q "$GW_EMAIL"; then
    ok "user profile fetched via gateway (auto-provisioned)"
else
    bad "user profile unavailable via gateway" "http=$HTTP body=$(gw_body)"
fi

section "gateway  --  POST /api/v1/admin/catalog/artists  (non-admin → 403)"
HTTP=$(gw_post /api/v1/admin/catalog/artists '{"name":"Wannabe","country":"KZ"}' "$GW_ACCESS")
if [[ "$HTTP" == "403" ]]; then
    ok "admin endpoint rejected non-admin caller with 403"
else
    bad "expected 403, got $HTTP" "$(gw_body)"
fi

section "oauth template  --  AuthorizeURL issues a real state into Redis"
OA=$(call_auth OAuthAuthorizeURL '{"provider":"google","redirect_uri":"http://localhost:8080/oauth/cb"}')
OA_STATE=$(echo "$OA" | jq -r '.state // empty')
OA_URL=$(echo "$OA" | jq -r '.authorizeUrl // .authorize_url // empty')
show "url=${OA_URL:0:80}..."
show "state=$OA_STATE"
if [[ -n "$OA_STATE" && "$OA_URL" == https://accounts.google.com/* ]]; then
    ok "AuthorizeURL produced a real Google URL and state token"
else
    bad "AuthorizeURL malformed" "$OA"
fi
EX=$(docker exec faqears-redis-1 redis-cli EXISTS "auth:oauth_state:$OA_STATE" 2>&1 | tr -d '\r\n ')
if [[ "$EX" == "1" ]]; then
    ok "state token persisted in Redis (auth:oauth_state:*)"
else
    show "EXISTS returned: '$EX'"
    show "all state keys: $(docker exec faqears-redis-1 redis-cli KEYS 'auth:oauth_state:*' 2>&1)"
    bad "state token not found in Redis"
fi

section "oauth template  --  Callback consumes state then hits NotImplemented stub"
if expect_error "$AUTH_ADDR" "auth.v1.AuthService/OAuthCallback" "{\"provider\":\"google\",\"code\":\"fake-code\",\"state\":\"$OA_STATE\"}" "not implemented"; then
    ok "provider.Exchange returns 'not implemented' (template behavior)"
else
    bad "expected 'not implemented' from stub"
fi

section "db isolation  --  postgres-auth"
docker exec -t faqears-postgres-auth-1 psql -U faqears -d auth -c "\dt" 2>&1 | sed "s/^/    /"
if docker exec -t faqears-postgres-auth-1 psql -U faqears -d auth -c "\dt" 2>&1 | grep -qE "user_follows|tracks|albums|artists"; then
    bad "postgres-auth contains foreign-domain tables -- isolation broken"
else
    ok "auth db has only auth-domain tables"
fi

section "db isolation  --  postgres-users"
docker exec -t faqears-postgres-users-1 psql -U faqears -d users -c "\dt" 2>&1 | sed "s/^/    /"
if docker exec -t faqears-postgres-users-1 psql -U faqears -d users -c "\dt" 2>&1 | grep -qE "auth_users|refresh_tokens|tracks|albums|artists"; then
    bad "postgres-users contains foreign-domain tables -- isolation broken"
else
    ok "users db has only user-domain tables"
fi

section "db isolation  --  postgres-catalog"
docker exec -t faqears-postgres-catalog-1 psql -U faqears -d catalog -c "\dt" 2>&1 | sed "s/^/    /"
if docker exec -t faqears-postgres-catalog-1 psql -U faqears -d catalog -c "\dt" 2>&1 | grep -qE "auth_users|refresh_tokens|user_follows"; then
    bad "postgres-catalog contains foreign-domain tables -- isolation broken"
else
    ok "catalog db has only catalog-domain tables"
fi

section "kafka topics"
docker exec -t faqears-kafka-1 rpk topic list 2>&1 | sed "s/^/    /"

section "container status"
docker compose ps --format "table {{.Name}}\t{{.Status}}" | sed "s/^/    /"

section "summary"
TOTAL=$((PASS+FAIL))
printf "  ${BOLD}PASS:${NC} ${GREEN}%d${NC}    ${BOLD}FAIL:${NC} ${RED}%d${NC}    ${BOLD}TOTAL:${NC} %d\n" "$PASS" "$FAIL" "$TOTAL"
if [[ $FAIL -eq 0 ]]; then
    printf "  ${BOLD}${GREEN}all checks passed${NC}\n"
    exit 0
else
    printf "  ${BOLD}${RED}%d check(s) failed${NC}\n" "$FAIL"
    exit 1
fi
