# Backlog Podcast — Implementation Plan

Oct 25, 2026 · @Liam

## Goals and scope

v1 builds the whole pipeline end to end, then spends its effort on a voice bake-off, not on producing episodes. The product is a private, two-host technical podcast (a cast of ten in four standing pairings, with a woman as lead expert in three) of 20–40 minutes depending on the topic, 2–3 times a week, for walks to and from the office.

**v1 delivers**

- A Postgres-backed backlog you can add to from your phone in under 30 seconds
- Every stage scaffolded and runnable: research, outline, script, review, audio, publish
- A script quality system aimed squarely at not sounding like AI slop
- Swappable TTS vendors behind one interface, so the same script renders in several voices
- A simple private site plus an RSS feed your podcast app subscribes to
- A bake-off: a few short test renders across vendors and voice pairs, then one full episode on the winner

**v1 deliberately does not**

- Produce a stream of episodes. A hard cap of 2 full episodes until the bake-off is decided.
- Auto-schedule anything. Scheduled runs come only after a full episode is worth hearing.
- Build the product features: roadmaps, interactive Q&A, multiple users.

## Architecture overview

One Go service holds the site, the feed and the pipeline workers; everything external is either an API it calls or storage it writes to.

&#91;embedded content: system architecture · one binary, two external APIs\]

You add topics and feedback through the site; the workers turn a topic into audio using Claude and a TTS vendor; your podcast app pulls episodes from the private feed.

## Data model

Postgres is the backlog, the pipeline state and the job queue. Large artifacts (dossiers, scripts, audio) live in object storage and are referenced by key. A script and its renders are separate rows, so one script can be rendered in many voices. That is the bake-off.

| Table | Purpose | Key columns |
| --- | --- | --- |
| `topics` | The backlog | title, angle, baseline (what you already know), depth, priority, status, series, created\_at |
| `episodes` | One per topic attempt | topic\_id, status, title, slug, script\_key, prompt\_version, published\_render\_id, published\_at |
| `stage_runs` | Every stage execution | episode\_id, stage, attempt, status, input\_hash, output\_key, tokens\_in, tokens\_out, cost\_cents, error, started\_at, finished\_at |
| `voice_profiles` | A vendor + voice pairing | name, vendor, model, voices (json: host → voice ID), settings (json) |
| `renders` | One audio rendering of a script | episode\_id, voice\_profile\_id, status, audio\_key, duration\_s, chapters (json), wer, cost\_cents, is\_test |
| `feedback` | Your reactions after listening | render\_id, kind (cringe, wrong, deeper, too basic, rating), at\_seconds, note |
| `show_memory` | Facts the hosts can call back to | episode\_id, line ("we covered Raft leader election") |

- **Statuses:** `queued → researching → scripting → reviewing → rendering → published`, plus `failed` and `parked`.
- **Jobs:** [River](https://riverqueue.com/), a Go job queue on Postgres, gives retries, uniqueness and scheduling without a second datastore.
- **Versioned in git, not the database:** persona bible, style guide, slop rules and prompts. `prompt_version` on each episode records which commit wrote it.

## Pipeline stages

Each stage is a River job that reads its inputs from object storage, writes one output, and records a `stage_runs` row. A stage skips itself when an output already exists for the same `input_hash` (hash of upstream outputs + prompt version). Reruns are cheap and resumable, and voice tests reuse the script without paying for research again.

1. **Pick.** Take the top `queued` topic by priority, or a topic you trigger by hand. Status moves to `researching`.
2. **Research → `dossier.json`.** A lead agent splits the topic into 4–6 subtopics; a research agent per subtopic uses web search. Source priority: papers and engineering blogs from teams that run the systems in question, then official docs, general articles last. Spending a little more on research for quality is fine. The dossier holds:
   - concepts ranked by interestingness (surprising, foundational, has a story, useful to you)
   - stories: incidents, postmortems, who built it and why
   - misconceptions and open debates (the raw material for disagreement)
   - sources with URLs, as summaries in our own words, never copied text
   - follow-up topic suggestions
3. **Outline → `outline.md`.** Length follows the material: 20–40 minutes, set by how many strong concepts and stories the dossier holds, in scenes of about 5 minutes. Episodes over 25 minutes get a recap and a natural break near the middle, so the evening walk picks up cleanly.
4. **Beats → `beats.md`.** What's actually interesting, scene by scene. Accuracy pass only, no dialogue yet.
5. **Dialogue → `script.json`.** Beats become conversation one scene at a time. Vendor-neutral format: `scenes[] → turns[] {speaker, text, tags[]}`, where tags are abstract (`laugh`, `interrupt`, `sigh`) and adapters map them to vendor syntax. About 150 words per minute of target length.
6. **Review.** Slop linter, adversarial editor and fact-check against the dossier, looping at most twice (next section). A script that still fails is `parked` with the reasons, not published.
7. **Render → `renders`.** Per scene, per voice profile: 2–3 takes, auto-picked (audio section).
8. **Mix → `episode.mp3`.** Join scenes, normalize loudness, embed chapters.
9. **Publish.** Mark the render published, write show notes and transcript, regenerate the feed, add `show_memory` lines.

## Script quality system

The slop lives in the script, so most of the anti-slop effort goes here. Each piece is a file in the repo or a check in code, so it can be tuned from your feedback.

**Inputs every script prompt receives**

- `personas.md`: profiles of only the hosts in this episode, drawn from the cast of ten in the Hosts tab. Each profile has background, how they talk, opinions, blind spot and a "never says" line.
- `style-guide.md`: writing for the ear. No code or URLs read aloud, signpost lists, spell out acronyms once, short sentences, recap per scene.
- Recent `show_memory` lines, so the hosts can call back to past episodes.

**Rules the script must meet**

- At least one genuine disagreement, taken from the dossier's open debates.
- At least one misconception stated by a host and then corrected.
- Every scene carries a concrete specific: a named system, a date, a number or an incident.
- Disfluency and audio tags used sparingly, a handful per scene at most.

**Checks, in order**

1. **Slop linter** (deterministic Go, `slop-rules.yaml`):
   - banned phrases ("dive in", "fascinating", "great point", "Absolutely!", "In today's episode")
   - no turn opening with praise
   - a cap on consecutive agreement turns
   - a minimum variance in turn length
   - flags acronyms used before they are spelled out
2. **Adversarial editor**, a separate agent: flag every line a real person wouldn't say out loud, with a reason.
3. **Fact-check agent**: every factual claim must trace to the dossier; untraceable claims are cut or rewritten.
4. **Rewrite** only the flagged lines, then rerun 1–3. At most two loops.

**The calibration loop.** Your `feedback` rows ("cringe at 12:30", "too basic") are reviewed after each episode. Recurring problems become new linter rules or style-guide lines. Expect the first \~10 episodes to be mostly calibration.

## Audio and the voice bake-off

The vendor is chosen by listening on real walks, not by spec sheets. One script, rendered through several voice profiles, gets you there with about five short renders and one full episode.

**Render path**

- `TTSAdapter` interface in Go: `RenderScene(scene, profile) → audio`. One adapter per vendor maps abstract tags to that vendor's syntax and handles its length limits.
- Render a whole scene (2–3 minutes) in one dialogue-aware call, not line by line, so turn-taking and tone carry across speakers.
- Generate 2–3 takes per scene. Transcribe each with speech-to-text and score word error rate against the script; keep the lowest. [ElevenLabs reports that regenerating fixes about half of quality issues](https://elevenlabs.io/docs/de/capabilities/text-to-dialogue).
- A pronunciation lexicon (`lexicon.yaml`: kubectl, NATS, Argo, Zig) applied before rendering.
- Mix with ffmpeg: small varied gaps between scenes, faint room tone, loudness normalized to about −16 LUFS, mono MP3, chapters embedded.

**Test matrix** — one \~4-minute scene with real disagreement and jargon in it

| Profile | Vendor and model | Why it's in |
| --- | --- | --- |
| EL-1, EL-2 | [ElevenLabs v3 Text to Dialogue](https://elevenlabs.io/docs/capabilities/text-to-dialogue), two voice pairs | Built for multi-speaker podcast dialogue, inline audio tags, voice design for the female host |
| GM-1, GM-2 | [Gemini TTS](https://fal.ai/learn/tools/best-text-to-speech-apis), two voice pairs | Top of a May 2026 speech-quality ranking; supports multi-speaker |
| AZ-1 | Azure AI Speech, one pair | You're already on Azure; the control |

**Scoring** (1–5, logged as `feedback` rows on each test render)

- Naturalness of turn-taking and reactions
- Each voice on its own: would you listen to it for 30 minutes?
- Pronunciation of technical terms
- Automated: word error rate and cost per minute

Test renders publish to the feed as "Test: EL-1" and so on, flagged `is_test`. The winner renders one full episode. A second full episode only if the top two are too close to call.

## Listening site and RSS feed

The site is server-rendered Go templates in the same binary as the worker: no separate frontend to deploy. It does four things and stays single-user behind a passkey login (WebAuthn, registered once with a setup token from an environment variable; HttpOnly, Secure, SameSite=Strict session cookie). The feed has its own rotatable token because podcast apps can't log in.

| Page | What it does |
| --- | --- |
| Backlog | Add a topic (title, angle, baseline, depth) from your phone; reorder; park; trigger a run by hand |
| Episodes | Status of each run by stage, cost, links to dossier and script; render a script with another voice profile |
| Episode | Player, transcript, chapters, sources; feedback form with a timestamp field ("cringe at 12:30") and "go deeper on…", which creates a backlog topic |
| Feed | `/feed/<secret-token>.xml`, RSS 2.0 with Podcasting 2.0 `podcast:transcript` and `podcast:chapters` |

- Railway buckets have no public URLs, so feed enclosures point at stable app URLs (/audio/\<token>/\<slug>.mp3) that redirect to a short-lived presigned URL. Downloads come straight from the bucket with free egress. The feed token can be rotated if it leaks.
- Any podcast app that subscribes by URL works, so nothing goes to Apple or Spotify directories.
- Show notes link the episode page, so sources and the transcript are one tap away during a walk.

## Orchestration, infra and repo layout

One Go binary, one Postgres, one object-storage bucket. River jobs chain the stages, so there is no separate workflow engine to run. Manual triggers only in v1; River's periodic jobs add the 2–3 times a week schedule in M4.

**Runtime**

- `podcastd serve` runs the site, the feed and the River workers in one process.
- Hosting: Railway, one service plus Railway Postgres.
- Object storage: a Railway bucket, private and S3-compatible. Railway doesn't back buckets up automatically; audio can be re-rendered from a script, so only scripts and dossiers need a copy elsewhere.
- ffmpeg in the container image.
- Secrets: Anthropic, TTS vendor and storage keys as environment variables.
- Agent stages call the Claude API from Go. Research uses the API's web search tool; heavier stages use the strongest model.

**Repo**

```
cmd/podcastd/          main: serve, migrate, run-stage (for local debugging)
internal/pipeline/     one package per stage + River job wiring
internal/tts/          TTSAdapter + elevenlabs/, gemini/, azure/
internal/audio/        take scoring (STT + WER), ffmpeg mixing
internal/lint/         slop linter
internal/feed/         RSS + Podcasting 2.0
internal/web/          handlers + templates
migrations/            SQL
show/                  personas.md, style-guide.md, slop-rules.yaml, lexicon.yaml
prompts/               research, outline, beats, dialogue, editor, factcheck
```

`run-stage --episode <id> --stage dialogue` reruns any single stage locally against real artifacts, which is where most prompt tuning will happen.

## Observability and cost controls

Every episode should answer "what did this cost and where did the time go" without opening a log. The real cost per episode is unknown until the bake-off, so v1 measures first and sets budgets from what it sees.

- **Traces:** one OpenTelemetry trace per episode, a span per stage and per agent or TTS call, with tokens, characters rendered and cost as attributes.
- **Cost rollup:** `cost_cents` on `stage_runs` and `renders`, summed on the Episodes page.
- **Budget guard:** a per-episode cap checked before each paid call. Crossing it parks the episode instead of finishing it.
- **Research caps:** maximum tool calls per research agent and maximum subagents per topic.
- **Episode cap:** the 2-full-episode limit from scope, enforced in code until you lift it.
- **Alerts (M4):** a failed or parked run sends a push or email.

## Milestones

The build runs in five phases; the bake-off in M2 is the decision everything after it depends on.

&#91;embedded content: build plan · 5 phases, each with a gate\]

M0 ships a minimal feed on purpose, so the M2 test renders reach your phone. Nothing is scheduled until M4, and the 2-full-episode cap holds until M3's gate is met.

## Open decisions and risks

**Decisions**

- [x] Hosting and storage: Railway service, Railway Postgres, Railway bucket
- [x] Episode length: flexible, 20–40 minutes, set by how much the topic carries
- [x] Research sources: papers and engineering blogs first
- [x] Site login: passkey, with a separate rotatable feed token
- [x] First test topics: microservices, API design, intermediate/advanced React. You're an expert here, so errors are easy to hear.
- [x] Hosts: all ten profiled, four standing pairings (Hosts tab)
- [ ] TTS vendor and voice pair (decided by the bake-off)
- [ ] Per-episode budget cap (later, once the bake-off shows real cost)

**Risks**

| Risk | Mitigation |
| --- | --- |
| It still sounds like slop | Script checks first; calibration from timestamped feedback; park rather than publish a weak script |
| Confident errors are hard to catch by ear | Fact-check against the dossier; sources on every episode page; one test topic you know well |
| Voice drift between scenes | Fixed voice IDs and settings per profile; scene-level takes scored by WER |
| Vendor length limits break long scenes | Adapters split scenes at turn boundaries |
| Cost per episode is higher than expected | Measured in the bake-off; per-episode budget guard |
| Copying source text into scripts | Dossier stores summaries and links only; editor flags long verbatim phrasing |
| Private feed URL leaks | Rotatable token; audio behind signed URLs |
