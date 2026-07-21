## Plan: Echo To Gin Pre-Migration (Steps 9, 8, 5, 6)

This is an updated plan in the exact order you requested: 9 -> 8 -> 5 -> 6.

**Steps**
1. Expand migration-safety integration coverage (Step 9)
2. Introduce an internal HTTP error model decoupled from Echo (Step 8) depends on 1
3. Move Recover/ErrorConverter to the global middleware chain (Step 5) depends on 2
4. Strengthen response-envelope header guarantees (Step 6) depends on 3

**Phase 1: Migration-Safety Test Net (Step 9)**
1. Add integration assertions that non-raw routes always return a UUID request_id in both body and header, and that body/header request IDs match.
2. Add integration assertions that raw routes still return X-Request-Id while bypassing envelope wrapping.
3. Add panic and returned-error path checks asserting stable envelope shape: request_id, status, status_code, details format.
4. Add checks for envelope invariants under representative handlers: content-length consistency and status mapping correctness.
5. Keep tests behavior-oriented so they can be reused during the Gin swap with minimal rewrites.

**Phase 2: Error Model Decoupling (Step 8)**
1. Introduce an internal HTTP error type in middleware (status code, message, optional details/cause).
2. Replace direct Echo-specific error coupling in conversion logic with the internal type.
3. Add an Echo adapter at the framework boundary for translating internal errors.
4. Update tests to assert internal error semantics first, then adapter behavior.
5. Preserve external behavior (status/message) to avoid regressions.

**Phase 3: Middleware Placement Simplification (Step 5)**
1. Move Recover and ErrorConverter from per-route stack to base/global middleware registration.
2. Keep RequestId and route-controlled ResponseEnvelope behavior where toggling is needed.
3. Verify global order so panic/error transformation runs consistently before final response write.
4. Update unit/integration tests to reflect new composition source without changing expected outcomes.

**Phase 4: Response Header Hardening (Step 6)**
1. Enforce deterministic content-type for enveloped responses.
2. Ensure X-Request-Id is always present on enveloped responses, including fallback paths.
3. Keep Content-Length synchronized with wrapped payload size after envelope serialization.
4. Add tests for success/error/fallback-ID scenarios validating header invariants and body/header consistency.
5. Add an explicit check that raw routes keep non-envelope response behavior while still carrying X-Request-Id.

**Relevant files**
- pkg/server/middlewares_test.go - extend integration invariants.
- pkg/server/server_test.go - reinforce end-to-end envelope/error invariants.
- pkg/middleware/error_converter.go - move to internal HTTP error model.
- pkg/middleware/utils.go - remove Echo-coupled wrapping internals.
- pkg/middleware/recover.go - emit internal error type on panic path.
- pkg/server/server.go - register global Recover/ErrorConverter ordering.
- pkg/server/middlewares.go - simplify route-level middleware composition.
- pkg/middleware/response_envelope.go - enforce header guarantees.
- pkg/rest/response_envelope_writer.go - lock content-type/content-length behavior.
- pkg/middleware/response_envelope_test.go - add hardening assertions.

**Verification**
1. go test ./pkg/middleware -run ResponseEnvelope
2. go test ./pkg/middleware -run RequestId
3. go test ./pkg/middleware -run RequestTracer
4. go test ./pkg/server -run BuildMiddlewaresForRoute
5. go test ./pkg/server -run Server
6. go test ./pkg/server ./pkg/middleware ./pkg/rest
