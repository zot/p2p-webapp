---
description: Initialize CRC modeling in the current project
---

Run the CRC project initialization script:

```bash
python3 ./.claude/scripts/init-crc-project.py
```

This will:
- Create `specs/` and `design/` directories
- Check for required components (designer agent, PlantUML)
- Create or update `CLAUDE.md` with CRC workflow sections
- Show welcome message and next steps

Safe to run multiple times - idempotent, won't duplicate content.
