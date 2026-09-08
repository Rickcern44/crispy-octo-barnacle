# Plan and approve work

Cassor records a plan before implementation begins. Inspect current work, then record the approved packet:

```sh
cassor context --json
cassor plan record --file plan.json --approve --approved-by-user
```

Approved tasks run against that exact plan revision. Start one task, verify it, and record the outcome:

```sh
cassor task start TASK_ID
cassor task complete TASK_ID --outcome "Verification evidence"
cassor check
```

Request a revised plan before changing scope, public behavior, schemas, dependencies, or acceptance criteria.
