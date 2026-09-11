# ADR 0001: Use a monorepo

- Status: Accepted
- Date: 2026-09-11

## Decision

Keep the services, frontend, event contracts, infrastructure, and warehouse
models in one repository.

## Why

These parts will change together while the project is young. One repository
lets a single commit update a contract and all of its producers and consumers.

## Trade-off

The repository is easier to coordinate, but CI and code ownership will need
clear boundaries as it grows. We can split components later if they gain
independent teams or release schedules.
