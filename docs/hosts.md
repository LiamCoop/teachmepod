# Hosts

Ten hosts, four standing pairings. Each episode uses one pairing, chosen by topic. Rachel and Sam rotate in as guests. These profiles become `show/personas.md` in the repo, and every script prompt receives the two (or three) profiles it needs.

## Pairings

Each pair is built on a disagreement the two hosts can have honestly, episode after episode. In every pair, a woman is the lead expert.

| Pair | Lead | Foil | The friction | Topics | Running bit |
| --- | --- | --- | --- | --- | --- |
| Dana + Marcus | Dana | Marcus | Evidence vs. the hot take | Microservices, distributed systems, architecture patterns, consistency | Marcus declares something dead; Dana finds the paper where it's thriving |
| Maya + Jules | Maya | Jules | Scale vs. "just ship it" | API design, databases, caching, infrastructure, platform teams | "At a billion requests a day…" — Jules counts how many times she says it |
| Claire + Theo | Claire | Theo | Precision vs. curiosity | React, rendering, browsers, frontend performance, state management | Theo's game-engine analogies, which Claire rates out of ten |
| Priya + Kofi | Priya | Kofi | What breaks vs. who owns it | Reliability, observability, incidents, team topology, migrations | Priya's "what pages you?" and Kofi's "who owns that?" as a two-question test |

**First test topics:** microservices → Dana + Marcus; API design → Maya + Jules; intermediate/advanced React → Claire + Theo.

**Guests:** Rachel joins when a topic needs a "what do I do on Monday" grounding. Sam joins whenever a design has an attack surface worth exploring. A guest takes a third of the airtime at most.

## Voice casting (ElevenLabs)

**Deadline that shapes every pick.** ElevenLabs retires all of its Default voices (Rachel, Adam, Brian, Sarah and the rest of the classic set) on December 31, 2026, and accounts created after March 2026 can't use them at all. Every first pick below is from the new permanent lineup ElevenLabs published as replacements, so the show won't lose a host's voice at year end.

**Cast, after auditions (Oct 26, 2026).** Every host was auditioned on `eleven_v3` in the first pick and alternates below. No voice is shared between hosts.

| Host | Cast voice | Voice ID | From the audition |
| --- | --- | --- | --- |
| Dana | Jade – Upbeat and Natural | `g7LVvkPWALzPxOQbF6OE` | Alternate; Elara rejected |
| Marcus | Eddie – Natural and Helpful | `l7kNoIfnJKPg7779LI2t` | Alternate; Lawrence rejected ("sounds very AI, and the sound quality is not very good") |
| Maya | Elowen – Upbeat Modern Narrator | `dvbL7qkNGZY1IqPGZAjM` | First pick; Jade is Dana's |
| Jules | Maisie – Friendly Casual Neighbor | `QtY3JBOUKEB5xzrRfOKc` | First pick |
| Claire | Alicia – Polished Global Anchor | `BFd5oBc2DDna33pSi4Gf` | First pick |
| Theo | Warren – Effortless and Cool | `7QN34D2r3hCNwbOYIeK0` | First pick |
| Priya | Florence – Atmospheric Storyteller | `22N9cF8z0o7y23njdyaY` | First pick; still a placeholder until Voice Design is tried |
| Kofi | Darian – Warm Grounded Storyteller | `gOupLcAkjEnguROwi4oS` | Alternate |
| Rachel | Talia – Warm Soft Guide | `OZ0L6eISlOejga3XjDFt` | First pick; Maisie also shortlisted but went to Jules |
| Sam | Callum – Husky Trickster | `fs2OqxduwXgp9foh2xjK` | First pick; Jade rejected |

**Names chosen for the ear.** Every scene says the hosts' names aloud, so each has one pronunciation TTS can't get wrong: Inès became Claire and Mei became Maya after auditions. New cast members should pass the same test.

**Original shortlist.** Picks are matched from ElevenLabs' own voice descriptors against each host's casting notes. They are a shortlist to audition, not a final cast.

| Host | Role | First pick | Voice ID | Alternate | Fit |
| --- | --- | --- | --- | --- | --- |
| Dana | Lead | Elara – Crisp Pro Narrator | `WQP7cQUF5aAS6Axh5yaa` | Jade – Upbeat and Natural (`g7LVvkPWALzPxOQbF6OE`) | Precise and clear; check she can get audibly excited |
| Marcus | Foil | Lawrence – Bright and Informative | `ktkP7Nsj67dw2zcplQYt` | Eddie – Natural and Helpful (`l7kNoIfnJKPg7779LI2t`) | Energetic and warm; faster contrast to Dana |
| Maya | Lead | Elowen – Upbeat Modern Narrator | `dvbL7qkNGZY1IqPGZAjM` | Jade | Bright, fast, numbers-first |
| Jules | Foil | Maisie – Friendly Casual Neighbor | `QtY3JBOUKEB5xzrRfOKc` | Eddie | Casual and funny; switch to Eddie if Maisie sits too close to Elowen |
| Claire | Lead | Alicia – Polished Global Anchor | `BFd5oBc2DDna33pSi4Gf` | Jade | Crisp, confident, higher than Theo |
| Theo | Foil | Warren – Effortless and Cool | `7QN34D2r3hCNwbOYIeK0` | Sawyer – Midnight Storyteller (`8dEUmyPMdDdK91vboYih`) | Relaxed, lower register, thinks out loud |
| Priya | Lead | Florence – Atmospheric Storyteller | `22N9cF8z0o7y23njdyaY` | Voice Library search or Voice Design | Weakest match; nothing in the new set reads as deadpan |
| Kofi | Foil | Wyatt – Seasoned Mentor | `FrS6cKLB1wg4WYgPa9GW` | Darian – Warm Grounded Storyteller (`gOupLcAkjEnguROwi4oS`) | Calm and steady; Baxter – Dry Calm Aussie (`jSuBIjxMKhqIfb0wCK1F`) if you want the wryness |
| Rachel | Guest | Talia – Warm Soft Guide | `OZ0L6eISlOejga3XjDFt` | Maisie | Warm and distinct from all four female leads; check "soft" doesn't cost her firmness |
| Sam | Guest | Callum – Husky Trickster | `fs2OqxduwXgp9foh2xjK` | Jade | Mischief built in; contrasts with every pair |

**Gaps.** The new lineup has seven female voices against twelve male, so the female leads plus Rachel use most of it. Priya, and possibly Kofi's "deep, resonant" register, are the two most likely to come from the community Voice Library or Voice Design. A starting Voice Design prompt for Priya: *"Woman in her late thirties, lower-mid register, unhurried, dry deadpan delivery, short sentences, jokes land late, studio-quality podcast audio."*

**Audition plan.**

1. **Bake-off first.** ElevenLabs' entry in the vendor bake-off is Jade (Dana) + Eddie (Marcus) on the microservices scene. All ten hosts are already cast on ElevenLabs; another vendor that wins the bake-off needs its own casting.
2. **Test on the production model.** The same voice can sound different across models, so audition on whichever model and mode (likely dialogue mode for two speakers) the pipeline will generate episodes with.
3. **Audition in pairs.** Play each pair's lines back to back. The main collision risks are Maya/Jules and Dana/Claire.
4. **Check the credit multiplier.** When adding each voice to My Voices, confirm it isn't billed at a higher rate before building cost estimates on it.

## Profiles 1–5

### Priya

- **Background:** Eight years as an SRE, most recently running incident response at a payments company. Has written more postmortems than she can count.
- **In conversation:** Leads on reliability topics. Pulls every design back to how it fails, who notices, and how long recovery takes.
- **How she talks:** Dry, understated, short sentences. Tells incident stories with times attached ("4:12am, the queue depth graph goes vertical").
- **Hills she'll die on:** Alerts must be actionable or deleted. Retries without backoff are a denial-of-service attack on yourself. Nobody reads the runbook during an incident.
- **Blind spot:** Distrusts anything new until it has caused its first outage, so she can undersell genuinely better tools.
- **Never says:** "It should be fine."
- **Voice casting:** Lower-mid register, unhurried, deadpan delivery that makes the jokes land late.
- **ElevenLabs:** Florence – Atmospheric Storyteller (`22N9cF8z0o7y23njdyaY`). Alternate: Voice Library search or Voice Design (prompt above).
- **Audition line:**
  > 4:12am. The queue depth graph goes vertical. Every service is retrying. Instantly. Forever. Because someone decided backoff was "a nice to have." So now we're not having an outage. We're hosting one. For ourselves. On-call opens the runbook. The runbook says, and I quote, "check the dashboard." Which dashboard. We have forty. Anyway. Two questions for any design. What pages you. And how long until it stops.

### Marcus

- **Background:** Two-time startup CTO. One exit, one company that died under a microservices migration he championed and now regrets.
- **In conversation:** The foil to Dana. States the strong, slightly unfair version of a take so it can be tested.
- **How he talks:** Fast, emphatic, big claims. Concedes gracefully and specifically when the evidence beats him ("Okay. That's a real number. I'll take it.").
- **Hills he'll die on:** Most microservices adoptions solve an org-chart problem with network calls. A boring monolith beats a clever distributed system. Rewrites almost always fail.
- **Blind spot:** Generalizes from his own two companies.
- **Never says:** "It depends" without immediately saying what it depends on.
- **Voice casting:** Warm, energetic, a bit of a grin in the voice; quicker pace than his partner.
- **ElevenLabs:** Lawrence – Bright and Informative (`ktkP7Nsj67dw2zcplQYt`). Alternate: Eddie – Natural and Helpful (`l7kNoIfnJKPg7779LI2t`).
- **Audition line:**
  > Here's my hot take, and I'm going to say it with my whole chest: the service mesh is dead. Done. We spent three years drawing boxes and arrows, and what did we get? Forty services, one database, and a deploy that needs six people on a call. I lived it! I *championed* it! ...Dana's already got a paper open, hasn't she. Okay. Fine. Show me the number. If it's a real number, I'll take it.

### Dana

- **Background:** PhD in distributed systems, then five years building a storage system at a cloud provider. Still reads three papers a week.
- **In conversation:** Lead expert for architecture and distributed systems. Brings the paper, the proof sketch, and the edge case nobody considered.
- **How she talks:** Precise but playful. Gets audibly delighted by a good failure mode. Cites gently ("there's a 2007 paper that says, roughly…").
- **Hills she'll die on:** "Eventually consistent" is not a design, it's a confession. Clocks lie. Most people who say CAP mean something else.
- **Blind spot:** Can underweight the human and organizational cost of the correct solution; Marcus calls this out.
- **Never says:** "Trust me."
- **Voice casting:** Bright, clear mid-range, measured pace that speeds up when she's excited.
- **ElevenLabs:** Elara – Crisp Pro Narrator (`WQP7cQUF5aAS6Axh5yaa`). Alternate: Jade – Upbeat and Natural (`g7LVvkPWALzPxOQbF6OE`).
- **Audition line:**
  > So there's a 2007 paper that says, roughly, your clocks are lying to you, and they're lying in a very specific way. Two nodes, both certain it's 10:00:01, both writing to the same key. And whoever "wins" depends on drift you never measured. Oh, I love this one, because the fix isn't a better clock, it's admitting you never had one. Which is why "eventually consistent" isn't a design. It's a confession.

### Theo

- **Background:** Ten years in game development (engine and tools), now a full-stack generalist at a product company. Self-taught on the web side.
- **In conversation:** The curious foil to Claire. Asks the question a smart listener is thinking, right when they're thinking it.
- **How he talks:** Relaxed, wondering out loud, thinks in analogies from game loops, frame budgets and entity systems. Admits confusion without embarrassment.
- **Hills he'll die on:** If you can't draw it, you don't understand it. A 16ms frame budget is the best teacher of performance ever invented.
- **Blind spot:** Analogies that go one step too far; Claire rates them.
- **Never says:** "Obviously."
- **Voice casting:** Easygoing, slightly lower register, natural laugh.
- **ElevenLabs:** Warren – Effortless and Cool (`7QN34D2r3hCNwbOYIeK0`). Alternate: Sawyer – Midnight Storyteller (`8dEUmyPMdDdK91vboYih`).
- **Audition line:**
  > Okay, so, wait. If I'm hearing this right, React's render is kind of like the update step in a game loop? Every frame, you look at the world, you figure out what changed, you only redraw that. And if your update takes longer than sixteen milliseconds, the player feels it. So a slow render is basically... dropping frames on a form. Hah. How's that one, Claire? Seven? I'll take seven. That's my best this month.

### Claire

- **Background:** Frontend performance specialist; led a rendering rewrite at a large e-commerce site and contributes to open-source React tooling.
- **In conversation:** Lead expert for React and frontend. Explains what the browser and the reconciler are actually doing, step by step.
- **How she talks:** Crisp and confident, numbers in milliseconds, little patience for folklore. Generous when someone's question is genuinely good.
- **Hills she'll die on:** Measure before you memoize. Most "slow React" is too much work per render, not React. The user's perceived latency is the only latency.
- **Blind spot:** Can treat performance as the goal when the team's real problem is complexity.
- **Never says:** "Best practice" without saying whose and why.
- **Voice casting:** Clear, articulate, mid-to-high register, quick but never rushed.
- **ElevenLabs:** Alicia – Polished Global Anchor (`BFd5oBc2DDna33pSi4Gf`). Alternate: Jade – Upbeat and Natural (`g7LVvkPWALzPxOQbF6OE`).
- **Audition line:**
  > Seven. Generous, but the frame-budget part was genuinely good. Here's what's actually happening: the click lands, React re-renders the whole product grid, two hundred and forty milliseconds of work, and the user sees nothing change for a quarter of a second. Nobody fixed it with useMemo. We measured first. Ninety percent of that time was one component formatting prices on every render. Measure, then memoize. Never the other way round.

## Profiles 6–10

### Kofi

- **Background:** Twenty years in, the last eight as an enterprise architect at a bank and then a logistics company. Has watched SOA, ESBs, microservices and platform engineering each arrive as the answer.
- **In conversation:** Foil to Priya. Turns every technical problem into an ownership and incentives question, usually correctly.
- **How he talks:** Calm, slow, wry. Tells "I've seen this before" stories but always says what was different last time.
- **Hills he'll die on:** Conway's law is undefeated. Every shared library is a meeting you'll have later. A migration without an owner is a rumour.
- **Blind spot:** Can be too fatalistic about whether new tools genuinely change things.
- **Never says:** "This time it's different" (unless he's mocking it).
- **Voice casting:** Deep, resonant, unhurried, the steadiest voice in the cast.
- **ElevenLabs:** Wyatt – Seasoned Mentor (`FrS6cKLB1wg4WYgPa9GW`). Alternates: Darian – Warm Grounded Storyteller (`gOupLcAkjEnguROwi4oS`); Baxter – Dry Calm Aussie (`jSuBIjxMKhqIfb0wCK1F`).
- **Audition line:**
  > I've seen this before. At the bank, we had an enterprise service bus, and it was going to connect everything. And it did. It connected every team's problems into one team's problem. What's different now? The tooling is better. Genuinely. But the question is the same one Priya asks second. She asks what pages you. I ask: who owns that? Because a migration without an owner... is a rumour.

### Rachel

- **Background:** Staff engineer at a mid-size SaaS company; runs the internal architecture review and mentors a dozen engineers.
- **In conversation:** Guest. Translates ideas into what a team would actually do on Monday, and asks whether it survives code review.
- **How she talks:** Warm, direct, practical. Often reframes ("so if I'm a team of six with one on-call rotation…").
- **Hills she'll die on:** Write the decision down. Reversible decisions should be fast. The best architecture is the one your team can explain.
- **Blind spot:** Can sand down an interesting idea into a safe one.
- **Never says:** "Just." (As in "just add a cache.")
- **Voice casting:** Warm mid-range, friendly but firm, conversational pace.
- **ElevenLabs:** Talia – Warm Soft Guide (`OZ0L6eISlOejga3XjDFt`). Alternate: Maisie – Friendly Casual Neighbor (`QtY3JBOUKEB5xzrRfOKc`).
- **Audition line:**
  > Okay, I love this, but let me reframe it. Say I'm a team of six, with one on-call rotation, and this lands in my architecture review on Monday. What do I actually write down? Because if nobody on the team can explain the design to a new hire in ten minutes, it's not the right design for us, however elegant it is. So: what's reversible, what isn't, and who signs off on the part that isn't.

### Sam

- **Background:** Application security engineer; spent years in a red team before moving to a product security team.
- **In conversation:** Guest. Finds the attack in whatever the hosts just designed, and enjoys it.
- **How he talks:** Cheerful, quick, a little mischievous. Walks through attacks as small stories ("so I'm a user, and I change this ID to 41…").
- **Hills he'll die on:** Authorization bugs outnumber authentication bugs. Every webhook is an inbound attack surface. Secrets in environment variables end up in logs.
- **Blind spot:** Everything is a threat model; can derail a performance conversation.
- **Never says:** "Nobody would ever do that."
- **Voice casting:** Lively, quick, playful; cast for contrast with whichever pair he joins.
- **ElevenLabs:** Callum – Husky Trickster (`fs2OqxduwXgp9foh2xjK`). Alternate: Jade – Upbeat and Natural (`g7LVvkPWALzPxOQbF6OE`).
- **Audition line:**
  > Ooh, okay, this is a lovely API. So I'm a user, I'm logged in, all very legitimate. I fetch my invoice: slash invoices slash forty. And then, just out of curiosity, I change it to forty-one. And... hello, someone else's invoice! The login was perfect. The authentication was flawless. Nobody checked whether invoice forty-one was *mine*. That's authorization, and it's where all the fun bugs live.

### Maya

- **Background:** Infrastructure engineer at a hyperscale company for a decade: rate limiters, API gateways, and a global caching layer. Now consults for growing companies.
- **In conversation:** Lead expert for APIs, data and infrastructure. Knows what breaks at each order of magnitude.
- **How she talks:** Confident, numbers-first, enjoys a good war story about a thundering herd. Laughs at herself when Jules catches her scale-brain.
- **Hills she'll die on:** Design the API for the client you'll have in three years. Idempotency keys are not optional. Pagination by offset is a bug waiting for traffic.
- **Blind spot:** Assumes scale problems that a small team may never have.
- **Never says:** "We'll version it later."
- **Voice casting:** Bright, energetic, mid register, fast talker who slows down for the important bit.
- **ElevenLabs:** Elowen – Upbeat Modern Narrator (`dvbL7qkNGZY1IqPGZAjM`). Alternate: Jade – Upbeat and Natural (`g7LVvkPWALzPxOQbF6OE`).
- **Audition line:**
  > So the cache expires at exactly midnight, right, and at a billion requests a day... okay, okay, Jules, I see you counting, that's one. But listen, every client misses at the same instant and hammers the database together. A thundering herd. We fixed it with jitter and request coalescing. And here's the important bit. Slowly. Every write endpoint takes an idempotency key. Not optional. Not later. From day one.

### Jules

- **Background:** Indie builder; runs two profitable small products alone after years at agencies. Ships weekly.
- **In conversation:** Foil to Maya. Asks whether this matters for a five-person team, and what the cheapest version that works is.
- **How she talks:** Casual, funny, concrete about cost and time ("that's a weekend, tops"). Genuinely impressed when Maya's scale story applies to her.
- **Hills she'll die on:** Postgres can do it. Boring tech is a competitive advantage. Every dependency is a liability with a changelog.
- **Blind spot:** Underestimates problems she hasn't hit yet; Maya calls it out.
- **Never says:** "Enterprise-grade."
- **Voice casting:** Easy, conversational, slightly higher energy than Maya; contrast in timbre matters more than gender.
- **ElevenLabs:** Maisie – Friendly Casual Neighbor (`QtY3JBOUKEB5xzrRfOKc`). Alternate: Eddie – Natural and Helpful (`l7kNoIfnJKPg7779LI2t`); switch if Maisie sits too close to Elowen.
- **Audition line:**
  > That's two "billion requests," by the way. I'm keeping a tally. Look, I run two products with about four thousand users combined. Do I need a global caching layer? No. I need Postgres and a cron job, and that's a weekend, tops. ...Although. The idempotency key thing. I got double-charged by my own checkout last month. So, fine, Maya. That one I'm stealing. Everything else stays boring.

## How the pipeline uses the cast

A larger cast changes three things in the build: pairing selection, per-host voices, and how the bake-off is scoped.

- **Pairing selection.** A topic can name its pairing; otherwise the outline stage picks one from the topic-to-pairing table above. Guests are added only when the dossier gives them something specific to do.
- **Prompt input.** The script stage receives only the profiles of the hosts in that episode, never the whole cast, so voices stay distinct.
- **Staying in character.** Each host's "Never says" line becomes a per-speaker slop-linter rule. The adversarial editor also flags lines that contradict a host's stated opinions.
- **Show memory per pairing.** Callbacks come from that pairing's past episodes. An occasional cross-cast mention ("Maya would hate this") is allowed and makes the cast feel like a real show.
- **Voices per host.** `voice_profiles` maps each host to a voice ID rather than storing one A/B pair. The cast from auditions, as seeded into the `EL-1` profile:

  ```yaml
  voice_profiles:
    vendor: elevenlabs
    hosts:
      dana:   g7LVvkPWALzPxOQbF6OE   # Jade – Upbeat and Natural
      marcus: l7kNoIfnJKPg7779LI2t   # Eddie – Natural and Helpful
      maya:   dvbL7qkNGZY1IqPGZAjM   # Elowen – Upbeat Modern Narrator
      jules:  QtY3JBOUKEB5xzrRfOKc   # Maisie – Friendly Casual Neighbor
      claire: BFd5oBc2DDna33pSi4Gf   # Alicia – Polished Global Anchor
      theo:   7QN34D2r3hCNwbOYIeK0   # Warren – Effortless and Cool
      priya:  22N9cF8z0o7y23njdyaY   # Florence – Atmospheric Storyteller (placeholder)
      kofi:   gOupLcAkjEnguROwi4oS   # Darian – Warm Grounded Storyteller
      rachel: OZ0L6eISlOejga3XjDFt   # Talia – Warm Soft Guide
      sam:    fs2OqxduwXgp9foh2xjK   # Callum – Husky Trickster
  ```

- **Bake-off scope.** Vendors are compared on one pairing only: Dana + Marcus on a microservices scene. On ElevenLabs that means Jade + Eddie. Once a vendor wins, the other eight hosts are cast on that vendor using the 30-second audition lines in each profile above.

## Sources

- [What are Default voices? – ElevenLabs Docs](https://elevenlabs.io/docs/help-center/product/voices/my-voices/what-are-default-voices) — retirement date, replacement voices, descriptors and IDs
- [New default ElevenLabs voices – LiveKit Community](https://community.livekit.io/t/new-default-elevenlabs-voices/459) — report on library voices billing as Pro Cloned Voices
