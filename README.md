A pre-commit hook to analyze Java files by [PMD](https://pmd.github.io/).

# Requirements

- [Go](https://go.dev/) >= v1.18
- [pre-commit](https://pre-commit.com/)

# Usage

`.pre-commit-hooks.yaml`:

```yaml
repos:
  - repo: https://github.com/maxwellhertz/pmd-java-analysis-hook
    rev: v0.1.0
    hooks:
      - id: pmd-java-analysis
        args:
          - rulesets/java/quickstart.xml
```