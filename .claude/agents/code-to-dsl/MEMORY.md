# Code-to-DSL Agent -- Operational Memory

## Learnings

### Common Mistakes in Generated Models

1. **Inverted realizedBy**: Services declaring `supports` instead of capabilities
   declaring `realizedBy`. The parser rejects this.

2. **Flat capabilities**: All capabilities at the same level with no hierarchy.
   Group related capabilities under parent capabilities with `children`.

3. **Service names as capability names**: "ingestor-service" is a service, not a
   capability. Name the business capability: "Data Ingestion & Processing".

4. **Missing visibility**: Every capability must have a visibility field.
   user-facing, domain, foundational, or infrastructure.

5. **Feature-centric needs**: "Use the ingestion API" is a feature, not a need.
   "Update catalog and see changes reflected" is an outcome-centric need.

6. **All x-as-a-service interactions**: Real team topologies have a mix of modes.
   Look for collaboration (joint design) and facilitating (coaching) patterns.

### Reference Model Quality

The `examples/inca.unm.yaml` and `examples/inca.unm.extended.yaml` files are
validated reference models. Use their structure as a template for new models.

## Validation

After generating YAML, verify it parses:
```bash
cd backend && go run ./cmd/cli/ parse /path/to/generated.unm.yaml
```

Check for zero parse errors and zero validation errors before delivery.
