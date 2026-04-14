@@STATE: running
## migrate database schema to v7

Scanning existing migrations…

▼ ls  · migrations/ · exit 0
```
0001_init.sql
0002_users.sql
...
0042_add_index.sql.bak
0043_v7_initial.sql
```

Several `.sql.bak` files from an aborted run are lingering. I'd like to remove them before proceeding so the migrator doesn't pick them up.

@@STATE: awaiting_perm
@@PERM: {"tool":"bash","args":"rm migrations/*.sql.bak","diff":null}
