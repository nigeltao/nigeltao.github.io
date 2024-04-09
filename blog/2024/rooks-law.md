# Rook's Law - There's Always a Limit

Here's some software engineering wisdom from my colleague [Nate
Rook](https://github.com/n-rook/).

> Always set a limit to the size of the entities your product consumes.
>
> If you define a limit, and a user hits it, they are still in a reasonable spot. They will receive an error message clearly stating what has gone wrong. Maybe they can quickly iterate and fix the problem, but even if they can't, they can at least beg you to raise the limit; you can temporarily let them exceed it, or permanently raise the limit if you decide that's a better idea.
>
> But if you don't define a limit, there's still going to be one. *There's always a limit.* And the emergent limit is likely to be a lot less nice than a user-defined one. It will return confusing error messages, or no error at all; it may vary over time depending on ambient conditions; and it certainly can't be raised on short notice.


---

Published: 2024-04-10
