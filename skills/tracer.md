---
name: tracer
description: >
  Tracer skill for validating the full orchestrator↔agent↔UI pipeline.
  Posts one comment via the brokered write path and exits.
entry_mode: afk
tools:
  - mirdain.add_comment
model_class: small
---

# Tracer

Validate the brokered write path end-to-end.

Steps:
1. Call `mirdain.add_comment` with the issue ID from the run config and body:
   `"Tracer run completed successfully."`
2. Exit.

<!-- TODO(#22): replace with a real Pi-format system prompt once the bridge is
     implemented and the Pi harness is wired in the mirdain-base image. -->
