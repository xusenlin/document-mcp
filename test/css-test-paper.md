# Document Title

Author Name

2026-05-18

## Introduction

This is a paragraph with **bold text**, *italic text*, and `inline code`. Here is a [link](https://example.com) for testing.

> This is a blockquote with multiple lines.
> It should render with a left border and italic style.

## Code Examples

### Go Code

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) > 1 {
        fmt.Printf("Hello, %s!\n", os.Args[1])
        return
    }
    fmt.Println("Hello, World!")
}
```

### Shell Command

```bash
#!/bin/bash
set -euo pipefail

echo "Build started..."
go build -o bin/app ./cmd/server/
echo "Build complete."
```

### Inline Elements

Press `Ctrl+C` to exit. Use `--flag=value` syntax.

## Tables

### Simple Table

| Name | Age | City |
|------|-----|------|
| Alice | 28 | Beijing |
| Bob | 35 | Shanghai |
| Charlie | 42 | Shenzhen |

### Wide Table

| Feature | Description | Status | Priority |
|---------|-------------|--------|----------|
| Auth | OAuth2 integration | Done | High |
| Export | PDF generation | In Progress | Medium |
| Search | Full-text search | Planned | Low |
| Backup | Automatic backup | Done | High |

## Lists

### Unordered

- First level item
  - Second level item
    - Third level item
- Another first level
- Yet another

### Ordered

1. Step one
2. Step two
   1. Sub-step A
   2. Sub-step B
3. Step three

### Task List

- [x] Completed task
- [ ] Pending task
- [ ] Another pending

## Nested Content

### Blockquote with Code

> Here is a code block inside a quote:
>
> ```python
> def hello():
>     print("Hello from blockquote")
> ```

### List with Code

1. First item with code:

   ```json
   {
     "key": "value",
     "nested": {
       "deep": true
     }
   }
   ```

2. Second item

## Definition List

Term 1
: Definition of term one.

Term 2
: Definition of term two with more text.

## Horizontal Rule

Above the rule.

---

Below the rule.

## Footnotes

Here is a sentence with a footnote.[^1]

[^1]: This is the footnote content.

## Images

![Alt text](https://via.placeholder.com/400x200.png?text=Test+Image)

## Long Text

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.

## Mathematical Formulas

Inline formula: $E = mc^2$.

Block formula:

$$
\int_{0}^{\infty} e^{-x^2} dx = \frac{\sqrt{\pi}}{2}
$$

## Multiple Tables

| Left | Center | Right |
|:-----|:------:|------:|
| L1 | C1 | R1 |
| L2 | C2 | R2 |

## Code with Highlighting

```javascript
const greeting = "Hello, World!";
function sayHello(name) {
    console.log(`${greeting} My name is ${name}`);
}
sayHello("Test");
```
