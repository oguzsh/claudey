---
name: generating-mermaid-diagrams
description: Use when asked to create diagrams, flowcharts, architecture visuals, or render Mermaid syntax to PNG/SVG files using the local mmdc CLI tool. Not for inline Mermaid in markdown.
---

<goal>
	Render Mermaid diagram source files (.mmd) to PNG using the local `mmdc` CLI.
<goal>

<when-to-use>
- User asks for a diagram, flowchart, architecture visual, or sequence diagram as an image file
- Need to render `.mmd` files to PNG/SVG locally
- **NOT for**: inline Mermaid code blocks in markdown documents
</when-to-use>

<core-command>

```bash
mmdc -i diagram.mmd -o diagram.png -w 2400 -H 1800 -b white -q
```

| Flag | Purpose |
|------|---------|
| `-i` | Input .mmd file |
| `-o` | Output file (.png or .svg) |
| `-w 2400` | Width in pixels (prevents narrow images) |
| `-H 1800` | Height in pixels |
| `-b white` | Background color (default is transparent) |
| `-q` | Quiet mode (suppress stderr noise) |

</core-command>


<supported-diagram-types>

| Type      | Syntax Prefix     | Use Case                              |
| --------- | ----------------- | ------------------------------------- |
| Flowchart | `flowchart TD`    | Process flows, decision trees         |
| Sequence  | `sequenceDiagram` | API calls, message passing            |
| Class     | `classDiagram`    | OOP structures, relationships         |
| ERD       | `erDiagram`       | Database schemas, table relationships |
| State     | `stateDiagram-v2` | State machines, lifecycles            |
| Gantt     | `gantt`           | Project timelines, scheduling         |
| Pie       | `pie`             | Proportional data                     |
| Git Graph | `gitGraph`        | Branch strategies, merge flows        |
| Mindmap   | `mindmap`         | Brainstorming, topic hierarchies      |
| Timeline  | `timeline`        | Chronological events                  |
</supported-diagram-types>

<workflow>
1. **Write `.mmd` file** with correct diagram prefix and syntax
2. **Create puppeteer config** (only if sandbox errors occur)
3. **Run `mmdc`** with explicit width/height flags
4. **Verify PNG** by reading the output file to confirm it rendered correctly
</workflow>

<common-mistakes>
- *Using deprecated `graph TD`*: Use `flowchart TD` instead 
- *Missing direction on flowchart*: Always include `TD`, `LR`, `BT`, or `RL` 
- *Special chars in labels unquoted*: Wrap in quotes: `A["Label (with parens)"]` 
- *Using `-s` (scale) for dimensions*: Use `-w` and `-H` for explicit pixel control 
- *Forgetting `end` after `subgraph`*: Every `subgraph` block needs a closing `end` 
- *Arrow `->` instead of `-->`*: Flowchart arrows require `-->` (solid) or `-.->` (dotted) 
- *Colons in labels*: Escape or quote: `A["Step: Init"]` 
- *Not verifying output*: Always read the PNG after rendering to confirm correctness 
- *mmdc command not found*: If the user doesn't have mermaid-cli installed, they should use npx @mermaid-js/mermaid-cli <args> instead.
</common-mistakes>
