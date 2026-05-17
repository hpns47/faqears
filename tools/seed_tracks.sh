#!/usr/bin/env bash
#
# Seeds the FAQears catalog with artists, albums and real, playable tracks.
# Audio is fetched from SoundHelix (freely usable, algorithmically generated music).
#
# Usage:
#   ./tools/seed_tracks.sh
#
# Env overrides:
#   FAQEARS_API          gateway base URL          (default http://localhost:8080)
#   FAQEARS_ADMIN_EMAIL  admin login               (default admin@faqears.local)
#   FAQEARS_ADMIN_PASS   admin password            (default admin12345)
#
# Idempotent: existing artists/albums/tracks (matched by name/title) are reused,
# so it is safe to re-run — only missing audio gets uploaded.

set -euo pipefail

API="${FAQEARS_API:-http://localhost:8080}"
ADMIN_EMAIL="${FAQEARS_ADMIN_EMAIL:-admin@faqears.local}"
ADMIN_PASS="${FAQEARS_ADMIN_PASS:-admin12345}"
SH="https://www.soundhelix.com/examples/mp3"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

log()  { printf '\033[0;36m%s\033[0m\n' "$*"; }
ok()   { printf '  \033[0;32m♫\033[0m %s\n' "$*"; }
skip() { printf '  \033[0;90m·\033[0m %s\n' "$*"; }
warn() { printf '  \033[0;33m!\033[0m %s\n' "$*"; }

jget() { python3 -c "import sys,json
try:
    d=json.load(sys.stdin)
    print(d.get('$1','') if isinstance(d,dict) else '')
except Exception:
    print('')"; }

# find_by name|title  <kind: artists|tracks>  -> echoes first exact-match id
find_entity() {
  local q="$1" kind="$2" field
  [ "$kind" = artists ] && field=name || field=title
  curl -fsS -G "$API/api/v1/search" -H "$AUTH" \
    --data-urlencode "q=$q" --data-urlencode "limit=50" \
    | python3 -c "import sys,json
d=json.load(sys.stdin)
for x in d.get('$kind',[]) or []:
    if x.get('$field','').strip().lower()=='''$q'''.strip().lower():
        print(x['id']); break"
}

TOKEN=$(curl -fsS -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}" | jget access_token)
if [ -z "$TOKEN" ]; then
  echo "ERROR: admin login failed at $API — is the stack up?"
  exit 1
fi
AUTH="Authorization: Bearer $TOKEN"
log "Logged in as $ADMIN_EMAIL → $API"

SEEDED=0
REUSED=0

ensure_artist() {  # name country bio -> id
  local id
  id=$(find_entity "$1" artists)
  if [ -n "$id" ]; then echo "$id"; return; fi
  curl -fsS -X POST "$API/api/v1/admin/catalog/artists" -H "$AUTH" \
    -H 'Content-Type: application/json' \
    -d "{\"name\":\"$1\",\"country\":\"$2\",\"biography\":\"$3\"}" | jget artist_id
}

ensure_album() {  # artist_id title year -> id
  local id
  id=$(find_entity "$2" albums)
  if [ -n "$id" ]; then echo "$id"; return; fi
  curl -fsS -X POST "$API/api/v1/admin/catalog/albums" -H "$AUTH" \
    -H 'Content-Type: application/json' \
    -d "{\"artist_id\":\"$1\",\"title\":\"$2\",\"year\":$3}" | jget album_id
}

# album_id artist_id title duration genre soundhelix_song_number
seed_track() {
  local tid file="$TMP/song-$6.mp3" code
  tid=$(find_entity "$3" tracks)
  if [ -z "$tid" ]; then
    tid=$(curl -fsS -X POST "$API/api/v1/admin/catalog/tracks" -H "$AUTH" \
      -H 'Content-Type: application/json' \
      -d "{\"album_id\":\"$1\",\"artist_id\":\"$2\",\"title\":\"$3\",\"duration_sec\":$4,\"genres\":[\"$5\"]}" \
      | jget track_id)
  fi
  if [ -z "$tid" ]; then
    warn "track '$3' — create failed"
    return
  fi
  if ! curl -fsS -o "$file" --max-time 90 "$SH/SoundHelix-Song-$6.mp3"; then
    warn "track '$3' — audio download failed"
    return
  fi
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST \
    "$API/api/v1/admin/streaming/tracks/$tid/audio" \
    -H "$AUTH" -F "file=@$file;type=audio/mpeg")
  case "$code" in
    200|201)
      # one play so the track surfaces in recommendations
      local sid
      sid=$(curl -fsS -X POST "$API/api/v1/streaming/play" -H "$AUTH" \
        -H 'Content-Type: application/json' -d "{\"track_id\":\"$tid\"}" | jget session_id)
      [ -n "$sid" ] && curl -fsS -X POST "$API/api/v1/streaming/complete" -H "$AUTH" \
        -H 'Content-Type: application/json' \
        -d "{\"track_id\":\"$tid\",\"session_id\":\"$sid\",\"duration_sec\":$4}" > /dev/null || true
      ok "$3"
      SEEDED=$((SEEDED + 1))
      ;;
    409)
      skip "$3 (audio already present)"
      REUSED=$((REUSED + 1))
      ;;
    *)
      warn "track '$3' — audio upload failed (HTTP $code)"
      ;;
  esac
}

log "Aurora Bloom"
A=$(ensure_artist "Aurora Bloom" "IS" "Reykjavik producer crafting glacial ambient electronica.")
AL=$(ensure_album "$A" "Polar Lights" 2023)
seed_track "$AL" "$A" "Glacier Drift"     377 "Ambient" 1
seed_track "$AL" "$A" "Northern Spectrum" 373 "Ambient" 2
seed_track "$AL" "$A" "Aurora Borealis"   421 "Ambient" 3

log "Midnight Cassette"
A=$(ensure_artist "Midnight Cassette" "US" "Synthwave duo chasing neon-lit nostalgia.")
AL=$(ensure_album "$A" "Neon Boulevard" 2024)
seed_track "$AL" "$A" "Neon Boulevard"  360 "Synthwave" 4
seed_track "$AL" "$A" "Midnight Drive"  334 "Synthwave" 5
seed_track "$AL" "$A" "Afterglow"       381 "Synthwave" 6

log "Echoform"
A=$(ensure_artist "Echoform" "DE" "Berlin modular live act exploring generative techno.")
AL=$(ensure_album "$A" "Modular Dreams" 2022)
seed_track "$AL" "$A" "Patch Bay"       241 "Techno" 7
seed_track "$AL" "$A" "Voltage Control" 300 "Techno" 8
seed_track "$AL" "$A" "Sequencer"       273 "Techno" 9

log "Saryarka Sound"
A=$(ensure_artist "Saryarka Sound" "KZ" "Almaty collective blending steppe motifs with house.")
AL=$(ensure_album "$A" "Steppe Frequencies" 2024)
seed_track "$AL" "$A" "Steppe Frequencies" 244 "House" 10
seed_track "$AL" "$A" "Dombyra Pulse"      240 "House" 11
seed_track "$AL" "$A" "Almaty Nights"      260 "House" 12

echo
log "Done — $SEEDED track(s) seeded, $REUSED already had audio."
echo "Open the app → Browse: all tracks should be there and playable."
