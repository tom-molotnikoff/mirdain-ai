# Ubiquitous Language

> Semantics-only. No implementation details, file paths, or technology names.
> When a concept has multiple names in use, this file is the tiebreaker.

---

## Actors

| Term                | Definition                                                                                      | Aliases to avoid              |
| ------------------- | ----------------------------------------------------------------------------------------------- | ----------------------------- |
| **Operator**        | A person who installs and configures a mirdain instance for a team or themselves                | Admin, sysadmin               |
| **Engineer**        | A person who uses mirdain to guide work through the workflow                                    | User, developer, dev          |
| **Agent**           | An autonomous AI actor that executes a skill within a run                                       | Bot, AI, assistant, LLM       |
| **Bot Identity**    | The non-human account identity through which mirdain performs all writes on behalf of the agent | Bot, service account, mirdain |

> An **Engineer** interacts with mirdain through the UI; an **Agent** interacts through tools. They are distinct actors even when they collaborate on the same issue.

---

## Workflow

| Term              | Definition                                                                                              | Aliases to avoid             |
| ----------------- | ------------------------------------------------------------------------------------------------------- | ---------------------------- |
| **Workflow**      | The full staged pipeline from raw idea to merged pull request                                           | Pipeline, process, lifecycle |
| **Phase**         | A named, discrete step in the workflow (design, PRD, planned, in-review, done, bug-triage)              | Stage, step, state, status   |
| **Gate**          | A decision point between two consecutive phases; either `human` (blocks until approved) or `auto`      | Transition, checkpoint       |
| **Feature Track** | The workflow path for new features: design → PRD → planned → in-review → done                          | Feature flow, feature path   |
| **Bug Track**     | The workflow path for defects: bug-triage → planned → in-review → done                                 | Bug flow, hotfix path        |
| **Direct Entry**  | The shortcut path where an issue enters at `planned` without passing through design or PRD phases       | Fast track, bypass           |

---

## Tracker & Issues

| Term                  | Definition                                                                                                      | Aliases to avoid                    |
| --------------------- | --------------------------------------------------------------------------------------------------------------- | ----------------------------------- |
| **Tracker**           | The external system of record for all durable workflow state: issues, labels, comments, and pull requests       | Issue tracker, GitHub, ticket board |
| **Issue**             | A single unit of work in the tracker; the primary contract between engineer and agent                           | Ticket, task, card, story           |
| **Pull Request**      | The code-change proposal produced by the agent during the `in-review` phase                                     | PR, diff, patch                     |
| **Label**             | A tag applied to an issue; mirdain uses two orthogonal label namespaces: **Workflow Label** and **Status Flag** | Tag, attribute                      |
| **Workflow Label**    | A label encoding which phase an issue is currently in; mutually exclusive — exactly one per active issue        | Phase label, state label            |
| **Status Flag**       | A label encoding runtime health or operational status; orthogonal to the workflow label; multiple may coexist   | Runtime label, flag label           |
| **Tenancy Marker**    | The single label that opts an issue into mirdain management; the reconciler ignores any issue lacking it        | Opt-in label, mirdain label         |
| **Eligible Issue**    | An issue that the reconciler considers actionable: carries the tenancy marker, has a workflow label, and no blocking status flag | Actionable issue, ready issue |

---

## Agent Execution

| Term            | Definition                                                                                                                          | Aliases to avoid                         |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| **Run**         | A single invocation of an agent to execute a skill against a specific issue                                                         | Session, job, task, execution            |
| **Run Mode**    | Whether a run is interactive (HITL) or autonomous (AFK)                                                                             | Run type, mode                           |
| **HITL Run**    | A run where the engineer is in live conversation with the agent via the UI; the agent pauses at `awaiting_input` events             | Interactive run, live run, chat run      |
| **AFK Run**     | A run where the agent acts autonomously without human interaction; it runs to completion or failure without pausing                 | Autonomous run, background run, headless |
| **Skill**       | A markdown document with structured frontmatter declaring an agent's objective, required tools, and model class for a workflow phase | Prompt, system prompt, agent script      |
| **Artifact**    | A named, structured output produced by an agent during a run (e.g. a PRD document, an issue decomposition)                         | Output, result, deliverable              |
| **Run Config**  | The set of parameters passed to an agent at the start of a run (which issue, which skill, which repo, secrets)                     | Agent config, launch params              |
| **Model Class** | An abstract LLM tier (`small`, `medium`, `large`) declared by a skill; resolved to a concrete model by the operator config          | Model size, LLM tier, model type         |

---

## PR Feedback Loop

| Term                | Definition                                                                                                          | Aliases to avoid                 |
| ------------------- | ------------------------------------------------------------------------------------------------------------------- | -------------------------------- |
| **Feedback Loop**   | The automated cycle in which agents react to pull request events, iterate on code, and drive the PR to merge        | PR loop, review loop, CI loop    |
| **Wake Event**      | Any event that triggers a new agent invocation in the feedback loop (comment, review, CI status change, branch push) | Trigger, notification, hook      |
| **Failure Budget**  | The maximum number of feedback-loop iterations allowed on a PR; when exhausted, the issue is escalated to a human  | Iteration cap, retry limit       |
| **Debounce Window** | A short configurable time window within which multiple wake events are coalesced into a single agent invocation     | Burst window, cooldown, throttle |

---

## Infrastructure

| Term            | Definition                                                                                                           | Aliases to avoid               |
| --------------- | -------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| **AgentRunner** | The component responsible for launching and terminating agent containers on behalf of the orchestrator               | Executor, scheduler, launcher  |
| **Container**   | The isolated, ephemeral runtime environment in which an agent executes a run                                         | Sandbox, VM, process           |
| **Workspace**   | A persistent storage volume mounted into a container, holding the repository clone and the agent's working files     | Working directory, volume, dir |

---

## Brokered Writes & Tools

| Term               | Definition                                                                                                                    | Aliases to avoid                   |
| ------------------ | ----------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| **Brokered Write** | Any write operation that an agent requests through the orchestrator rather than performing directly against the external system | Direct write, agent write          |
| **Tool**           | A named, typed operation that an agent may invoke during a run; write tools are always brokered                               | Function, action, command, API     |
| **Tool Surface**   | The complete set of tools available to an agent for a given skill, declared in the skill's frontmatter and validated at load  | Tool set, API surface, tool schema |

---

## Configuration & Capabilities

| Term                | Definition                                                                                           | Aliases to avoid                    |
| ------------------- | ---------------------------------------------------------------------------------------------------- | ----------------------------------- |
| **Operator Config** | The global configuration file on the host machine; owned and edited by the operator                  | Global config, server config        |
| **Repo Config**     | The per-repository configuration file committed to the target repository; owned by the engineering team | Project config, local config      |
| **Cap**             | A configurable hard limit on a run's resource consumption (wall-clock time, PR iteration count)       | Limit, budget, quota                |
| **Polling Interval** | The frequency at which the reconciler checks the tracker for issues eligible for action              | Check interval, poll rate, interval |

---

## System Roles (at concept level)

| Term             | Definition                                                                                                                        | Aliases to avoid               |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------------------------------ |
| **Orchestrator** | The central mirdain process; manages the workflow, brokers all writes, and coordinates agents across repositories                  | Server, backend, controller    |
| **Reconciler**   | The component within the orchestrator that continuously polls the tracker and initiates or queues runs for eligible issues        | Poller, watcher, scheduler     |
| **Bridge**       | The component running inside an agent container that mediates between the agent harness and the orchestrator's WebSocket protocol | Adapter, plugin, shim, harness |

---

## Relationships

- A **Workflow** consists of one or more ordered **Phases** connected by **Gates**.
- An **Issue** is in exactly one **Phase** at a time, indicated by its **Workflow Label**.
- An **Issue** carries exactly one **Tenancy Marker** and zero or more **Status Flags**.
- An **Eligible Issue** produces at most one active **Run** at a time.
- A **Run** executes exactly one **Skill** against exactly one **Issue**.
- A **Run** is either a **HITL Run** or an **AFK Run** — the **Run Mode** is determined by the phase and skill.
- An **Agent** may produce one or more **Artifacts** within a single **Run**.
- A **Brokered Write** is always initiated by an **Agent** invoking a **Tool** and fulfilled by the **Orchestrator**.
- A **Pull Request** is associated with exactly one **Issue**; it is the entry condition for the `in-review` phase.
- **Wake Events** on a **Pull Request** trigger new **Runs** within the **Feedback Loop**.
- Each **Feedback Loop** invocation consumes one unit of the **Failure Budget**.

---

## Example dialogue

> **Engineer:** "I filed a new issue — when will the **agent** pick it up?"
>
> **Domain expert:** "Only once it's an **eligible issue**: it needs the **tenancy marker** plus a **workflow label** showing which **phase** it's in. Without both, the **reconciler** ignores it entirely."
>
> **Engineer:** "I added both. The agent started a **run** and produced a PRD **artifact**. What happens next?"
>
> **Domain expert:** "The PRD phase ends at a **gate**. If that gate is set to `human`, the issue stays in the `prd` **phase** until you approve it and advance the **workflow label** to `planned`. If it's `auto`, the orchestrator advances it immediately."
>
> **Engineer:** "Once it's `planned`, what does the agent do?"
>
> **Domain expert:** "The **reconciler** picks it up and starts an **AFK run** — the agent opens a **pull request** autonomously. From that point on, the **feedback loop** takes over: every **wake event** on the PR (a CI failure, a review comment, a new push) triggers a fresh **run** until the PR merges or the **failure budget** runs out."
>
> **Engineer:** "What if the agent gets stuck?"
>
> **Domain expert:** "When the **failure budget** is exhausted, the orchestrator sets the `needs-human` **status flag** on the issue. That blocks any further automatic **runs** until you clear the flag."

---

## Flagged ambiguities

- **"Stage" vs "Phase"**: the constitution's prose uses both interchangeably. **Phase** is canonical. "Stage" appears only in repo-config YAML keys (`stages:`) as a config concept, not a domain term — do not use it when describing the workflow to a domain expert.
- **"Agent" overload**: "agent" is used to mean (a) the AI actor persona and (b) the running container. The canonical split is: **Agent** = the AI actor; **Run** = a single execution; **Container** = the execution environment. Never say "the agent is running" when you mean "there is an active run."
- **"Bot"**: used in the constitution to mean the **Bot Identity** (the GitHub account). Do not conflate with **Agent** (the AI actor). The bot identity is how writes appear externally; the agent is the decision-maker internally.
- **"Skill" vs "phase" mapping**: configs map phases to skills (`stages: design: brainstorming`). This is a configuration concern. In domain conversation, say "the **brainstorming** skill handles the design phase", not "the design stage runs the brainstorming prompt."
- **"Label" ambiguity**: "label" alone is ambiguous. Always qualify: **workflow label**, **status flag**, or **tenancy marker**. Reserve bare "label" only when the namespace is already established from context.
- **"Tracker" vs specific backend**: **Tracker** is the domain concept. The backing technology is an implementation detail. Domain conversations should never need to name the specific tracker system.
