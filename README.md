# teachmepod

A private two-host technical podcast generated from a backlog. See [PLAN.md](PLAN.md).

## Deployment

Railway project `teachmepod`, production environment:

- `podcastd`: builds the `Dockerfile` from `main` on GitHub and redeploys on push, but only when code, migrations, `show/` or the `Dockerfile` change (docs-only pushes don't redeploy). Health check `/healthz`. Public at https://podcastd-production.up.railway.app.
- `Postgres`: reached by `podcastd` over the private network (`DATABASE_URL` references `${{Postgres.DATABASE_URL}}`).
- `audio` bucket (sjc): private, audio only. `podcastd` gets `AWS_ENDPOINT_URL_S3`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AUDIO_BUCKET` as references to the bucket's variables.

## Feed

- `/feed/<FEED_TOKEN>.xml` is the private RSS feed. Subscribe to it by URL in any podcast app. `FEED_TOKEN` is a variable on the `podcastd` service; changing it rotates the feed and every audio link in it.
- `/audio/<FEED_TOKEN>/<render id>/<slug>.mp3` redirects to a presigned bucket URL valid for 6 hours.
- The feed lists finished renders that are either voice tests or an episode's published render.

To hear a voice before the pipeline exists, add an MP3 as a test render. With the tunnel open and `DATABASE_URL` set as below, and the bucket variables copied from the `podcastd` service:

```sh
eval "$(railway variables --service podcastd --json | python3 -c 'import json,sys,shlex;v=json.load(sys.stdin);print("\n".join(f"export {k}={shlex.quote(v[k])}" for k in ["AWS_ENDPOINT_URL_S3","AWS_REGION","AWS_ACCESS_KEY_ID","AWS_SECRET_ACCESS_KEY","AUDIO_BUCKET"]))')"
go run ./cmd/podcastd add-test-render -profile EL-1 -title "Dana + Marcus, pair 1" clip.mp3
```

It needs `ffprobe` locally to read the duration.

## Voice auditions

[docs/hosts.md](docs/hosts.md) is the cast tour and the ElevenLabs shortlist. The machine-readable copy lives in `show/hosts.yaml` (`audition_line` and `elevenlabs` per host); edit both together.

```sh
go run ./cmd/podcastd auditions -dry-run        # what would be rendered, and how many characters
go run ./cmd/podcastd auditions [-host dana]    # render what's missing into the bucket
```

Variants for comparing a voice more fairly, each labelled on the review page:

```sh
go run ./cmd/podcastd auditions -host dana -voice Elara -takes 3            # more takes of the same line
go run ./cmd/podcastd auditions -host dana -voice Elara -stability robust   # creative | natural | robust
go run ./cmd/podcastd auditions -host dana -voice Jade -line "...[excited] Oh, I love this one..." -label "[excited] tag"
```

It needs the tunnel, the bucket variables (as for `add-test-render`) and `ELEVENLABS_API_KEY`, which lives on the `podcastd` service. A clip is never rendered twice for the same host, voice, model and line; changing a line renders it fresh.

Review at `/auditions/<FEED_TOKEN>`: hosts grouped by pairing, a player per voice, "Play pair" to hear lead then foil, and shortlist / reject with a note per clip.

## Database

Railway project `teachmepod`, Postgres service `Postgres` (no public proxy).

```sh
# open a tunnel (prints the connection URL)
railway connect Postgres --tunnel-only -P 55432

# in another shell
export DATABASE_URL=postgresql://postgres:<password>@127.0.0.1:55432/railway
go run ./cmd/podcastd setup   # migrate + seed
PORT=8080 go run ./cmd/podcastd serve   # migrate, then serve; GET /healthz checks the database
```

The `Dockerfile` builds the same binary into a Debian image with ffmpeg; its default command is `serve`.

- `migrations/` holds the schema (goose, embedded in the binary). `00002` is the knowledge graph from [ADR-001](adrs/ADR-001%20Topics%20as%20a%20knowledge%20graph.md).
- Text artifacts (dossiers, scripts, …) live in the `artifacts` table (`00004`); the bucket holds audio only.
- `show/hosts.yaml` is the cast and pairings. It is not stored in Postgres; topics and episodes refer to pairings and guests by slug.
- `show/seed.yaml` holds the initial user, backlog and bake-off voice profiles; `seed` inserts them only if missing.

### Concept graph conventions

- `prerequisite`: src must be learned before dst. `part_of`: src is a child of dst. `related`: stored once with `src_id < dst_id`.
- Both rules hold for approved edges only, so research can propose competing parents and edges side by side: a trigger rejects approving an edge that would close a cycle in `prerequisite` or `part_of`, and a partial unique index allows one approved `part_of` parent per concept.
- `proposed_by_episode_id` on concepts and edges records which episode's research proposed them (`00003`).
- `concepts.embedding` is an untyped `vector`; fix its dimension and add an HNSW index once the embedding model is chosen.
