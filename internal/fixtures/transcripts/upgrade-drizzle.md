@@STATE: running
## upgrade drizzle to 0.33

Reading `package.json` and `drizzle.config.ts`…

Current version is `drizzle-orm@0.29.4`. Bumping to `0.33.0` and updating the matching `drizzle-kit`.

▼ pnpm up drizzle-orm@0.33.0 drizzle-kit@0.24.0  · 8.2s · exit 0

Types changed — `InferSelectModel` is now in `drizzle-orm/table` instead of the root. Patching imports…

▼ edit  · src/db/schema.ts · exit 0
```diff
-import { InferSelectModel } from "drizzle-orm";
+import { InferSelectModel } from "drizzle-orm/table";
```

Running the migration generator against the staging database…

▼ pnpm drizzle-kit generate  · 4.1s · exit 1
```
Error: Failed to introspect postgres schema
  at Introspector.run (node_modules/drizzle-kit/dist/index.js:8412)
  at async generate (node_modules/drizzle-kit/dist/index.js:9234)
Caused by: relation "public._drizzle_migrations" does not exist
```

The new drizzle-kit expects a migrations tracking table that the 0.29 series never created. Rolling this back is non-trivial and I don't have permission to run DDL against staging from here.

@@STATE: failed
