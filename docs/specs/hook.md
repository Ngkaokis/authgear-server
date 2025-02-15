- [Hooks](#hooks)
- [Kinds of Hooks](#kinds-of-hooks)
- [Event Delivery](#event-delivery)
- [Lifecycle of Event Delivery](#lifecycle-of-event-delivery)
- [Blocking Events](#blocking-events)
  * [Blocking Event Mutations](#blocking-event-mutations)
- [Non-blocking Events](#non-blocking-events)
  * [Future works of non-blocking events](#future-works-of-non-blocking-events)
- [Webhook](#webhook)
  * [Webhook Signature](#webhook-signature)
- [Hooks Event Management](#hooks-event-management)
  * [Hooks Event Alerts](#hooks-event-alerts)
  * [Hooks Past Events](#hooks-past-events)
  * [Hooks Manual Re-delivery](#hooks-manual-re-delivery)
- [Considerations](#considerations)
  * [Recursive Hooks](#recursive-hooks)
  * [Delivery Reliability](#delivery-reliability)
  * [Eventual Consistency](#eventual-consistency)
  * [CAP Theorem](#cap-theorem)
- [Configuration in `authgear.yaml`](#configuration-in-authgearyaml)

# Hooks

Hooks is a mechanism to notify external services about [events](./event.md).

# Kinds of Hooks

There are 2 kinds of Hooks.

- [Webhook](#webhook)
- [Deno Hook](#deno-hook)

# Event Delivery

Each event can have many Hooks. The order of delivery is unspecified for [non-blocking event](#non-blocking-events). [Blocking events](#blocking-events) are delivered in the source order as in the configuration.

Hooks should be idempotent, since non-blocking events may be delivered multiple times due to retries.

# Lifecycle of Event Delivery

1. Begin transaction
1. Perform operation
1. Deliver blocking events to Hooks.
1. If failed, rollback the transaction.
1. Perform mutations
1. Commit transaction
1. Deliver non-blocking events to Hooks.

# Blocking Events

Blocking events are delivered to hooks synchronously, right before committing changes to the database.

Hooks must return a JSON document to indicate whether the operation should continue.

To let the operation to proceed, respond with `is_allowed` set to `true`.

```json
{
  "is_allowed": true
}
```

To fail the operation, respond with `is_allowed` set to `false`, and a non-empty `title` and `reason`.

```json
{
  "is_allowed": false,
  "title": "any title",
  "reason": "any string"
}
```

If any hook fails the operation, the operation is failed. The operation fails with error

```json
{
  "error": {
    "name": "Forbidden",
    "reason": "WebHookDisallowed",
    "info": {
      "reasons": [
        {
          "title": "any title",
          "reason": "any string"
        }
      ]
    }
  }
}
```

> For backward compatibility, the reason is called "WebHookDisallowed".

The time spent in a blocking event delivery must not exceed 5 seconds, otherwise it will be considered as a failed delivery. Also, the total time spent in all deliveries of the event must not exceed 10 seconds, otherwise it would also be considered as failed delivery. Both timeouts are configurable.

Blocking events are not persisted, and their failed deliveries are not retried.

## Blocking Event Mutations

Hooks can optionally mutate the object in the Event payload.

Hooks cannot request mutation if the operation is failed by them.

Hooks specify the mutations in the JSON document they return.

Given the event

```json
{
  "payload": {
    "user": {
      "standard_attributes": {
        "name": "John"
      }
    }
  }
}
```

Hooks can mutate the standard attributes of the user with the following JSON document.

```json
{
  "is_allowed": true,
  "mutations": {
    "user": {
      "standard_attributes": {
        "name": "Jane"
      }
    }
  }
}
```

Objects not appearing in `mutations` are left intact.

The mutated objects do NOT merge with the original ones.

The mutated payload are NOT validated and are propagated along the Hooks chain.
The payload will only be validated after traversing the Hooks chain.

Mutations do NOT generate extra events to avoid infinite loop.

Currently, only `standard_attributes`, `custom_attributes` and `roles` of the user object are mutable.

# Non-blocking Events

Non-blocking events are delivered to Hooks asynchronously after the operation is performed (i.e. changes committed to the database).

The time spent in an non-blocking event delivery must no exceed 60 seconds, otherwise it would be considered as a failed delivery.

The return value of non-blocking event Hooks is ignored.

## Future works of non-blocking events

All non-blocking events with registered Hooks are persisted in the database, with minimum retention period of 30 days.

If any delivery failed, all deliveries will be retried after some time, regardless of whether some deliveries may have succeeded. The retry is performed with a variant of exponential back-off algorithm. Specifically for Webhooks, if `Retry-After:` HTTP header is present in the response, the delivery will not be retried before the specific time.

If the delivery keeps on failing after 3 days from the time of first attempted delivery, the event will be marked as permanently failed and will not be retried automatically.

# Webhook

Webhook is a kind of Hook via the HTTPS protocol. This ensures integrity and confidentiality of the delivery.

Events are POSTed to the Webhook.

The endpoint of the Webhook must be an absolute URL.

The Webhook must return a status code within the 2xx range. Other status code is considered as a failed delivery.

## Webhook Signature

Each request is signed with a secret key shared between Authgear and the Webhook. The developer must validate the signature and reject requests with invalid signature to ensure the request originates from Authgear.

The signature is calculated as the hex encoded value of HMAC-SHA256 of the request body.

The signature is included in the header `x-authgear-body-signature:`.

> For advanced end-to-end security scenario, some network admin may wish to
> use mTLS for authentication. It is not supported at the moment.

# Deno Hook

Deno Hook is a kind of Hook in form of a TypeScript / JavaScript module. The module is executed by [Deno](https://deno.land/).

The module **MUST** have a [default export](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/export#description) of a function taking 1 argument. The argument is the event payload. The function can either be synchronous or asynchronous. An asynchronous function is a function returning [Promise](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise), or an [async function](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/async_function)

If the Deno Hook is registered for a [blocking event](#blocking-events), the function **MUST** return a value according to the [specification](#blocking-events).

If the Deno hook is registered for a [non-blocking event](#non-blocking-events), the return value is ignored.

Program run with Deno has [no access](https://deno.land/manual@v1.27.2/basics/permissions) to file, network or environment by default. In case of Deno Hook, it only has access to external network. Other access is blocked. For example, A Deno Hook is **NOT** allowed to read or write the file system.

The stdout and the stderr of the Deno Hook is ignored currently.
The arguments and the stdin is intentionally unspecified. A Deno Hook **MUST NOT** assume anything on them.

Deno Hooks are stored along with other app resources, such as `authgear.yaml` and templates.
The size limit of a Deno Hook is 100KiB. A module larger than 100KiB **CANNOT** be registered as a Deno Hook.

## Example of a blocking Deno Hook

```typescript
import { HookEvent, HookResponse } from "https://deno.land/x/authgear-deno-hook@0.1.0/mod.ts"

export default async function(e: HookEvent): Promise<HookResponse> {
  return { is_allowed: true };
}
```

## Example of a non-blocking Deno Hook

```typescript
import { HookEvent } from "https://deno.land/x/authgear-deno-hook@0.1.0/mod.ts"

export default async function(e: HookEvent): Promise<void> {
  // Do something with e.
}
```

# Hooks Event Management

## Hooks Event Alerts

If an event delivery is permanently failed, an ERROR log is generated to notify developers.

## Hooks Past Events

An API is provided to list past events. This can be used to reconcile self-managed database with the failed events.

> NOTE: Blocking events are not persisted, regardless of success or failure.

## Hooks Manual Re-delivery

The developer can manually trigger a re-delivery of failed event, bypassing the retry interval limit.

> NOTE: Blocking events cannot be re-delivered.

# Considerations

## Recursive Hooks

An ill-designed Hook may be triggered recursively. For example, calling API that will trigger other events.

The developer is responsible for ensuring that:
- Hooks would not be triggered recursively; or
- Recursive Hooks have well-defined termination condition.

## Delivery Reliability

The main purpose of Hooks is to allow external services to observe state changes.

Therefore, AFTER events are persistent, immutable, and delivered reliably. Otherwise, external services may observe inconsistent changes.

It is not recommended to perform side effects in blocking event Hooks. Otherwise, the developer should consider how to compensate for the side effects of potential failed operation.

## Eventual Consistency

Fundamentally, Hooks is a distributed system. When Hooks have side effects, we need to choose between guaranteeing consistency or availability of the system (See [CAP Theorem](#cap-theorem)).

We decided to ensure the availability of the system. To maintain consistency, the developer should take eventual consistency into account when designing their system.

The developer should regularly check the past events for unprocessed events to ensure consistency.

## CAP Theorem

To simplify, the CAP theorem states that a distributed data store can satisfy
only two of the three properties simultaneously:
- Consistency
- Availability
- Network Partition Tolerance

Since network partition cannot be avoided practically, distributed system would
need to choose between consistency and availability. Most microservice
architecture prefer availability over strong consistency, and instead application
state is eventually consistent.

# Configuration in `authgear.yaml`

```
hook:
  blocking_handlers:
    - event: "user.pre_create"
      url: "https://myapp.com/check_user_create"
    - event: "user.pre_create"
      url: "authgeardeno:///deno/randomstring.ts"
  non_blocking_handlers:
    - events: ["*"]
      url: 'https://myapp.com/all_events'
    - events: ["*"]
      url: "authgeardeno:///deno/randomstring.ts"
    - events: ["user.created"]
      url: 'https://myapp.com/sync_user_creation'
```
