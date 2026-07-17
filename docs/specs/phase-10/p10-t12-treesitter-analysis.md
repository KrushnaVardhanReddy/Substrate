# Phase 10 - Task 12: Tree-sitter Deterministic Impact Analysis

## 1. Goal
Parse downstream consumer repositories using Tree-sitter AST to deterministically pinpoint exactly *which lines of code* are broken by an upstream API change.

## 2. Requirements
- Integrate go-tree-sitter.
- Pull downstream code on PR webhook.
- Find precise line numbers where the deleted API field is accessed.
