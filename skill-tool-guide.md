# MaruBot Agent Architecture Principles

## Purpose

This document defines the **core architectural principles** for building and extending MaruBot, ensuring high efficiency and modularity.

The goal is to ensure the system remains:

* lightweight
* modular
* secure
* extensible
* observable
* maintainable

This document is intended to guide **MaruBot's Auto-Evolution (Self-Improvement)** capability and developers when implementing new features or refactoring the system.

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

The core must **never directly implement domain functionality** such as specialized hardware control or complex web tasks. Those must always be implemented as tools or skills.

---

## 2.2 Tools

Tools represent **atomic executable capabilities**.

A tool must:

* perform one well-defined task
* accept structured input
* return structured output
* finish quickly
* be stateless whenever possible

Tools represent the **hands of MaruBot**.

---

## 2.3 Skills

Skills define **behavioral strategies**.

A skill describes:

* when it should be used
* which tools it may call
* how those tools should be combined
* the policies governing execution

Skills represent the **knowledge of how to use the hands**.

---

# 3. Auto-Evolution Workflow

MaruBot has the capability to expand itself. Use the following tools to evolve:

1. **`create_tool`**: Use this to create a new atomic capability (Bash/Python script).
2. **`create_skill`**: Use this to define a new high-level behavior (SKILL.md).

When a user requests a new feature:
1. Identify if it requires an atomic action (Tool) or complex reasoning/guidelines (Skill).
2. Use the appropriate creation tool.
3. The new capability is available IMMEDIATELY.

---

# 4. Golden Rule

The architecture must always respect this principle:

> **Tools execute actions. Skills decide how actions are combined. The core orchestrates everything.**

---

MaruBot Architecture Principles
