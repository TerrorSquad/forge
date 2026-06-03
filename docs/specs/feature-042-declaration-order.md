# Feature 042: Declaration-Order Tool Execution

Preserve tool execution order according to the declaration order in `forge.toml`.

## Summary

Tool ordering is meaningful in hooks. `forge` should execute tools in the same order
that the user declares them in the config, without requiring an explicit `order` field.

## Acceptance criteria

- Tool sections under `[hooks.<hook>.tools.<name>]` are executed in declaration order.
- Tools not declared in the ordered section appear after declared tools in alphabetical order.
- Documentation clearly explains that declaration order drives execution order.

## Example

```toml
[hooks.pre-commit.tools.gofmt]
command = "gofmt"

[hooks.pre-commit.tools.eslint]
command = "eslint"

[hooks.pre-commit.tools.phpstan]
command = "phpstan"
```

The tools should run in the order: `gofmt`, `eslint`, `phpstan`.
