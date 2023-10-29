# Message Broker

As the final project in Bale Camp, We were asked to implement a fast and simple gRPC based broker. the main purpose of this project was to gain hands-on experience with golang and some common tools such as :

- Kubernetes:
  - Gained practical experience in utilizing Kubernetes for orchestrating workflows, ensuring smooth deployment, scaling, and management of containerized applications.
-Docker:
  -Acquired hands-on knowledge of Docker for streamlined containerization, facilitating consistent deployment across diverse environments.
- Envoy:
  - Worked extensively with Envoy as a proxy server for load balancing and rate limiting, elevating the reliability and performance of applications.
- Prometheus:
  - Utilized Prometheus for in-depth monitoring, enabling real-time tracking of system metrics and proactive identification of potential issues.
- Grafana:
  - Developed proficiency in Grafana to craft insightful dashboards, visually representing system performance for effective decision-making.
- Jaeger:
  - Gained valuable experience with Jaeger for distributed tracing, offering insights into request flow across the system and aiding in debugging and optimization.
- gRPC:
  - Implemented gRPC for seamless communication between components, leveraging its high-performance RPC capabilities to amplify system speed and responsiveness.
- k6:
  - Applied k6 for rigorous performance testi
- Concurrency in golang:
  - Gained practical experience with concurrency features in golang, such as goroutines and channels.
- Data Storages:
  - Used different datasources like Postgres (SQL), Cassandra/Scylla (noSQL) to persist the messages. We compared the performance of each and decided the best use case for each one.
- Caching and Batching:
  - In order to overcome the performance bottlenecks, we took advantage of different methods such as caching and batching.

<p align="center">
  <a href="https://skillicons.dev">
    <img src="https://skillicons.dev/icons?i=go,prometheus,grafana,postgres,cassandra,redis,kubernetes,docker" />
  </a>
</p>

## RPCs Interface
- Publish Requst
```protobuf
message PublishRequest {
  string subject = 1;
  bytes body = 2;
  int32 expirationSeconds = 3;
}
```
- Fetch Request
```protobuf
message FetchRequest {
  string subject = 1;
  int32 id = 2;
}
```
- Subscribe Request
```protobuf
message SubscribeRequest {
  string subject = 1;
}
```
- RPC Service
```protobuf
service Broker {
  rpc Publish (PublishRequest) returns (PublishResponse);
  rpc Subscribe(SubscribeRequest) returns (stream MessageResponse);
  rpc Fetch(FetchRequest) returns (MessageResponse);
}
```

## Performance Report 
In order to test our performance, We had several load testing phases on broker's publish service using K6 and custom-built 
golang clients.

#### RAM
![RAM](https://github.com/aref81/Message-Broker/assets/76614003/69b2e1f3-c84f-464a-b6ea-6474c5d9ff9a)
#### Postgres
![Postgres_2](https://github.com/aref81/Message-Broker/assets/76614003/a1ac2a22-036a-4355-84db-6222299603ce)
#### Cassandra
![Cassandra](https://github.com/aref81/Message-Broker/assets/76614003/5b18cf2a-277e-411a-afab-b1bdc1e473fa)
