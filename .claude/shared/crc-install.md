# CRC Modeling - Quick Setup

## What is This?

This directory contains the **CRC (Class-Responsibility-Collaboration) modeling system** - a reusable three-tier development methodology.

**Three-tier system:**
```
Level 1: Human specs (specs/*.md)        ← Requirements
   ↓
Level 2: Design models (design/*.md)     ← CRC cards, sequences, UI specs
   ↓
Level 3: Implementation (source code)    ← Actual code
```

---

## Quick Start

### Initialize CRC in Your Project

Run this command to set up CRC modeling:

```bash
# Using slash command (recommended)
/init-crc-project

# Or run the script directly
python3 ./.claude/scripts/init-crc-project.py
```

**What it does:**
- Creates `specs/` and `design/` directories
- Checks for required components (designer agent, PlantUML)
- Creates or updates `CLAUDE.md` with CRC workflow sections (if missing)
- Shows welcome message and next steps

**Safe to run multiple times** - idempotent, won't duplicate content.

**Note:** This command requires manual approval to protect your CLAUDE.md file.

---

## Complete Documentation

📚 **See `.claude/doc/crc.md` for complete CRC modeling documentation**

Topics covered:
- What is CRC modeling?
- Three-tier system explained
- Directory structure
- Required components
- Complete workflow
- CRC cards, sequence diagrams, UI specs
- Traceability and bidirectional updates
- Benefits and best practices

---

## Manual Setup (Alternative)

If you prefer to set up manually or need to copy to another project:

### 1. Copy Required Files

**Essential:**
- `.claude/agents/designer.md` - Core CRC spec generator
- `.claude/scripts/plantuml.py` - Sequence diagram generator
- `.claude/skills/plantuml.md` - PlantUML skill documentation
- `.claude/scripts/init-crc-project.py` - Initialization command
- `.claude/skills/init-crc-project.md` - Initialization skill
- `.claude/doc/crc.md` - Complete documentation

**Optional:**
- `.claude/agents/diagram-converter.md` - Alternative PlantUML converter
- `.claude/agents/gap-analyzer.md` - Gap analysis agent

### 2. Download PlantUML

```bash
mkdir -p .claude/bin
cd .claude/bin
curl -L -o plantuml.jar https://github.com/plantuml/plantuml/releases/download/v1.2024.3/plantuml-1.2024.3.jar
```

### 3. Create Directories

```bash
mkdir -p specs design
```

### 4. Update CLAUDE.md

Add CRC workflow sections to your project's `CLAUDE.md`:

```markdown
## CRC Modeling Workflow

**DO NOT generate code directly from `specs/*.md` files!**

**Use a three-tier system:**
```
Level 1: Human specs (specs/*.md)
   ↓
Level 2: Design models (design/*.md) ← CREATE THESE FIRST
   ↓
Level 3: Implementation (source code)
```

**Workflow:**
1. Read human specs (`specs/*.md`) for design intent
2. Use `designer` agent to create Level 2 specs (CRC cards, sequences, UI specs)
3. Generate code following complete specification with traceability comments

See `.claude/doc/crc.md` for complete documentation.
```

---

## CLAUDE.md Snippets

If you need specific sections to add to CLAUDE.md, here are the key pieces:

### Core Workflow Section

```markdown
## 🏗️ CRC Modeling Development Process

**Three-tier process**: Human-readable specs (`specs/*.md`) → CRC cards + Sequence diagrams + UI specs → Generated code/templates/tests

- **CRC Cards** (`design/crc-*.md`): Classes, responsibilities, collaborators → Source classes
  - One card per class in markdown format
  - Defines what each class knows (data) and does (behavior)
  - Identifies collaborations between classes

- **Sequence Diagrams** (`design/seq-*.md`): Object interactions over time → Method implementations
  - One diagram per scenario/use case
  - Shows how objects collaborate to fulfill requirements
  - Guides implementation details

- **UI Specs** (`design/ui-*.md`): Layout structure → Templates/views
  - Organized by view (may group small related components together)
  - Defines HTML structure, CSS classes, data bindings
  - References CRC cards for data types and behavior
  - Use `designer` agent to create UI specs

**Key principle**: CRC models are **source of truth** for structure, UI specs are **source of truth** for layout, human specs are **source of truth** for intent.

**Creating Level 2 Specs**: Use `designer` agent (`.claude/agents/designer.md`) for complete workflow
**Traceability**: [`design/traceability.md`](design/traceability.md) - Links from specs → CRC → code

See `.claude/doc/crc.md` for complete documentation.
```

### Daily Reminders Section

```markdown
⚠️ **CRITICAL - Creating Level 2 Specs:**
- **Use `designer` agent** (`.claude/agents/designer.md`) when creating CRC cards, sequences, or UI specs
  - Handles complete workflow: specs → CRC cards → sequence diagrams → UI specs
  - Manages traceability and gap analysis
  - Use: `Task(subagent_type="designer", ...)`
```

### Bidirectional Traceability Section

```markdown
### 🔄 Bidirectional Traceability Principle

**When changes occur at any level, propagate updates through the documentation hierarchy:**

**Source Code Changes → Design Specs:**
- Modified implementation → Update CRC cards/sequences/UI specs if structure/behavior changed
- New classes/methods → Create corresponding CRC cards
- Changed interactions → Update sequence diagrams
- Template/view changes → Update UI specs

**Design Spec Changes → Architectural Specs:**
- Modified CRC cards/sequences → Update high-level specs if requirements/architecture affected
- New components → Document in feature specs
- Changed workflows → Update architectural documentation
- UI pattern changes → Update UI principles

**Key Rules:**
1. **Always update up**: When code/design changes, ripple changes upward through documentation
2. **Maintain abstraction**: Each level documents at its appropriate abstraction
3. **Keep consistency**: All three tiers must tell the same story at their respective levels
4. **Update traceability comments**: When docs change, update CRC/spec references in code comments
```

---

## What This Provides

✅ **Better Architecture** - Explicit design phase prevents shotgun surgery and god classes
✅ **Complete Specifications** - Sequences catch edge cases before coding
✅ **Traceability** - Every line of code traces to design and requirements
✅ **Maintainability** - Changes propagate through all documentation levels
✅ **Onboarding** - New developers understand system from design docs
✅ **SOLID Principles** - CRC naturally encourages single responsibility and clear collaborations

---

## Learn More

📚 **Complete documentation:** `.claude/doc/crc.md`
🚀 **Initialize CRC:** `/init-crc-project`
🔧 **Designer agent:** `.claude/agents/designer.md`

---

**Last updated:** 2025-11-14
