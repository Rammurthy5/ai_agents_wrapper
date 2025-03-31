# ai_agents_wrapper
distributed wrapper api using Golang for GPT, huggingface, Gemini.


expose a wrapper API using gin framework.
Adapt Facade design pattern for the wrapper API alongside the fork-join pattern for concurrent requests to underlying APIs.
Abstract interface class to implement certain mandatory methods for relevant AI agents api.
we add retry for 3rd party api call to handle transient failures.
This is only text based search support.
Add a queue between API server and worker to make the service work in a distributed way.
Store the results mapping a taskID in Redis.
Implement rate-limiting. [no plans to do]
Implement circuit breaker pattern using 'gobreaker' to prevent the system failure.
store API keys securely on env file.
opentelemetry tracing, metrics collection.
Logging to Loki.
tests with coverage, ci/cd pipelines.
dockerise the solution.

Reasoning to choose RabbitMQ for queue:
robust, and prod-ready. 
redis is single threaded and isnt as feature-rich.


docker run -d --name rabbitmq -p 5672:5672 rabbitmq:3-management
docker run -d --name redis -p 6379:6379 redis:latest

go build -o api_server ./cmd/api_server
go build -o worker ./cmd/worker
go build -o cli ./cmd/cli

docker run -d --name api_server -p 8080:8080 -v $(pwd)/.env:/app/.env api_server:latest

docker build -t worker:latest .
docker run -d --name worker -v $(pwd)/.env:/app/.env worker:latest