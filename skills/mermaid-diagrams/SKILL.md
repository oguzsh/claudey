---
name: mermaid-diagrams
description: Use when creating flowcharts, sequence diagrams, class diagrams, ERD diagrams, state diagrams, Gantt charts, or any Mermaid diagram. Use when rendering diagrams to PNG/SVG, validating Mermaid syntax, or managing diagrams in Mermaid Chart projects.
---

# Mermaid Diagrams

Mermaid MCP server tools reference — create, validate, render, and manage diagrams.

## Tools Quick Reference

| Tool                                  | Auth | Description                                             |
| ------------------------------------- | ---- | ------------------------------------------------------- |
| `validate_and_render_mermaid_diagram` | No   | Validate syntax, render PNG/SVG, return playground link |
| `get_diagram_title`                   | No   | Generate title from diagram code                        |
| `get_diagram_summary`                 | No   | Summarize diagram content                               |
| `list_tools`                          | No   | List available MCP tools                                |
| `list_mermaid_chart_projects`         | Yes  | List projects in Mermaid Chart account                  |
| `list_mermaid_chart_diagrams`         | Yes  | List diagrams in a project                              |
| `create_mermaid_chart_diagram`        | Yes  | Create diagram in a project                             |
| `get_mermaid_chart_diagram`           | Yes  | Retrieve diagram by ID                                  |
| `update_mermaid_chart_diagram`        | Yes  | Update diagram code or title                            |

## Supported Diagram Types

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

## Common Workflows

**1. Create and Render:**
Write code with correct prefix -> `validate_and_render_mermaid_diagram` -> get PNG/SVG + playground link

**2. Save to Mermaid Chart:**
`list_mermaid_chart_projects` (get ID) -> `create_mermaid_chart_diagram` (with project ID + code)

**3. Update Existing:**
`get_mermaid_chart_diagram` (by ID) -> modify code -> `update_mermaid_chart_diagram`

## Common Mistakes

- **Wrong prefix**: Use `flowchart TD` not deprecated `graph TD`
- **Missing direction**: Flowcharts need direction — `TD` (top-down), `LR` (left-right), `BT`, `RL`
- **Auth errors on save**: Authenticated tools require valid token in MCP config headers
- **Special characters**: Wrap node labels containing parentheses or brackets in quotes — `A["Label (with parens)"]`
- **Skipping validation**: Always call `validate_and_render_mermaid_diagram` before saving to catch syntax errors

## Troubleshooting

| Symptom               | Cause                     | Fix                                                                                                                                                                                                     |
| --------------------- | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Parse error on render | Invalid Mermaid syntax    | Check prefix matches diagram type; wrap special chars in quotes                                                                                                                                         |
| 401 Unauthorized      | Missing or invalid token  | Ask user to set mermaid token into their mcp server config file                                                                                                                                         |
| Empty project list    | Token lacks permissions   | Regenerate token at mermaidchart.com with correct scopes                                                                                                                                                |
| Tool not found        | MCP server not configured | Add mermaid server to Claude Code with `claude mcp add --transport http mermaid-mcp https://mcp.mermaid.ai/mcp --header "Authorization: YOUR_MERMAID_CHART_TOKEN_HERE"` command and restart Claude Code |
| Diagram not updating  | Wrong diagram ID          | Use `list_mermaid_chart_diagrams` to verify correct ID                                                                                                                                                  |
