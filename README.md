# ai_agents_wrapper
wrapper api using Golang for GPT, huggingface, Gemini.


expose a wrapper API using gin framework.
Adapt Facade design pattern for the wrapper API alongside the fork-join pattern for concurrent requests to underlying APIs.
Abstract interface class to implement certain mandatory methods for relevant AI agents api.
we add retry for 3rd party api call to handle transient failures.
This is only text based search support.
Add a queue between API server and worker to make the service work in a distributed way.
Store the results mapping a taskID in Redis.
Implement rate-limiting.
Implement circuit breaker pattern using 'gobreaker' to prevent the system failure.
store API keys securely on env file.
opentelemetry tracing, metrics collection.
Logging to Loki.
tests with coverage, ci/cd pipelines.
dockerise the solution.

Reasoning to choose RabbitMQ for queue:
robust, and prod-ready. 
redis is single threaded and isnt as feature-rich.