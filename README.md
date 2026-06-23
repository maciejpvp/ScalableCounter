# ScalableCounter

ScalableCounter is a Go-based microservice for tracking and incrementing video likes at scale. It sits behind AWS CloudFront, an Application Load Balancer, an EC2 instance running Docker, and DynamoDB.

## Purpose and Architecture

The service is built for write-heavy workloads — specifically, high volumes of real-time video "like" events — while keeping database transaction costs low and throughput high.

### Key Components

1. **Go API Service (`/go`)**:
   - Uses the `go-chi` router.
   - Exposes three REST endpoints:
     - `POST /video` - Create a new video record.
     - `GET /video/{id}` - Retrieve a video and its current like count.
     - `POST /video/{id}/like` - Increment the like count for a video.
   - **Buffered-Write Pattern**: Rather than hitting DynamoDB on every incoming like (expensive and slow under high concurrency), the service accumulates likes in memory and flushes them to the database in a single batch every 30 seconds via a background worker. Read responses merge the in-memory buffer with the stored value to maintain eventual consistency.

2. **Terraform Infrastructure (`/infra`)**:
   - **CloudFront CDN**: Acts as the public entry point with global low-latency distribution. A CloudFront Function strips the `/api` prefix from incoming requests (rewriting `/api/video/*` to `/video/*`) before they reach the load balancer. Caching is disabled (`Managed-CachingDisabled`) so all requests pass through to the backend.
   - **Application Load Balancer (ALB)**: Routes traffic to EC2. The ALB listener only forwards requests that include a secret HTTP header (`X-From-CloudFront`) generated at deploy time, rejecting direct origin access.
   - **EC2 Instance**: A single `t3.micro` instance running Docker, configured to automatically pull and start the application container from ECR.
   - **DynamoDB**: An on-demand table (`Videos`) that scales automatically without provisioned capacity.

---

## AWS Cost Estimation (eu-central-1 Region)

The architecture minimizes idle costs by using on-demand/serverless pricing where possible, with a single low-cost compute instance.

### 1. Static Monthly Cost (0 requests)

These costs are incurred regardless of traffic volume:

| Service | Configuration | Monthly Cost (Est.) | Details |
| :--- | :--- | :--- | :--- |
| **EC2 Compute** | 1x `t3.micro` instance | ~$8.76 | $0.012/hour * 730 hours |
| **EBS Storage** | 8 GB gp3 Volume | ~$0.64 | $0.08/GB/month |
| **Application Load Balancer** | 1x ALB (Base Rate) | ~$16.43 | $0.0225/hour * 730 hours |
| **CloudFront CDN** | CDN Distribution | $0.00 | No baseline cost |
| **DynamoDB** | On-Demand Table | $0.00 | No baseline cost for empty table |
| **ECR Registry** | Docker image (~100MB) | ~$0.01 | $0.10/GB/month |
| **VPC & Network** | Public subnet (No NAT GW) | $0.00 | Uses default internet gateway |
| **Total Static Cost** | | **~$25.84 / month** | |

---

### 2. Dynamic Cost (Per 1 Million Requests)

Estimates assume an average payload of **1 KB** (standard JSON) and a 50/50 split between reads (`GET /video/{id}`) and likes (`POST /video/{id}/like`).

#### Without Memory Buffering

Every like hits DynamoDB directly. 1M requests = 500k reads + 500k writes:

- **CloudFront**: ~$1.10 (requests + ~1 GB egress + Function executions)
- **ALB LCU**: ~$0.00 (1M requests/month is ~0.38 req/sec, within the base LCU allowance)
- **DynamoDB Reads (RRU)**: $0.125 ($0.25 per million RRUs)
- **DynamoDB Writes (WRU)**: $0.625 ($1.25 per million WRUs)
- **Total Dynamic Cost: ~$1.86 per 1M requests**

#### With 30-Second Memory Buffering

The Go server holds likes in memory and flushes the accumulated count to DynamoDB once every 30 seconds. Assuming 500k likes arrive at a sustained rate of 10 requests/second over 50,000 active seconds, the number of actual DynamoDB write calls drops from 500,000 to roughly 1,667 (one flush per 30-second window).

- **CloudFront**: ~$1.10
- **ALB LCU**: ~$0.00
- **DynamoDB Reads (RRU)**: $0.125 (reads are not buffered)
- **DynamoDB Writes (WRU)**: ~$0.002 (only ~1,667 write calls reach DynamoDB)
- **Total Dynamic Cost: ~$1.23 per 1M requests**

#### Cost Savings Summary

| Metric | Without Buffering | With 30s Buffer | Reduction |
| :--- | :--- | :--- | :--- |
| DynamoDB write calls (per 1M reqs) | 500,000 | ~1,667 | ~99.7% |
| DynamoDB write cost (per 1M reqs) | $0.625 | ~$0.002 | ~99.7% |
| Total dynamic cost (per 1M reqs) | ~$1.86 | ~$1.23 | ~33.9% |

At 10 million requests per month, the 30-second buffer saves roughly **$6.30/month on dynamic costs alone** compared to direct per-request writes — a reduction of about **34%** on the variable portion of the bill. The higher the write volume, the more pronounced the savings.

---

## Deploying the Infrastructure

1. Initialize Terraform:
```bash
   cd infra
   terraform init
```

2. Inspect the planned changes:
```bash
   terraform plan
```

3. Deploy to AWS:
```bash
   terraform apply
```