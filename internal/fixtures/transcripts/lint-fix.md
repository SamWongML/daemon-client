@@STATE: running
## fix lint errors in ingestion worker

Running the linter to see the current state…

▼ pnpm lint  · 1.4s · exit 1
```
src/ingest/worker.ts:42:7 error  'ctx' is never reassigned  prefer-const
src/ingest/worker.ts:118:5 error Unexpected console.log  no-console
src/ingest/queue.ts:9:1   warning Missing return type     @typescript-eslint/explicit-function-return-type
```

Three issues total. Fixing them in one pass.

▼ edit  · src/ingest/worker.ts · exit 0
```diff
-let ctx = buildContext(job);
+const ctx = buildContext(job);
...
-console.log("enqueued", id);
+logger.debug("enqueued", { id });
```

▼ edit  · src/ingest/queue.ts · exit 0
```diff
-export function enqueue(item) {
+export function enqueue(item: QueueItem): Promise<void> {
```

Re-running the linter to confirm…

▼ pnpm lint  · 1.2s · exit 0
```
(no errors)
```

All green.

@@STATE: completed
