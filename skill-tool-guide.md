# MaruBot Agent Architecture Principles

## Purpose

This document defines the **core architectural principles** for building a lightweight, extensible AI agent system similar to OpenClaw while maintaining PicoClaw-level efficiency.

The goal is to ensure the system remains:

* lightweight
* modular
* secure
* extensible
* observable
* maintainable

This document is intended to guide **AI coding agents (Claude, AntiGravity, etc.)** and developers when implementing new features or refactoring the system.

---

# 1. Core Philosophy

The system must follow the rule:

> **Minimal core, maximal extension.**

The core agent should remain extremely small and stable, while all heavy functionality is implemented through modular extensions.

Core responsibilities should be limited to:

* session management
* tool dispatch
* skill orchestration
* permission enforcement
* worker supervision
* memory management
* logging and observability

Everything else must be implemented outside the core.

---

# 2. Strict Separation of Concepts

The architecture distinguishes **three different layers**:

1. Core
2. Tools
3. Skills

These must never be mixed.

---

## 2.1 Core

The **core** is the runtime of the agent.

Responsibilities:

* LLM communication
* tool invocation
* skill orchestration
* session tracking
* permission validation
* worker lifecycle management
* event routing

The core must **never directly implement domain functionality** such as:

* camera control
* GPIO control
* web browsing
* audio recording
* RAG indexing

Those must always be implemented as tools.

---

## 2.2 Tools

Tools represent **atomic executable capabilities**.

A tool must:

* perform one well-defined task
* accept structured input
* return structured output
* finish quickly
* be stateless whenever possible

Examples:

* read_file
* http_get
* gpio_write
* camera_capture
* rag_search
* rag_index

Tools must never contain complex orchestration logic.

Tools are the **hands of the agent**.

---

## 2.3 Skills

Skills define **behavioral strategies**.

A skill describes:

* when it should be used
* which tools it may call
* how those tools should be combined
* the policies governing execution

Skills are typically composed of:

* documentation (`SKILL.md`)
* manifest (`manifest.json`)
* optional workflow definition
* optional execution runner

Skills are the **knowledge of how to use the hands**.

---

# 3. Execution Model

Tools may run in one of three execution modes.

---

## 3.1 In-Process Tools

Used only for extremely lightweight operations.

Examples:

* configuration lookup
* session state queries
* cache retrieval
* small database queries
* simple string transformations

Rules:

* must not block
* must not allocate large memory
* must not interact with hardware
* must not spawn subprocesses

---

## 3.2 Subprocess Tools

Most tools should run as subprocesses.

Examples:

* RAG search/index
* shell commands
* camera capture
* audio recording
* video processing
* OCR
* browser automation
* external ML models

Advantages:

* crash isolation
* memory containment
* security separation
* dynamic loading/unloading

Subprocess communication must use structured IPC.

---

## 3.3 Long-Running Workers

Some capabilities require persistent services.

Examples:

* ragd
* visiond
* voiced
* browserd

These workers should be managed by a **worker supervisor** in the core.

Workers must support:

* health checks
* restart
* version reporting
* capability reporting

---

# 4. Inter-Process Communication

Subprocess tools and workers must communicate with the core through **structured IPC**.

Recommended approach:

JSON messages over stdin/stdout.

Message types:

* tool_call
* tool_result
* ping
* pong
* log

Example request:

```
{
  "type": "tool_call",
  "id": "req_001",
  "tool": "camera.capture",
  "input": {
    "device": "front_cam"
  }
}
```

Example response:

```
{
  "type": "tool_result",
  "id": "req_001",
  "ok": true,
  "content": {
    "image_path": "/tmp/capture_001.jpg"
  }
}
```

---

# 5. Permission Model

All tools and skills must declare required permissions.

Example capabilities:

* fs.read
* fs.write
* net.http
* net.notify
* gpio.read
* gpio.write
* camera.capture
* audio.record
* browser.open
* rag.search
* rag.index
* shell.exec

The core must verify permissions before executing any tool.

Skills must explicitly declare:

* allowed_tools
* permissions

Execution must fail if permissions are not granted.

---

# 6. Skill Structure

Each skill must follow this directory structure:

```
skills/
  skill-name/
    SKILL.md
    manifest.json
    workflow.json (optional)
    runner (optional)
```

### SKILL.md

Human and LLM readable documentation describing:

* when to use the skill
* how it behaves
* limitations
* policies

### manifest.json

Machine readable metadata.

Example fields:

* id
* version
* description
* allowed_tools
* permissions
* triggers
* timeout

### workflow.json (optional)

Declarative execution graph describing tool usage.

### runner (optional)

Executable script or binary implementing complex logic.

---

# 7. Memory Architecture

The agent must support three memory layers.

### Short-Term Memory

Session context and recent interaction history.

### Long-Term Memory

User preferences and persistent facts.

### Retrieval Memory

Document or knowledge retrieval through search or vector index.

RAG must be implemented as **tools or workers**, never directly inside the core.

---

# 8. Observability

The system must provide complete execution transparency.

Required logging:

* tool invocation
* tool result
* skill execution
* permission checks
* worker lifecycle
* errors and retries

Logs must include:

* session_id
* trace_id
* tool_id
* timestamp

Observability is required for debugging agent behavior.

---

# 9. Dynamic Loading and Unloading

The architecture must support dynamic capability management.

Required features:

* install tool
* uninstall tool
* enable tool
* disable tool
* start worker
* stop worker
* reload skill

Tools and skills must be discoverable at runtime through manifests.

---

# 10. Design Constraints

To maintain a lightweight agent architecture:

The core must:

* remain below minimal memory footprint
* avoid heavy libraries
* avoid embedding large models
* avoid long blocking tasks
* avoid hardware access

Heavy computation must always run outside the core.

---

# 11. Golden Rule

The architecture must always respect this principle:

> **Tools execute actions. Skills decide how actions are combined. The core orchestrates everything.**

Violating this rule will lead to:

* bloated core
* reduced modularity
* increased maintenance complexity
* reduced system stability

---

# 12. Implementation Guidance for AI Coding Agents

When adding new functionality:

1. Determine if the feature is a **tool or a skill**.
2. Never place domain functionality inside the core.
3. Prefer subprocess tools over in-process execution.
4. Ensure every tool declares permissions.
5. Provide structured input/output schemas.
6. Ensure all actions are observable via logs.

If unsure, always choose **modularity over convenience**.

---

# End of Document

MaruBot Architecture Principles
