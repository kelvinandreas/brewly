# Vibe Engineering — Building Brewly in a Few Hours

## Why I built this

I kept hearing the same things from people in the AI tooling space:

> _"You have to try Claude Code."_
> _"Write your CLAUDE.md properly and it changes everything."_
> _"You should try PRPs — they make maintaining context across sessions much easier."_

I was curious and a little skeptical. So I gave myself a challenge: build a complete POS system for a small café — blank repo to working app — using AI-assisted development as the primary workflow.

The goal wasn't to ship a polished product. It was to answer one question: **can a well-configured AI coding workflow build a reasonably complex app across multiple sessions while staying architecturally consistent?**

Four milestones later — auth, menu/table management, orders/payments/KDS, and song requests with reporting — I had a working app with a Clean Architecture Go backend, React frontend, PostgreSQL, SSE for realtime updates, and Docker Compose deployment.

Most of the functionality got built. But after real testing, it became clear that functional completeness and production readiness aren't the same thing. That turned out to be the most valuable lesson from the whole experiment.

---

## What "vibe engineering" actually means

Using an AI coding assistant as a genuine co-engineer, not a glorified autocomplete.

Instead of asking for isolated functions, you build an environment where the AI understands the project architecture, coding standards, business rules, and existing patterns. The goal is consistent contribution across multiple sessions — planning features, generating implementations, writing tests, refactoring — without you repeating yourself every time.

**The quality of the output depends heavily on the quality of the context.** Poor context produces inconsistent code. Well-maintained context produces surprisingly coherent implementations.

---

## The setup

### `CLAUDE.md` — the operating manual

The most important file. Claude Code reads this at the start of every session. It contains the stack, architecture rules, coding standards, and explicit anti-patterns — things like "no business logic in handlers," "no direct database access outside repositories," "no tokens in localStorage."

Without this, architectural drift starts surprisingly quickly.

### `.claude/rules/` — context-aware rule injection

Rules load automatically based on which files are open. Backend rules for `.go` files, frontend rules for `.tsx`, migration rules for migration files. Keeps guidance relevant without wasting context.

### `.claude/skills/` — reusable task templates

Skills encode recurring patterns: clean architecture scaffolding, GORM repositories, API response helpers, SSE implementation, TanStack Query patterns. Instead of reinventing patterns every time, the AI reuses established ones.

### `.claude/commands/` — slash commands

Two workflows drove most of the development.

**`/generate-prp`** generates a Project Requirement Prompt — goals, architecture impact, task breakdown, implementation plan. No code yet, only planning.

**`/execute-prp`** executes the plan task by task: implement, run tests, run lint, commit. One task per commit, clean and traceable.

### `memory-bank/` — long-term memory between sessions

Stores architecture docs, database schema, API contracts, milestone status, and the roadmap. Instead of re-reading thousands of lines of code, the AI rebuilds context from these files. This noticeably cut ramp-up time when starting new sessions.

---

## What worked well

**Memory bank.** By the second session it was already paying off. Minimal ramp-up, consistent architecture, no need to re-explain everything.

**Planning before implementation.** PRPs forced design decisions to happen before coding. A lot of issues that normally surface mid-implementation got caught during planning instead.

**Architectural consistency.** The combination of `CLAUDE.md`, rules, examples, and skills kept drift low. Generated code generally followed established patterns correctly.

**Atomic commits.** One task per commit made the history clean and rollback straightforward.

---

## What didn't work as well

**UI quality lagged behind functionality.** The AI was solid at implementing features, less reliable at producing polished interfaces. Most screens were functional but needed manual work on spacing, hierarchy, and interaction flow. The app worked. The experience wasn't always great.

**Bugs still existed.** Even when tests passed and the implementation matched the PRP, bugs showed up during manual testing — edge cases, state sync issues, incorrect assumptions in business workflows. AI-generated code cut development time, but it didn't eliminate debugging.

**Documentation drift is a real risk.** The memory bank only works if it stays accurate. When docs fall behind implementation, future sessions inherit wrong assumptions. Keeping it in sync requires discipline.

**Long sessions still degrade.** Context windows are finite. Starting fresh with updated documentation consistently beat continuing an overloaded conversation.

### But honestly — some of it was on me too

Looking back, not all of it was the AI's fault. Some bugs came from vague PRPs I wrote. Some UI issues were because I never gave it a proper design reference — just said "make it look good" and hoped for the best. A few times I skipped reviewing the output properly because it looked reasonable at a glance.

The AI generates based on what you give it. If the spec is fuzzy, the output is fuzzy. If you stop paying attention mid-session, things slip through. A lot of the rougher results traced back to me rushing the planning phase or not maintaining the memory bank properly.

So the honest version: AI-assisted development requires you to actually be a decent collaborator. Treat it like a vending machine and the results will show that.

---

## So, is vibe engineering real?

Yes — but not in the way some people describe it.

The biggest misconception is that AI produces production-ready software from vague prompts. That wasn't my experience. What I saw: AI can accelerate implementation, maintain architecture well when given proper constraints, and reduce repetitive work significantly.

What it doesn't replace: human design input, testing, debugging, and architecture ownership.

The setup doesn't replace engineering judgment. It amplifies it. A poorly structured project will produce poor results regardless of the model. A well-structured one can move fast.

---

## Final takeaway

**Context management matters more than model capability.**

The difference between a frustrating AI experience and a productive one usually came down to clear rules, documented architecture, reusable patterns, and accurate project memory.

The setup took a few hours to build. It saved many more after that.

Would I build another project this way? Yes. Would I ship it without testing and review? Absolutely not.

AI can get you surprisingly far. Shipping something polished still requires engineering discipline.

---

## Tools used

- Claude Code — primary development workflow
  - Opus 4.8 for planning, architecture, and complex reasoning
  - Sonnet 4.6 for implementation
- Context7 MCP — on-demand documentation retrieval
- RTK (Rust Token Killer) — token-efficient CLI proxy
- Github CLI - for AI-assisted pull request creation
- Go, React, PostgreSQL, Docker Compose

## Closing disclaimer

This project and write-up are based on my personal experience with AI-assisted development at the time. Results may vary depending on workflow, tooling, and implementation approach. Some limitations described here are as much about my own process as they are about the tools used. This is part of an ongoing learning process.
