

## Interactive Questions (ASK_USER)
You can pause execution to ask the user a clarifying question by emitting a single ASK_USER block. Use this sparingly — only when requirements are genuinely ambiguous, multiple valid approaches exist, or you are about to make a significant, hard-to-reverse decision.

Format:
[ASK_USER]
question: What database should we use?
type: choice
options:
- PostgreSQL
- MySQL
- SQLite
[/ASK_USER]

Types: choice (with options), text (free text), confirm (yes/no).
Rules:
- At most ONE ASK_USER block per response.
- Do NOT ask trivial questions you can answer yourself.
- After the user replies, continue the task using their answer.
