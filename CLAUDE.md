# p2p-webapp
- A Go application to host peer-to-peer applications
- Proxies opinionated IPFS and libp2p operations for managed peers
- Provides a TypeScript library for easy communication with the Go application

## 🎯 Core Principles
- Use **SOLID principles** in all implementations
- Create comprehensive **unit tests** for all components
- code and specs are as MINIMAL as POSSIBLE

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
2. Use `designer` agent to create Level 2 specs (CRC cards, sequences, UI specs, architecture mapping)
3. Generate code following complete specification with traceability comments

**Design Entry Point:**
- `design/architecture.md` serves as the "main program" for the design
- Shows how design elements are organized into logical systems
- Start here to understand the overall architecture
- **Use for problem diagnosis and impact analysis** - quickly localize issues and assess change scope

**When to Read architecture.md:**
- **When working with design files, implementing features, or diagnosing issues, always read `design/architecture.md` first to understand the system structure and component relationships.**

**Traceability Comment Format:**
- Use simple filenames WITHOUT directory paths
- ✅ Correct: `CRC: crc-Person.md`, `Spec: main.md`, `Sequence: seq-create-user.md`
- ❌ Wrong: `CRC: design/crc-Person.md`, `Spec: specs/main.md`

**Test Implementation:**
- Test files belong in top-level `tests/` directory (NOT nested under `src/`)
- Test designs reference: `Test Design: test-ComponentName.md`
- When configuring build tools (Vite, Webpack, etc.), ensure test runner configurations are separate from application build configurations
- If build config sets a custom `root` directory, create a separate test configuration file to avoid test discovery issues
- Run `npm test` to verify test discovery works correctly before considering tests complete

See `.claude/doc/crc.md` for complete documentation.

### 🔄 Bidirectional Traceability Principle

**When changes occur at any level, propagate updates through the documentation hierarchy:**

**Source Code Changes → Design Specs:**
- Modified implementation → Update CRC cards/sequences/UI specs if structure/behavior changed
- New classes/methods → Create corresponding CRC cards
- Changed interactions → Update sequence diagrams
- Template/view changes → Update UI specs

**Design Spec Changes → Architectural Specs:**
- Modified CRC cards/sequences → Update high-level specs if requirements/architecture affected
- New components → Document in feature specs and update `design/architecture.md`
- Changed workflows → Update architectural documentation
- System reorganization → Update `design/architecture.md` to reflect new system boundaries

**Key Rules:**
1. **Always update up**: When code/design changes, ripple changes upward through documentation
2. **Maintain abstraction**: Each level documents at its appropriate abstraction
3. **Keep consistency**: All three tiers must tell the same story at their respective levels
4. **Update traceability comments**: When docs change, update CRC/spec references in code comments

### 📚 Documentation Generation

**After completing design or implementation work, offer to generate or update project documentation.**

Use the `documenter` agent to create:
- `docs/requirements.md` - Requirements documentation from specs
- `docs/design.md` - Design overview from CRC cards and sequences
- `docs/developer-guide.md` - Developer documentation with architecture and setup
- `docs/user-manual.md` - User manual with features and how-to guides
- `design/traceability-docs.md` - Documentation traceability map

**When to offer documentation generation:**
- After creating/updating Level 2 design specs
- After implementing Level 3 code
- When specs or design changes significantly
- When user explicitly requests it

**Example offer:**
"I've completed the [design/implementation]. Would you like me to generate/update the project documentation (requirements, design overview, developer guide, and user manual)?"

## 🔨 Building
```bash
# Run the demo (automatically builds and extracts bundled demo)
make demo

# Build everything (TypeScript library + Go binary with bundled demo)
make build      # or just: make
                # automatically installs dependencies if needed

# Clean all build artifacts
make clean
```

## 🧪 Testing with Playwright
When testing with Playwright MCP:
1. **ALWAYS check for running instances BEFORE starting tests**
   ```bash
   pgrep -a p2p-webapp  # Check for any running instances
   kill -9 <PID>         # Kill if found
   ```
2. **Always use an empty tmp directory** for the extract command
   ```bash
   cd /tmp/ipfs-extract-test && rm -rf * && /path/to/p2p-webapp extract --noopen -v
   ```
3. **Track and kill processes properly**
   - **IMPORTANT**: DO NOT use `ps aux | grep p2p-webapp` to find the PID!
     - This grep pattern will match BOTH the p2p-webapp binary AND the Claude process
     - The Claude process command line contains the working directory path which includes "p2p-webapp"
     - Using this pattern with kill will accidentally kill Claude too!
   - **Safe alternatives**:
     - Use `pgrep -f "p2p-webapp extract"` to find only the actual binary
     - Use `pgrep p2p-webapp` to find by process name only
     - Capture the PID when starting: `./p2p-webapp extract --noopen -v & echo $!`
   - Kill and verify: `kill <PID> && sleep 1 && ps -p <PID>`
   - The `extract` command requires an empty directory or it will fail

The build process:
1. Checks and installs TypeScript dependencies if `node_modules` is missing
2. Compiles the TypeScript client library (`pkg/client/src/`) to ES modules
3. Copies the compiled library to `internal/commands/demo/` for bundling
4. Builds a temporary Go binary
5. Bundles the demo site (from `internal/commands/demo/`) into the final binary using ZIP append
6. The final binary always ships with the demo bundled and ready to extract
7. Users can extract the demo with the `extract` command and get the client library files
8. The `ls` and `cp` commands operate directly on the bundled content (no go:embed needed)
