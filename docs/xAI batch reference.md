#### Advanced API Usage

# Batch API

## Implementation Memo (support-cloud-llm-across-app-and-batch / Task 3.1)

### 現行実装との差分（対応方針）
- status 正規化:
  - 旧: `validating` / `in_progress` など xAI ローカル状態を返却
  - 新: 共通 `queued / running / completed / partial_failed / failed / cancelled` へ正規化
- status 判定入力:
  - `state.num_requests`
  - `state.num_pending`
  - `state.num_success`
  - `state.num_error`
  - `state.num_cancelled`
- results ページング:
  - `GET /v1/batches/{batch_id}/results` の `pagination_token` を再帰的に追跡して全件取得
  - `page_size` は明示指定可能（実装側は token 駆動で終端判定）

### xAI 状態→共通状態マッピング
- `num_pending > 0 -> running`
- `num_pending = 0 && num_success = num_requests -> completed`
- `num_pending = 0 && num_cancelled = num_requests -> cancelled`
- `num_pending = 0 && num_error = num_requests -> failed`
- `num_pending = 0 && num_success > 0 && (num_error > 0 || num_cancelled > 0) -> partial_failed`
- `num_requests = 0` は開始前として `queued`

The Batch API lets you process large volumes of requests asynchronously with reduced pricing and higher rate limits. For pricing details, see [Batch API Pricing](/developers/models#batch-api-pricing).

## What is the Batch API?

When you make a standard API call to Grok, you send a request and wait for an immediate response. This approach is perfect for interactive applications like chatbots, real-time assistants, or any use case where users are waiting for a response.

The Batch API takes a different approach. Instead of processing requests immediately, you submit them to a queue where they're processed in the background. You don't get an instant response—instead, you check back later to retrieve your results.

**Key differences from real-time API requests:**

| | Real-time API | Batch API |
|---|---|---|
| **Response time** | Immediate (seconds) | Typically within 24 hours |
| **Cost** | Standard pricing | Reduced pricing ([see details](/developers/models#batch-api-pricing)) |
| **Rate limits** | Per-minute limits apply | Requests don't count towards rate limits |
| **Use case** | Interactive, real-time | Background processing, bulk jobs |

**Processing time:** Most batch requests complete within **24 hours**, though processing time may vary depending on system load and batch size.

You can also create, monitor, and manage batches through the [xAI Console](https://console.x.ai/team/default/batches). The Console provides a visual interface for tracking batch progress and viewing results.

## When to use the Batch API

The Batch API is ideal when you don't need immediate results and want to **reduce your API costs**:

* **Running evaluations and benchmarks** - Test model performance across thousands of prompts
* **Processing large datasets** - Analyze customer feedback, classify support tickets, extract entities
* **Content moderation at scale** - Review backlogs of user-generated content
* **Document summarization** - Process reports, research papers, or legal documents in bulk
* **Data enrichment pipelines** - Add AI-generated insights to database records
* **Scheduled overnight jobs** - Generate daily reports or prepare data for dashboards

## How it works

The Batch API workflow consists of four main steps:

1. **Create a batch** - A batch is a container that groups related requests together
2. **Add requests** - Submit your inference requests to the batch queue
3. **Monitor progress** - Poll the batch status to track completion
4. **Retrieve results** - Fetch responses for all processed requests

Let's walk through each step.

## Step 1: Create a batch

A batch acts as a container for your requests. Think of it as a folder that groups related work together—you might create separate batches for different datasets, experiments, or job types.

When you create a batch, you receive a `batch_id` that you'll use to add requests and retrieve results.

```bash
curl -X POST https://api.x.ai/v1/batches \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer $XAI_API_KEY" \\
  -d '{
    "name": "customer_feedback_analysis"
  }'
```

```pythonXAI
from xai_sdk import Client

client = Client()

# Create a batch with a descriptive name
batch = client.batch.create(batch_name="customer_feedback_analysis")
print(f"Created batch: {batch.batch_id}")

# Store the batch_id for later use
batch_id = batch.batch_id
```

## Step 2: Add requests to the batch

With your batch created, you can now add requests to it. Each request is a standard chat completion that will be processed asynchronously.

**With the xAI SDK, adding batch requests is simple:** create `Chat` objects the same way you would for regular chat completions, then pass them as a list. You don't need to construct JSONL files or deal with complex request formats. Just use the familiar `chat.create()` and `chat.append()` pattern you already know.

**Important:** Assign a unique `batch_request_id` to each request. This ID lets you match results back to their original requests, which becomes important when you're processing hundreds or thousands of items. If you don't provide an ID, we generate a UUID for you. Using your own IDs is useful for idempotency (ensuring a request is only processed once) and for linking batch requests to records in your own system.

```pythonXAI
from xai_sdk import Client
from xai_sdk.chat import system, user

client = Client()

# Sample data to process
feedback_items = [
    {"id": "feedback_001", "text": "The product exceeded my expectations!"},
    {"id": "feedback_002", "text": "Shipping took way too long."},
    {"id": "feedback_003", "text": "It works as described, nothing special."},
]

# Build batch requests using familiar Chat objects
batch_requests = []
for item in feedback_items:
    # Create a Chat exactly like you would for a regular request
    chat = client.chat.create(
        model="grok-4.20-beta-latest-non-reasoning",
        batch_request_id=item["id"],  # Add an ID to track this request
    )
    # Append messages the same way as always
    chat.append(system("Classify the sentiment as positive, negative, or neutral."))
    chat.append(user(item["text"]))
    
    batch_requests.append(chat)

# Pass the list of Chat objects to the batch
client.batch.add(batch_id=batch.batch_id, batch_requests=batch_requests)
print(f"Added {len(batch_requests)} requests to batch")
```

```bash
curl -X POST https://api.x.ai/v1/batches/{batch_id}/requests \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer $XAI_API_KEY" \\
  -d '{
    "batch_requests": [
      {
        "batch_request_id": "feedback_001",
        "batch_request": {
          "chat_get_completion": {
            "messages": [
              {"role": "system", "content": "Classify the sentiment as positive, negative, or neutral."},
              {"role": "user", "content": "The product exceeded my expectations!"}
            ],
            "model": "grok-4.20-beta-latest-non-reasoning"
          }
        }
      },
      {
        "batch_request_id": "feedback_002",
        "batch_request": {
          "chat_get_completion": {
            "messages": [
              {"role": "system", "content": "Classify the sentiment as positive, negative, or neutral."},
              {"role": "user", "content": "Shipping took way too long."}
            ],
            "model": "grok-4.20-beta-latest-non-reasoning"
          }
        }
      }
    ]
  }'
```

## Step 3: Monitor batch progress

After adding requests, they begin processing in the background. Since batch processing is asynchronous, you need to poll the batch status to know when results are ready.

The batch state includes counters for pending, successful, and failed requests. Poll periodically until `num_pending` reaches zero, which indicates all requests have been processed (either successfully or with errors).

```bash
# Check batch status
curl https://api.x.ai/v1/batches/{batch_id} \\
  -H "Authorization: Bearer $XAI_API_KEY"

# Response includes state with request counts:
# {
#   "state": {
#     "num_requests": 100,
#     "num_pending": 25,
#     "num_success": 70,
#     "num_error": 5
#   }
# }
```

```pythonXAI
import time
from xai_sdk import Client

client = Client()

# Poll until all requests are processed
print("Waiting for batch to complete...")
while True:
    batch = client.batch.get(batch_id=batch.batch_id)
    
    pending = batch.state.num_pending
    completed = batch.state.num_success + batch.state.num_error
    total = batch.state.num_requests
    
    print(f"Progress: {completed}/{total} complete, {pending} pending")
    
    if pending == 0:
        print("Batch processing complete!")
        break
    
    # Wait before polling again (avoid hammering the API)
    time.sleep(5)
```

### Understanding batch states

The Batch API tracks state at two levels: the **batch level** and the **individual request level**.

**Batch-level state** shows aggregate progress across all requests in a given batch,
accessible through the `batch.state` object returned by the `client.batch.get()` method:

| Counter | Description |
|---|---|
| `num_requests` | Total number of requests added to the batch |
| `num_pending` | Requests waiting to be processed |
| `num_success` | Requests that completed successfully |
| `num_error` | Requests that failed with an error |
| `num_cancelled` | Requests that were cancelled |

When `num_pending` reaches zero, all requests have been processed (either successfully, with errors, or cancelled).

**Individual request states** describe where each request is in its lifecycle, accessible through the `batch_request_metadata` object returned by the `client.batch.list_batch_requests()` [method](#check-individual-request-status):

| State | Description |
|---|---|
| `pending` | Request is queued and waiting to be processed |
| `succeeded` | Request completed successfully, result is available |
| `failed` | Request encountered an error during processing |
| `cancelled` | Request was cancelled (e.g., when the batch was cancelled before this request was processed) |

**Batch lifecycle:** A batch can also be cancelled or expire. [If you cancel a batch](#cancel-a-batch), pending requests won't be processed, but already-completed results remain available. Batches have an expiration time after which results are no longer accessible—check the `expires_at` field when retrieving batch details.

## Step 4: Retrieve results

You can retrieve results at any time, even before the entire batch completes. Results are available as soon as individual requests finish processing, so you can start consuming completed results while other requests are still in progress.

Each result is linked to its original request via the `batch_request_id` you assigned earlier. The `result.response` object is the same SDK `Response` you'd get from a regular chat completion, with all the familiar fields: `.content`, `.usage`, `.finish_reason`, and more.

The SDK provides convenient `.succeeded` and `.failed` properties to separate successful responses from errors.

**Pagination:** Results are returned in pages. Use the `limit` parameter to control page size and `pagination_token` to fetch subsequent pages. When `pagination_token` is `None`, you've reached the end.

```pythonXAI
from xai_sdk import Client

client = Client()

# Paginate through all results
all_succeeded = []
all_failed = []
pagination_token = None

while True:
    # Fetch a page of results (limit controls page size)
    page = client.batch.list_batch_results(
        batch_id=batch.batch_id,
        limit=100,
        pagination_token=pagination_token,
    )
    
    # Collect results from this page
    all_succeeded.extend(page.succeeded)
    all_failed.extend(page.failed)
    
    # Check if there are more pages
    if page.pagination_token is None:
        break
    pagination_token = page.pagination_token

# Process all results
print(f"Successfully processed: {len(all_succeeded)} requests")
for result in all_succeeded:
    # Access the full Response object
    print(f"[{result.batch_request_id}] {result.response.content}")
    print(f"  Tokens used: {result.response.usage.total_tokens}")

if all_failed:
    print(f"\\nFailed: {len(all_failed)} requests")
    for result in all_failed:
        print(f"[{result.batch_request_id}] Error: {result.error_message}")
```

```bash
# Fetch first page
curl "https://api.x.ai/v1/batches/{batch_id}/results?page_size=100" \\
  -H "Authorization: Bearer $XAI_API_KEY"

# Use pagination_token from response to fetch next page
curl "https://api.x.ai/v1/batches/{batch_id}/results?page_size=100&pagination_token={token}" \\
  -H "Authorization: Bearer $XAI_API_KEY"
```

## Additional operations

Beyond the core workflow, the Batch API provides additional operations for managing your batches.

### Cancel a batch

You can cancel a batch before all requests complete. Already-processed requests remain available in the results, but pending requests will not be processed. You cannot add more requests to a cancelled batch.

```bash
curl -X POST https://api.x.ai/v1/batches/{batch_id}:cancel \\
  -H "Authorization: Bearer $XAI_API_KEY"
```

```pythonXAI
from xai_sdk import Client

client = Client()

# Cancel processing
cancelled_batch = client.batch.cancel(batch_id=batch.batch_id)
print(f"Cancelled batch: {cancelled_batch.batch_id}")
print(f"Completed before cancellation: {cancelled_batch.state.num_success} requests")
```

### List all batches

View all batches belonging to your team. Batches are retained until they expire (check the `expires_at` field). This endpoint supports the same `limit` and `pagination_token` parameters for paginating through large lists.

```bash
curl "https://api.x.ai/v1/batches?page_size=20" \\
  -H "Authorization: Bearer $XAI_API_KEY"
```

```pythonXAI
from xai_sdk import Client

client = Client()

# List recent batches
response = client.batch.list(limit=20)

for batch in response.batches:
    status = "complete" if batch.state.num_pending == 0 else "processing"
    print(f"{batch.name} ({batch.batch_id}): {status}")
```

### Check individual request status

For detailed tracking, you can inspect the metadata for each request in a batch. This shows the status, timing, and other details for individual requests. This endpoint supports the same `limit` and `pagination_token` parameters for paginating through large batches.

```bash
curl "https://api.x.ai/v1/batches/{batch_id}/requests?page_size=50" \\
  -H "Authorization: Bearer $XAI_API_KEY"
```

```pythonXAI
from xai_sdk import Client

client = Client()

# Get metadata for individual requests
metadata = client.batch.list_batch_requests(batch_id=batch.batch_id)

for request in metadata.batch_request_metadata:
    print(f"Request {request.batch_request_id}: {request.state}")
```

### Track costs

Each batch tracks the total processing cost. Access the cost breakdown after processing to understand your spending. For pricing details, see [Batch API Pricing on the Models and Pricing page](/developers/models#batch-api-pricing).

```pythonXAI
from xai_sdk import Client

client = Client()

# Get batch with cost information
batch = client.batch.get(batch_id=batch.batch_id)

# Cost is returned in ticks (1e-10 USD) for precision
total_cost_usd = batch.cost_breakdown.total_cost_usd_ticks / 1e10
print("Total cost: $%.4f" % total_cost_usd)
```

## Complete example

This end-to-end example demonstrates a realistic batch workflow: analyzing customer feedback at scale. It creates a batch, submits feedback items for sentiment analysis, waits for processing, and outputs the results. For simplicity, this example doesn't paginate results—see [Step 4](#step-4-retrieve-results) for pagination when processing larger batches.

```pythonXAI
import time
from xai_sdk import Client
from xai_sdk.chat import system, user

client = Client()

# Sample dataset: customer feedback to analyze
feedback_data = [
    {"id": "fb_001", "text": "Absolutely love this product! Best purchase ever."},
    {"id": "fb_002", "text": "Delivery was late and the packaging was damaged."},
    {"id": "fb_003", "text": "Works fine, nothing special to report."},
    {"id": "fb_004", "text": "Customer support was incredibly helpful!"},
    {"id": "fb_005", "text": "The app keeps crashing on my phone."},
]

# Step 1: Create a batch
print("Creating batch...")
batch = client.batch.create(batch_name="feedback_sentiment_analysis")
print(f"Batch created: {batch.batch_id}")

# Step 2: Build and add requests
print("\\nAdding requests...")
batch_requests = []
for item in feedback_data:
    chat = client.chat.create(
        model="grok-4.20-beta-latest-non-reasoning",
        batch_request_id=item["id"],
    )
    chat.append(system(
        "Analyze the sentiment of the customer feedback. "
        "Respond with exactly one word: positive, negative, or neutral."
    ))
    chat.append(user(item["text"]))
    batch_requests.append(chat)

client.batch.add(batch_id=batch.batch_id, batch_requests=batch_requests)
print(f"Added {len(batch_requests)} requests")

# Step 3: Wait for completion
print("\\nProcessing...")
while True:
    batch = client.batch.get(batch_id=batch.batch_id)
    pending = batch.state.num_pending
    completed = batch.state.num_success + batch.state.num_error
    
    print(f"  {completed}/{batch.state.num_requests} complete")
    
    if pending == 0:
        break
    time.sleep(2)

# Step 4: Retrieve and display results
print("\\n--- Results ---")
results = client.batch.list_batch_results(batch_id=batch.batch_id)

# Create a lookup for original feedback text
feedback_lookup = {item["id"]: item["text"] for item in feedback_data}

for result in results.succeeded:
    original_text = feedback_lookup.get(result.batch_request_id, "")
    sentiment = result.response.content.strip().lower()
    print(f"[{sentiment.upper()}] {original_text[:50]}...")

# Report any failures
if results.failed:
    print("\\n--- Errors ---")
    for result in results.failed:
        print(f"[{result.batch_request_id}] {result.error_message}")

# Display cost
cost_usd = batch.cost_breakdown.total_cost_usd_ticks / 1e10
print("\\nTotal cost: $%.4f" % cost_usd)
```

## Limitations

**Batches**

* A team can have an **unlimited** number of batches.
* Maximum batch creation rate: **1** batch creation per second per team.

**Batch Requests**

* A batch can contain an **unlimited** number of requests in theory, but extremely large batches (>1,000,000 requests) may be throttled for processing stability.
* Each individual request that can be added to a batch has a maximum payload size of **25MB**.
* A team can send up to **100** add-batch-requests API calls every **30 seconds** (this is a rolling limit shared across all batches in the team).

## Tool Use

Both [server-side tools](/developers/guides/tools/overview) and client-side function tools are supported in batch requests.

* **Server-side tools** (web search, code execution, MCP, etc.) work the same as in the real-time API — they are executed during processing and the final response is returned.
* **Client-side function tools** are supported: the model returns `tool_calls` in the response for you to handle offline. Multi-turn tool calling requires submitting a new batch request with the tool result messages included in the conversation.

## Related

* [API Reference: Batch endpoints](/developers/rest-api-reference/inference/batches#create-a-new-batch)
* [gRPC Reference: Batch management](/developers/grpc-api-reference#batch-management)
* [Models and pricing — Batch API Pricing](/developers/models#batch-api-pricing)
* [xAI Python SDK](https://github.com/xai-org/xai-sdk-python)
