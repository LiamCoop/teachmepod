# teachmepod

A private two-host technical podcast generated from a backlog. See [PLAN.md](PLAN.md).

## Deployment

Railway project `teachmepod`, production environment:

- `podcastd`: builds the `Dockerfile` from `main` on GitHub and redeploys on push, but only when code, migrations, `show/` or the `Dockerfile` change (docs-only pushes don't redeploy). Health check `/healthz`. Public at https://podcastd-production.up.railway.app.
- `Postgres`: reached by `podcastd` over the private network (`DATABASE_URL` references `${{Postgres.DATABASE_URL}}`).
- `audio` bucket (sjc): private, audio only. `podcastd` gets `AWS_ENDPOINT_URL_S3`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AUDIO_BUCKET` as references to the bucket's variables.

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
