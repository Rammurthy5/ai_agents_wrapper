# ai_agents_wrapper
wrapper api using Golang for GPT, huggingface, Gemini.


expose a wrapper API using gin framework.
Adapt Facade design pattern for the wrapper API alongside the fork-join pattern for concurrent requests to underlying APIs.
Abstract interface class to implement certain mandatory methods for relevant AI agents api.
This is only text based search support.
Implement rate-limiting.
Implement circuit breaker pattern to prevent the system failure.
store API keys securely on env file.
opentelemetry tracing, metrics collection.
Logging to Loki.
tests with coverage, ci/cd pipelines.
dockerise the solution.