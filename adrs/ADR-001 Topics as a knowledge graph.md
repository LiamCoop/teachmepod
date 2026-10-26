# ADR-001: Topics as a knowledge graph

Oct 25, 2026 · @Liam

**Status:** Proposed. **Relates to:** Doc, without changing it. Adopting this ADR adds tables and steps to milestones M1, M3 and M4.

## Context

The plan treats each backlog topic as an independent string, so episodes can't build on each other. Technical topics have structure: concepts depend on other concepts, and an episode is far better when it knows what the listener has already heard.

- Without structure, every episode re-explains its foundations, or assumes ones the listener never got.
- There's no way to see which parts of a field are covered and which are gaps.
- The long-term idea of "streams of knowledge" other people could follow needs a shared map of concepts, not one person's list of topics.

## Decision

Model topics as **concepts in a directed prerequisite graph**, browsed through a tree of `part_of` links, with episodes attached to the concepts they cover.

- **Concepts** are stable, reusable nodes, independent of any episode.
- **Edges** come in three kinds only: `prerequisite` (learn A before B), `part_of` (the tree view), `related` (lateral).
- **Episodes** cover concepts many-to-many, with a role: introduces, deepens or mentions.
- **Backlog topics** target a concept. Research proposes new concepts as it goes; you approve the ones worth an episode into the backlog, so the graph grows ahead of the backlog and feeds it.

&#91;embedded content: example · one branch of the graph, two episodes\]

The tree is how you browse; the prerequisite edges, which cross branches, are what let episodes build on each other.

## Data model

Four new Postgres tables; nothing in the plan's existing tables is removed. Graph queries use recursive CTEs, and pgvector handles near-duplicate detection, so no graph database is needed.

| Table | Purpose | Key columns |
| --- | --- | --- |
| `concepts` | Global, user-independent nodes | slug, name, summary, embedding (vector), status (proposed, approved, rejected, merged), merged\_into, proposed\_by\_episode\_id, proposal\_reason |
| `concept_edges` | Typed links between concepts | src\_id, dst\_id, kind (prerequisite, part\_of, related), source (pipeline or manual), status (proposed, approved, rejected), reason, proposed\_by\_episode\_id |
| `episode_concepts` | What each episode covers | episode\_id, concept\_id, role (introduces, deepens, mentions) |
| `user_concepts` | What a given person knows | user\_id, concept\_id, state (unknown, heard, known), via\_episode\_id |

- `topics` gains an optional `target_concept_id`: a backlog item becomes a request to fill part of the graph.
- Constraints: no self-edges; prerequisite edges must stay acyclic, checked on insert with a recursive query.
- `part_of` gives each concept at most one primary parent, which is what makes the tree view possible.
- Acyclicity and the one-parent rule apply to **approved** edges only. Research will routinely propose competing parents and edges for the same concept across episodes; the queue needs to show them side by side, and an unreviewed edge must not block a correct one. Both are rechecked when an edge is approved.

## How the graph grows

The research stage proposes nodes and edges; a reconciliation step prevents duplicates; you approve changes until the matching earns trust. Deduplication is the hard part of this ADR.

1. **Extraction.** The dossier already ranks concepts. Research now also emits, per concept, its likely prerequisites and parent. It also emits the adjacent concepts it ran into but that are outside this episode (a related technique, a deeper subtopic, a prerequisite the topic assumes), each with its edges to this episode's concepts and a one-line reason it's worth an episode. These replace the dossier's free-text follow-up suggestions.
2. **Reconciliation.** Each proposed concept is compared with existing ones:
   - embedding similarity finds candidates ("idempotent APIs" vs. "idempotency keys")
   - a model judges same, parent/child, or distinct, with a short reason
   - matches attach to the existing node; the rest become `proposed`
3. **Approval queue.** A site page lists proposed nodes, edges and merges, each with the episode that proposed it and why. Actions: approve (into the graph only), approve and queue (into the graph, plus a backlog topic targeting it, with angle and depth set there), merge-into, or reject. A merge keeps the old slug as `merged_into` so links never break. This is now the main way the backlog grows.
4. **Feedback.** "Go deeper on X" creates a proposed child under X and a backlog topic targeting it.

Once your rejection rate on a kind of change is low, that kind can auto-approve.

## How episodes build on each other

The graph pays off in the audio itself: scripts skip what you know, call back to where you learned it, and flag what's missing.

- **Known prerequisites.** The outline and script prompts receive the target concept's prerequisites and your `user_concepts` state for each. Known ones get a callback, not an explanation ("Dana walked through Raft a few weeks ago"), and the show notes link that episode.
- **Missing prerequisites.** For an uncovered prerequisite the outline either budgets a two-minute explainer, or the pick stage defers the topic and proposes the prerequisite first.
- **Prerequisite-aware picking.** Among queued topics, prefer those whose prerequisites are already covered, then fall back to priority.
- **Gap suggestions.** Uncovered concepts that sit underneath covered ones surface in the approval queue as candidates to queue. This is how the tree fills itself in.
- **Marking knowledge.** Publishing an episode marks its introduced concepts as `heard` for you. Feedback can promote one to `known` or flag it as needing another pass.
- **Tree view.** The site shows the `part_of` hierarchy as a collapsible list: each concept with its covering episodes, and gaps greyed out.

## Streams and the multi-user boundary

Streams are far off, but one boundary decided now keeps them cheap: concepts, edges and episodes are global; knowledge is per user.

- **A stream** is a goal concept plus an ordered path to it through the prerequisite graph, starting from what a listener already knows, with episodes attached. The roadmap-driven product idea is a stream with a schedule.
- **Computing one** is a topological sort of the goal's prerequisites minus the listener's known concepts. Curated streams can pin or reorder steps by hand.
- **The boundary:** nothing about one person lives on `concepts` or `episodes`. Your baseline, what you've heard, and what you know live in `user_concepts`, keyed by user from day one, even while there is only one user.
- **Out of scope here:** stream tables, sharing, and per-user episode variants. These get their own ADR when they're real.

## Alternatives considered

| Option | Why not |
| --- | --- |
| A strict tree of topics | Concepts have several prerequisites (sagas need both distributed transactions and event-driven design); a tree forces one and loses the rest |
| Tags on topics | Cheap, but tags carry no order, so no prerequisites, no gap detection and no streams |
| A graph database (Neo4j and similar) | A second datastore to run and back up; graphs of a few thousand nodes are well within Postgres recursive queries |
| Let the model build the graph unsupervised | Duplicates and wrong prerequisites compound over time; an approval queue is cheap at one user's volume |
| Defer it until after v1 | Retrofitting concepts onto published episodes means re-extracting from transcripts; capturing them from the start costs one extra step in research |

## Consequences and rollout

Adopting this adds one research output, one reconciliation step and one review page, and makes every episode after the first few noticeably less repetitive.

**Costs**

- An extra model call per proposed concept for reconciliation, plus embeddings.
- A few minutes of your time per episode in the approval queue. This doesn't fall away with trust as fast as dedup does: choosing what goes into the backlog stays your call.
- Prompts grow slightly with prerequisite context.

**Rollout against the plan's milestones**

| Milestone | Adds |
| --- | --- |
| M1 · Text pipeline | The four tables; concept and edge extraction in research; reconciliation; the approval queue page |
| M3 · First full episode | Tree view on the site; prerequisite context and callbacks in scripts; `heard` marking on publish |
| M4 · Automation | Prerequisite-aware picking; gap suggestions into the approval queue; auto-approve for trusted change kinds |
| Later | Visual graph view; streams (separate ADR) |

**Open questions**

- [ ] Seed the graph by hand for the first three test topics, or let the first episodes build it?
- [ ] Should `related` edges be allowed to cross into callbacks, or prerequisites only?
