# AI summaries

Set `OPENROUTER_API_KEY` to enable DeepSeek Flash. `OPENROUTER_MODEL` optionally
pins a model; by default the bot uses `~deepseek/deepseek-flash-latest`.
This is a paid model, not an OpenRouter free-tier model. Set an OpenRouter key
credit limit as an additional account-wide spending safeguard.

The application router reserves at most **20 attempts per chat per UTC day** in
PostgreSQL. Failed attempts count, and usage survives restarts. Exhaustion,
missing configuration, quota-storage errors, and primary errors use the existing
default client (Google, then OpenAI, then direct DeepSeek, according to configured
credentials). There is one fallback attempt and no automatic primary retries.
Apply the migration with `make db-migrate` before enabling OpenRouter.

`/ai quota` shows remaining/total requests, usage, and the next UTC reset. It does
not consume quota. The quota is per chat, not per user, and limits dispatched
OpenRouter calls rather than any internal provider-routing attempts.

## Resource budgets

| Environment variable | Default | Accepted range |
| --- | --- | --- |
| `AI_MAX_INPUT_CHARS` | 32000 | 1000–60000 |
| `AI_MAX_OUTPUT_TOKENS` | 2048 | 256–4096 |
| `AI_MAX_CONCURRENT_SUMMARIES` | 3 | 1–8 |

Invalid settings use defaults. Input is also capped at 10,000 words and 2,000
messages. History is read newest-first in bounded pages, preserving authors and
message boundaries, then presented chronologically. Oversized history uses its
most recent portion and the response includes a warning. Character counts are
a conservative size budget, not exact tokenizer counts. Output-token limits
include reasoning tokens where the provider counts them in its output budget;
incomplete/blocked responses are not presented as complete summaries.

Only one summary per chat is admitted at a time. Admission is non-blocking and
does not consume quota when busy. Concurrency limits are process-local; the
database request quota remains atomic across instances. Primary and fallback
have separate timeouts. AI logs contain metadata, not conversation text or keys.
Message persistence is time-limited. Summaries still require readable message
history; a database outage returns a controlled error rather than calling AI
without the requested conversation.

For a local run, load your environment from `local/docker_env.sh`, run
`make db-bootstrap` and `make db-migrate`, then `make run`. When running the bot
outside Docker, set `DB_HOST` to the published local PostgreSQL host and port.
The Makefile and deployment workflow forward the OpenRouter and budget settings.
