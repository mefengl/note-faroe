---
title: "Database backup"
---

# Database backup

Faroe does not backup its SQLite database. We recommend using [Litestream](https://litestream.io) to create database replicas locally or externally in services like S3.

Use the `replicate` command to create a replica. Faroe should be ran as a child process using the `exec` option.

```
litestream replicate -exec="./faroe serve" faroe_data/sqlite.db file://backup
```

Use the `restore` command to restore the database from a replica.

```
litestream restore -o faroe_data/sqlite.db file://backup 
rm sqlite.db.tmp-shm
sqlite.db.tmp-wal
```