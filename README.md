A pre-commit hook to analyze Java files by [PMD](https://pmd.github.io/).

# Requirements

- [Go](https://go.dev/) >= v1.18
- [pre-commit](https://pre-commit.com/)

# Usage

`.pre-commit-hooks.yaml`:

```yaml
repos:
  - repo: https://github.com/maxwellhertz/pmd-java-analysis-hook
    rev: v0.1.1
    hooks:
      - id: pmd-java-analysis
        args:
          # (Optional) The path to a ruleset xml file. The default one is rulesets/java/quickstart.xml
          - "rulesets/java/quickstart.xml"
          # (Optional) Whether to exit with status 0 even if there are violations found. The default value is false.
          - "false"
```